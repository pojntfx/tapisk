package backend

import (
	"errors"
	"io"
	"os"
	"sync"
)

var (
	ErrNotImplemented   = errors.New("not implemented")
	ErrInvalidChunkSize = errors.New("chunk does not match block size")
	ErrInvalidOffset    = errors.New("offset is not a multiple of 512")
)

type TapeBackend struct {
	drive     *os.File
	size      int64
	blocksize uint64
	lock      sync.Mutex
}

func NewTapeBackend(drive *os.File, size int64, blocksize uint64) *TapeBackend {
	return &TapeBackend{drive, size, blocksize, sync.Mutex{}}
}

func (b *TapeBackend) ReadAt(p []byte, off int64) (n int, err error) {
	if len(p) != int(b.blocksize) {
		return -1, ErrInvalidChunkSize
	}

	if off%int64(b.blocksize) != 0 {
		return -1, ErrInvalidOffset
	}

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
	if len(p) != int(b.blocksize) {
		return -1, ErrInvalidChunkSize
	}

	if off%int64(b.blocksize) != 0 {
		return -1, ErrInvalidOffset
	}

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
	return ErrNotImplemented
}
