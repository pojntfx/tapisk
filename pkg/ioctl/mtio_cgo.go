//go:build linux && cgo

package ioctl

/*
#include <sys/mtio.h>
*/
import "C"

const (
	MTIOCGET = C.MTIOCGET

	MT_ST_BLKSIZE_MASK  = C.MT_ST_BLKSIZE_MASK
	MT_ST_BLKSIZE_SHIFT = C.MT_ST_BLKSIZE_SHIFT
)

type Mtget C.struct_mtget

func (m Mtget) Dsreg() uint64 {
	return uint64(m.mt_dsreg)
}
