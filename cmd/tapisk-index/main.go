package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"log"
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"
)

func main() {
	file := flag.String("file", filepath.Join(os.TempDir(), "tapisk.db"), "Path to index DB")
	bucket := flag.String("bucket", "tapisk", "Bucket to store the index in")

	flag.Parse()

	db, err := bbolt.Open(*file, os.ModePerm, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	key := make([]byte, 8) // uint64
	binary.BigEndian.PutUint64(key, uint64(5))

	if err := db.Update(func(tx *bbolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists([]byte(*bucket))

		return
	}); err != nil {
		panic(err)
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(*bucket))

		value := make([]byte, 8) // uint64
		binary.BigEndian.PutUint64(value, uint64(80))

		if err := bucket.Put(key, value); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}

	if err := db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(*bucket))

		item := bucket.Get(key)

		var value uint64
		if err := binary.Read(bytes.NewReader(item), binary.BigEndian, &value); err != nil {
			return err
		}

		log.Println(value)

		return nil
	}); err != nil {
		panic(err)
	}
}
