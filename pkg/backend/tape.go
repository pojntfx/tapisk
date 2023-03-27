package backend

import (
	"errors"
	"io"
	"os"
	"sync"
)

var (
	errNotImplemented = errors.New("not implemented")
)

type TapeBackend struct {
	drive *os.File
	size  int64
	lock  sync.Mutex
}

func NewTapeBackend(drive *os.File, size int64) *TapeBackend {
	return &TapeBackend{drive, size, sync.Mutex{}}
}

func (b *TapeBackend) ReadAt(p []byte, off int64) (n int, err error) {
	b.lock.Lock()

	_, err = b.drive.Seek(off, io.SeekStart)
	if err != nil {
		b.lock.Unlock()

		return -1, err
	}

	n, err = b.drive.Read(p)

	b.lock.Unlock()

	return n, err
}

func (b *TapeBackend) WriteAt(p []byte, off int64) (n int, err error) {
	b.lock.Lock()

	_, err = b.drive.Seek(off, io.SeekStart)
	if err != nil {
		b.lock.Unlock()

		return -1, err
	}

	n, err = b.drive.Write(p)

	b.lock.Unlock()

	return n, err
}

func (b *TapeBackend) Size() (int64, error) {
	return b.size, nil
}

func (b *TapeBackend) Sync() error {
	return errNotImplemented
}
