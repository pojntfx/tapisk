package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	gbackend "github.com/pojntfx/go-nbd/pkg/backend"
	"github.com/pojntfx/go-nbd/pkg/server"
	rbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/tapisk/pkg/backend"
	idx "github.com/pojntfx/tapisk/pkg/index"
	"github.com/pojntfx/tapisk/pkg/mtio"
)

func main() {
	drivePath := flag.String("drive", "/dev/nst0", "Path to tape drive")
	size := flag.Int64("size", 500*1024*1024, "Size of the tape (native size, not compressed size)")

	indexPath := flag.String("index", "tapisk.db", "Path to index file")
	bucket := flag.String("bucket", "tapsik", "Bucket in index file")

	cachePath := flag.String("cache", "tapisk.img", "Path to cache file")
	pullWorkers := flag.Int64("pull-workers", 1, "Amount of background pull workers (negative values disable background pull)")
	pushWorkers := flag.Int64("push-workers", 1, "Amount of background push workers (negative values disable background push)")
	pushInterval := flag.Duration("push-interval", 5*time.Minute, "Interval for pushing changes from the cache to the tape")

	readOnly := flag.Bool("read-only", false, "Whether the block device should be read-only")

	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	cacheFile, err := os.OpenFile(*cachePath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer cacheFile.Close()

	if err := cacheFile.Truncate(*size); err != nil {
		panic(err)
	}

	index := idx.NewBboltIndex(*indexPath, *bucket)

	if err := index.Open(); err != nil {
		panic(err)
	}
	defer index.Close()

	chunkSize, err := mtio.GetBlocksize(driveFile)
	if err != nil {
		panic(err)
	}

	b := backend.NewTapeBackend(driveFile, index, *size, chunkSize, *verbose)

	mnt := mount.NewManagedFileMount(
		ctx,

		rbackend.NewReaderAtBackend(
			chunks.NewArbitraryReadWriterAt(
				chunks.NewChunkedReadWriterAt(
					b,
					int64(chunkSize),
					*size/(int64(chunkSize)),
				),
				int64(chunkSize),
			),
			b.Size,
			b.Sync,
			false,
		),
		gbackend.NewFileBackend(cacheFile),

		&mount.ManagedMountOptions{
			ChunkSize: int64(chunkSize),

			PullWorkers: *pullWorkers,

			PushWorkers:  *pushWorkers,
			PushInterval: *pushInterval,

			Verbose: *verbose,
		},
		nil,

		&server.Options{
			ReadOnly: *readOnly,

			MinimumBlockSize:   uint32(chunkSize),
			PreferredBlockSize: uint32(chunkSize),
			MaximumBlockSize:   uint32(chunkSize),
		},
		nil,
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := mnt.Wait(); err != nil {
			panic(err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	go func() {
		<-done

		log.Println("Exiting gracefully")

		_ = mnt.Close()
	}()

	defer mnt.Close()
	path, err := mnt.Open()
	if err != nil {
		panic(err)
	}

	fmt.Println(path)

	wg.Wait()
}
