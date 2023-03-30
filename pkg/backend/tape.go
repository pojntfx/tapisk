package backend

import (
	"io"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/pojntfx/tapisk/pkg/ioctl"
)

type TapeBackend struct {
	drive     *os.File
	size      int64
	blocksize uint64
	lock      sync.Mutex
	compat    bool
	cursor    int
}

func NewTapeBackend(drive *os.File, size int64, blocksize uint64, compat bool) *TapeBackend {
	return &TapeBackend{drive, size, blocksize, sync.Mutex{}, compat, 0}
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

	if b.compat {
		if _, err := b.drive.Seek(off, io.SeekStart); err != nil {
			b.lock.Unlock()

			return -1, err
		}
	} else {
		if err = b.seekToBlock(int32(off / int64(b.blocksize))); err != nil {
			b.lock.Unlock()

			return -1, err
		}

		if err := b.discardBytes(off % int64(b.blocksize)); err != nil {
			b.lock.Unlock()

			return -1, err
		}
	}

	n, err = b.drive.Read(p)

	b.lock.Unlock()

	return n, err
}

func (b *TapeBackend) WriteAt(p []byte, off int64) (n int, err error) {
	b.lock.Lock()

	startBlock := off / int64(b.blocksize)
	startOffset := off % int64(b.blocksize)
	endBlock := (off + int64(len(p))) / int64(b.blocksize)
	endOffset := (off + int64(len(p))) % int64(b.blocksize)

	c := make([]byte, (endBlock-startBlock)*int64(b.blocksize))

	if b.compat {
		if _, err := b.drive.Seek(startBlock*int64(b.blocksize), io.SeekStart); err != nil {
			b.lock.Unlock()

			return -1, err
		}
	} else {
		if err = b.seekToBlock(int32(startBlock)); err != nil {
			b.lock.Unlock()

			return -1, err
		}
	}

	_, err = b.drive.Read(c)
	if err != nil {
		b.lock.Unlock()

		return -1, err
	}

	n = copy(c[startOffset:len(c)-int(endOffset)], p)

	if b.compat {
		if _, err := b.drive.Seek(startBlock*int64(b.blocksize), io.SeekStart); err != nil {
			b.lock.Unlock()

			return -1, err
		}
	} else {
		if err = b.seekToBlock(int32(startBlock)); err != nil {
			b.lock.Unlock()

			return -1, err
		}
	}

	for i := uint64(0); i < uint64((endBlock - startBlock)); i++ {
		_, err = b.drive.Write(c[i*b.blocksize : (i+1)*b.blocksize])
		if err != nil {
			b.lock.Unlock()

			return -1, err
		}
	}

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
