package backend

import (
	"errors"
	"io"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/pojntfx/tapisk/pkg/ioctl"
)

var (
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

func (b *TapeBackend) seekToBlock(block int32) error {
	mtop := &ioctl.Mtop{}
	mtop.SetOp(ioctl.MTSEEK)
	mtop.SetCount(block)

	if _, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		b.drive.Fd(),
		ioctl.MTIOCTOP,
		uintptr(unsafe.Pointer(mtop)),
	); err != 0 {
		return err
	}

	return nil
}

func (b *TapeBackend) discardBytes(count int64) error {
	if _, err := io.CopyN(io.Discard, b.drive, count); err != nil {
		return err
	}

	return nil
}

func (b *TapeBackend) ReadAt(p []byte, off int64) (n int, err error) {
	b.lock.Lock()

	if err = b.seekToBlock(int32(off / int64(b.blocksize))); err != nil {
		b.lock.Unlock()

		return -1, err
	}

	if err := b.discardBytes(off % int64(b.blocksize)); err != nil {
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

	if err = b.seekToBlock(int32(off / int64(b.blocksize))); err != nil {
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
	// nop, tapes are unbuffered
	return nil
}
