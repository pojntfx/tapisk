//go:build linux && cgo

package ioctl

/*
#include <sys/mtio.h>
*/
import "C"

const (
	MTIOCGET = C.MTIOCGET
)

type Mtget C.struct_mtget

func (m Mtget) Blkno() int32 {
	return int32(m.mt_blkno)
}
