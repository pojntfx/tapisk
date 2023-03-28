package utils

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/pojntfx/tapisk/pkg/ioctl"
)

func GetBlocksize(drive *os.File) (uint64, error) {
	mtget := &ioctl.Mtget{}
	if _, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		drive.Fd(),
		ioctl.MTIOCGET,
		uintptr(unsafe.Pointer(mtget)),
	); err != 0 {
		return 0, err
	}

	return uint64((mtget.Dsreg() & ioctl.MT_ST_BLKSIZE_MASK) >> ioctl.MT_ST_BLKSIZE_SHIFT), nil
}
