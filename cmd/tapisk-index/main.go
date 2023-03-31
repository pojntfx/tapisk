package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
)

func main() {
	file := flag.String("file", filepath.Join(os.TempDir(), "tapisk.db"), "Path to index DB")

	flag.Parse()

	opts := badger.DefaultOptions(*file)
	opts = opts.WithLoggingLevel(badger.ERROR)

	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	key := make([]byte, 8) // uint64
	binary.BigEndian.PutUint64(key, uint64(5))

	{
		tx := db.NewTransaction(true)

		value := make([]byte, 8) // uint64
		binary.BigEndian.PutUint64(value, uint64(80))

		if err := tx.Set(key, value); err != nil {
			tx.Discard()

			panic(err)
		}

		if err := tx.Commit(); err != nil {
			panic(err)
		}
	}

	{
		tx := db.NewTransaction(false)

		item, err := tx.Get(key)
		if err != nil {
			tx.Discard()

			panic(err)
		}

		if err := item.Value(func(val []byte) error {
			var value uint64
			if err := binary.Read(bytes.NewReader(val), binary.BigEndian, &value); err != nil {
				return err
			}

			log.Println(value)

			return nil
		}); err != nil {
			tx.Discard()

			panic(err)
		}

		tx.Discard()
	}
}
