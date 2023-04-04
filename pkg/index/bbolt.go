package index

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	"go.etcd.io/bbolt"
)

var (
	ErrNotExists = errors.New("location does not exist")
)

type BboltIndex struct {
	file   string
	bucket string
	db     *bbolt.DB
}

func NewBboltIndex(file, bucket string) *BboltIndex {
	return &BboltIndex{file, bucket, nil}
}

func (b *BboltIndex) Open() (err error) {
	b.db, err = bbolt.Open(b.file, os.ModePerm, nil)
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *bbolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists([]byte(b.bucket))

		return
	})
}

func (b *BboltIndex) Close() error {
	return b.db.Close()
}

func (b *BboltIndex) SetLocation(block uint64, location uint64) error {
	key := make([]byte, 8) // uint64
	binary.BigEndian.PutUint64(key, block)

	value := make([]byte, 8) // uint64
	binary.BigEndian.PutUint64(value, location)

	return b.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(b.bucket)).Put(key, value)
	})
}

func (b *BboltIndex) GetLocation(block uint64) (location uint64, err error) {
	key := make([]byte, 8) // uint64
	binary.BigEndian.PutUint64(key, block)

	return location, b.db.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket([]byte(b.bucket)).Get(key)

		if value == nil {
			return ErrNotExists
		}

		return binary.Read(bytes.NewReader(value), binary.BigEndian, &location)
	})
}
