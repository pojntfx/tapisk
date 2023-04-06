package mtio

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

func SeekToBlock(drive *os.File, block int32) error {
	mtop := &ioctl.Mtop{}
	mtop.SetOp(ioctl.MTSEEK)
	mtop.SetCount(block)

	if _, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		drive.Fd(),
		ioctl.MTIOCTOP,
		uintptr(unsafe.Pointer(mtop)),
	); err != 0 {
		return err
	}

	return nil
}

func Tell(drive *os.File) (uint64, error) {
	mtpos := &ioctl.Mtpos{}

	if _, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		drive.Fd(),
		ioctl.MTIOCPOS,
		uintptr(unsafe.Pointer(mtpos)),
	); err != 0 {
		return 0, err
	}

	return mtpos.Blkno(), nil
}
