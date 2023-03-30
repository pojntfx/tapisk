package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/tapisk/pkg/backend"
	"github.com/pojntfx/tapisk/pkg/utils"
)

func main() {
	file := flag.String("file", "/dev/nst6", "Path to device file to connect to")
	size := flag.Int64("size", 500*1024*1024, "Size of the tape to expose (native size, not compressed size)")
	laddr := flag.String("laddr", ":10809", "Listen address")
	network := flag.String("network", "tcp", "Listen network (e.g. `tcp` or `unix`)")
	name := flag.String("name", "default", "Export name")
	description := flag.String("description", "The default export", "Export description")
	readOnly := flag.Bool("read-only", false, "Whether the export should be read-only")

	flag.Parse()

	l, err := net.Listen(*network, *laddr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	log.Println("Listening on", l.Addr())

	var f *os.File
	if *readOnly {
		f, err = os.OpenFile(*file, os.O_RDONLY, os.ModeCharDevice)
		if err != nil {
			panic(err)
		}
	} else {
		f, err = os.OpenFile(*file, os.O_RDWR, os.ModeCharDevice)
		if err != nil {
			panic(err)
		}
	}
	defer f.Close()

	blocksize, err := utils.GetBlocksize(f)
	if err != nil {
		panic(err)
	}

	b := backend.NewTapeBackend(f, *size, blocksize)

	clients := 0
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Could not accept connection, continuing:", err)

			continue
		}

		clients++

		log.Printf("%v clients connected", clients)

		go func() {
			defer func() {
				_ = conn.Close()

				clients--

				if err := recover(); err != nil {
					log.Printf("Client disconnected with error: %v", err)
				}

				log.Printf("%v clients connected", clients)
			}()

			if err := server.Handle(
				conn,
				[]server.Export{
					{
						Name:        *name,
						Description: *description,
						Backend:     b,
					},
				},
				&server.Options{
					ReadOnly:           *readOnly,
					MinimumBlockSize:   uint32(blocksize),
					PreferredBlockSize: uint32(blocksize),
					MaximumBlockSize:   uint32(blocksize),
				}); err != nil {
				panic(err)
			}
		}()
	}
}
