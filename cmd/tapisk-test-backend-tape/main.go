package main

import (
	"flag"
	"log"
	"os"

	"github.com/pojntfx/tapisk/pkg/backend"
	"github.com/pojntfx/tapisk/pkg/utils"
)

func main() {
	file := flag.String("file", "/dev/nst4", "Path to device file to connect to")
	size := flag.Int64("size", 2.5*1024*1024*1024*1024*1024, "Size of the tape to expose (native size, not compressed size)")

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

	b := backend.NewTapeBackend(f, *size, blocksize)

	{
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
		input := make([]byte, blocksize)
		copy(input, []byte("Second message body"))
		if _, err := b.WriteAt(input, 0); err != nil {
			panic(err)
		}

		output := make([]byte, blocksize)
		if _, err := b.ReadAt(output, 0); err != nil {
			panic(err)
		}

		log.Println("Second message:", string(input), string(output), string(input) == string(output))
	}
}
