package main

import (
	"flag"
	"log"
	"os"

	"github.com/pojntfx/tapisk/pkg/backend"
)

func main() {
	file := flag.String("file", "/dev/nst4", "Path to device file to connect to")
	size := flag.Int64("size", 2.5*1024*1024*1024*1024*1024, "Size of the tape to expose (native size, not compressed size)")
	blocksize := flag.Int("blocksize", 512, "Block size for the tape to expose")

	flag.Parse()

	f, err := os.OpenFile(*file, os.O_RDWR, os.ModeCharDevice)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b := backend.NewTapeBackend(f, *size)

	{
		input := make([]byte, *blocksize)
		copy(input, []byte("Hello, world!"))
		if _, err := b.WriteAt(input, 0); err != nil {
			panic(err)
		}

		output := make([]byte, *blocksize)
		if _, err := b.ReadAt(output, 0); err != nil {
			panic(err)
		}

		log.Println(string(output))
	}
}
