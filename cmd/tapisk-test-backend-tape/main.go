package main

import (
	"flag"
	"log"
	"os"

	"github.com/pojntfx/tapisk/pkg/backend"
	"github.com/pojntfx/tapisk/pkg/utils"
)

func main() {
	file := flag.String("file", "/dev/nst6", "Path to device file to connect to")
	size := flag.Int64("size", 5*1024*1024, "Size of the tape to expose (native size, not compressed size)")
	compat := flag.Bool("compat", false, "Whether to not emulate all SCSI commands using manual seeks")

	flag.Parse()

	f, err := os.OpenFile(*file, os.O_RDWR, os.ModeCharDevice)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	blocksize, err := utils.GetBlocksize(f)
	if err != nil {
		panic(err)
	}

	// Initialize tape
	p := make([]byte, blocksize)
	for i := uint64(0); i < (uint64(*size) / blocksize); i++ {
		log.Println(i, (uint64(*size) / blocksize))

		if _, err := f.Write(p); err != nil {
			panic(err)
		}
	}

	b := backend.NewTapeBackend(f, *size, blocksize, *compat)

	{
		// Write and read
		input := make([]byte, blocksize)
		copy(input, []byte("First message body"))
		if _, err := b.WriteAt(input, 0); err != nil {
			panic(err)
		}

		output := make([]byte, blocksize)
		if _, err := b.ReadAt(output, 0); err != nil {
			panic(err)
		}

		log.Println("First message:", string(input), string(output), string(input) == string(output))
	}

	{
		// Overwrite and read
		input := make([]byte, blocksize)
		copy(input, []byte("Overwrite message body"))
		if _, err := b.WriteAt(input, 0); err != nil {
			panic(err)
		}

		output := make([]byte, blocksize)
		if _, err := b.ReadAt(output, 0); err != nil {
			panic(err)
		}

		log.Println("Overwrite:", string(input), string(output), string(input) == string(output))
	}

	{
		// Write and read part of a block
		input := []byte("Part of a block")

		inputBlock := make([]byte, blocksize)
		copy(inputBlock, input)
		if _, err := b.WriteAt(inputBlock, 0); err != nil {
			panic(err)
		}

		output := make([]byte, len(input)-3)
		if _, err := b.ReadAt(output, 3); err != nil {
			panic(err)
		}

		log.Println("Part of a block:", string(inputBlock), string(output), string(inputBlock) == string(output))
	}

	{
		// Write and read part of a second block
		input := []byte("Part of a second block")

		if _, err := b.WriteAt(make([]byte, blocksize), 0); err != nil {
			panic(err)
		}

		inputBlock := make([]byte, blocksize)
		copy(inputBlock, input)
		if _, err := b.WriteAt(inputBlock, int64(blocksize)); err != nil {
			panic(err)
		}

		output := make([]byte, len(input)-3)
		if _, err := b.ReadAt(output, int64(blocksize)+3); err != nil {
			panic(err)
		}

		log.Println("Part of a block:", string(inputBlock), string(output), string(inputBlock) == string(output))
	}
}
