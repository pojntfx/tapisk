package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/tapisk/pkg/backend"
)

func main() {
	file := flag.String("file", "/dev/nst0", "Path to device file to connect to")
	size := flag.Int64("size", 2.5*1024*1024*1024*1024*1024, "Size of the tape to expose (native size, not compressed size)")
	laddr := flag.String("laddr", ":10809", "Listen address")
	network := flag.String("network", "tcp", "Listen network (e.g. `tcp` or `unix`)")
	name := flag.String("name", "default", "Export name")
	description := flag.String("description", "The default export", "Export description")
	readOnly := flag.Bool("read-only", false, "Whether the export should be read-only")
	minimumBlockSize := flag.Uint("minimum-block-size", 1, "Minimum block size")
	preferredBlockSize := flag.Uint("preferred-block-size", 4096, "Preferred block size")
	maximumBlockSize := flag.Uint("maximum-block-size", 0xffffffff, "Maximum block size")

	flag.Parse()

	l, err := net.Listen(*network, *laddr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	log.Println("Listening on", l.Addr())

	var f *os.File
	if *readOnly {
		f, err = os.OpenFile(*file, os.O_RDONLY, 0644)
		if err != nil {
			panic(err)
		}
	} else {
		f, err = os.OpenFile(*file, os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
	}
	defer f.Close()

	b := backend.NewTapeBackend(f, *size)

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
					MinimumBlockSize:   uint32(*minimumBlockSize),
					PreferredBlockSize: uint32(*preferredBlockSize),
					MaximumBlockSize:   uint32(*maximumBlockSize),
				}); err != nil {
				panic(err)
			}
		}()
	}
}
