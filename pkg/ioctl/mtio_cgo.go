//go:build linux && cgo

package ioctl

/*
#include <sys/mtio.h>
*/
import "C"

const (
	MTIOCGET = C.MTIOCGET
	MTIOCTOP = C.MTIOCTOP
	MTIOCPOS = C.MTIOCPOS

	MT_ST_BLKSIZE_MASK  = C.MT_ST_BLKSIZE_MASK
	MT_ST_BLKSIZE_SHIFT = C.MT_ST_BLKSIZE_SHIFT

	MTSEEK = C.MTSEEK
	MTEOM  = C.MTEOM
)

type Mtget C.struct_mtget

func (m Mtget) Dsreg() uint64 {
	return uint64(m.mt_dsreg)
}

type Mtop C.struct_mtop

func (m *Mtop) SetOp(v int16) {
	m.mt_op = C.short(v)
}

func (m *Mtop) SetCount(v int32) {
	m.mt_count = C.int(v)
}

type Mtpos C.struct_mtpos

func (m Mtpos) Blkno() uint64 {
	return uint64(m.mt_blkno)
}
