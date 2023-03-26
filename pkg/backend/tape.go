package backend

import (
	"errors"
	"os"
	"sync"
)

var (
	errNotImplemented = errors.New("not implemented")
)

type TapeBackend struct {
	device *os.File
	size   int64
	lock   sync.Mutex
}

func NewTapeBackend(device *os.File, size int64) *TapeBackend {
	return &TapeBackend{device, size, sync.Mutex{}}
}

func (b *TapeBackend) ReadAt(p []byte, off int64) (n int, err error) {
	return -1, errNotImplemented
}

func (b *TapeBackend) WriteAt(p []byte, off int64) (n int, err error) {
	return -1, errNotImplemented
}

func (b *TapeBackend) Size() (int64, error) {
	return b.size, nil
}

func (b *TapeBackend) Sync() error {
	return errNotImplemented
}
