//go:build linux && !cgo && amd64

package ioctl

// See /usr/include/sys/mtio.h

const (
	MTIOCGET = 0x80306d02

	MT_ST_BLKSIZE_MASK  = 0xffffff
	MT_ST_BLKSIZE_SHIFT = 0x0
)

type Mtget struct {
	mt_type   int64
	mt_resid  int64
	mt_dsreg  int64
	mt_gstat  int64
	mt_erreg  int64
	mt_fileno int32
	mt_blkno  int32
}

func (m Mtget) Dsreg() uint64 {
	return uint64(m.mt_dsreg)
}
