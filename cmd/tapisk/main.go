package main

import (
	"flag"
	"log"
	"os"

	"github.com/pojntfx/go-nbd/pkg/server"
	bbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/device"
	"github.com/pojntfx/tapisk/pkg/backend"
	idx "github.com/pojntfx/tapisk/pkg/index"
	"github.com/pojntfx/tapisk/pkg/mtio"
)

func main() {
	drivePath := flag.String("drive", "/dev/nst0", "Path to tape drive file to connect to")
	index := flag.String("index", "tapisk.db", "Path to index file to use")
	bucket := flag.String("bucket", "tapsik", "Bucket in index file to use")
	size := flag.Int64("size", 500*1024*1024, "Size of the tape (native size, not compressed size)")
	devPath := flag.String("device", "/dev/nbd0", "Path to NBD device file to use")
	readOnly := flag.Bool("read-only", false, "Whether the device should be read-only")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	var (
		driveFile *os.File
		err       error
	)
	if *readOnly {
		driveFile, err = os.OpenFile(*drivePath, os.O_RDONLY, os.ModeCharDevice)
		if err != nil {
			panic(err)
		}
	} else {
		driveFile, err = os.OpenFile(*drivePath, os.O_RDWR, os.ModeCharDevice)
		if err != nil {
			panic(err)
		}
	}
	defer driveFile.Close()

	blocksize, err := mtio.GetBlocksize(driveFile)
	if err != nil {
		panic(err)
	}

	i := idx.NewBboltIndex(*index, *bucket)

	if err := i.Open(); err != nil {
		panic(err)
	}
	defer i.Close()

	rawBackend := backend.NewTapeBackend(driveFile, i, *size, blocksize, *verbose)

	chunkedRwat := chunks.NewArbitraryReadWriterAt(
		chunks.NewChunkedReadWriterAt(
			rawBackend,
			int64(blocksize),
			*size/(int64(blocksize)),
		),
		int64(blocksize),
	)

	b := bbackend.NewReaderAtBackend(chunkedRwat, rawBackend.Size, rawBackend.Sync, false)

	devFile, err := os.Open(*devPath)
	if err != nil {
		panic(err)
	}
	defer devFile.Close()

	dev := device.NewDevice(
		b,
		devFile,

		&server.Options{
			ReadOnly:           *readOnly,
			MinimumBlockSize:   uint32(blocksize),
			PreferredBlockSize: uint32(blocksize),
			MaximumBlockSize:   uint32(blocksize),
		},
		nil,
	)

	errs := make(chan error)

	go func() {
		if err := dev.Wait(); err != nil {
			errs <- err

			return
		}
	}()

	defer dev.Close()
	if err := dev.Open(); err != nil {
		panic(err)
	}

	log.Println("Ready on", *devPath)

	for range errs {
		if err != nil {
			panic(err)
		}
	}
}
