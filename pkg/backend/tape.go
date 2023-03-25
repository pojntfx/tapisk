package backend

import (
	"errors"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/pojntfx/tapisk/pkg/ioctl"
)

var (
	errNotImplemented = errors.New("not implemented")
)

type TapeBackend struct {
	device *os.File
	lock   sync.Mutex
}

func NewTapeBackend(device *os.File) *TapeBackend {
	return &TapeBackend{device, sync.Mutex{}}
}

func (b *TapeBackend) ReadAt(p []byte, off int64) (n int, err error) {
	return -1, errNotImplemented
}

func (b *TapeBackend) WriteAt(p []byte, off int64) (n int, err error) {
	return -1, errNotImplemented
}

func (b *TapeBackend) Size() (int64, error) {
	// TODO: We can't get the capacity of a tape drive, pass it in instead

	mtget := &ioctl.Mtget{}
	if _, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		b.device.Fd(),
		ioctl.MTIOCGET,
		uintptr(unsafe.Pointer(mtget)),
	); err != 0 {
		return -1, err
	}

	return int64(mtget.Blkno()), nil
}

func (b *TapeBackend) Sync() error {
	return errNotImplemented
}
