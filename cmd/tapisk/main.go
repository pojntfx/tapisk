package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/pojntfx/go-nbd/pkg/server"
	lbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/utils"
	"github.com/pojntfx/tapisk/pkg/backend"
	idx "github.com/pojntfx/tapisk/pkg/index"
	"github.com/pojntfx/tapisk/pkg/mtio"
)

func main() {
	drivePath := flag.String("drive", "/dev/nst0", "Path to tape drive")
	size := flag.Int64("size", 500*1024*1024, "Size of the tape (native size, not compressed size)")

	index := flag.String("index", "tapisk.db", "Path to index file")
	bucket := flag.String("bucket", "tapsik", "Bucket in index file")

	readOnly := flag.Bool("read-only", false, "Whether the block device should be read-only")

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

	chunkSize, err := mtio.GetBlocksize(driveFile)
	if err != nil {
		panic(err)
	}

	i := idx.NewBboltIndex(*index, *bucket)

	if err := i.Open(); err != nil {
		panic(err)
	}
	defer i.Close()

	rawBackend := backend.NewTapeBackend(driveFile, i, *size, chunkSize, *verbose)
	b := lbackend.NewReaderAtBackend(
		chunks.NewChunkedReadWriterAt(
			rawBackend,
			int64(chunkSize),
			*size/(int64(chunkSize)),
		),
		rawBackend.Size,
		rawBackend.Sync,
		false,
	)

	devPath, err := utils.FindUnusedNBDDevice()
	if err != nil {
		panic(err)
	}

	devFile, err := os.Open(devPath)
	if err != nil {
		panic(err)
	}
	defer devFile.Close()

	dev := mount.NewDirectPathMount(
		b,
		devFile,

		&server.Options{
			ReadOnly: *readOnly,

			MinimumBlockSize:   uint32(chunkSize),
			PreferredBlockSize: uint32(chunkSize),
			MaximumBlockSize:   uint32(chunkSize),
		},
		nil,
	)

	var (
		errs = make(chan error)
		wg   sync.WaitGroup
	)
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for err := range errs {
			if err != nil {
				panic(err)
			}
		}
	}()

	go func() {
		if err := dev.Wait(); err != nil {
			errs <- err

			return
		}

		close(errs)
	}()

	defer dev.Close()
	if err := dev.Open(); err != nil {
		panic(err)
	}

	fmt.Println(devPath)
}
