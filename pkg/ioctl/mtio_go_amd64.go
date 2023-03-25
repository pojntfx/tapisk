//go:build linux && !cgo && amd64

package ioctl

// See /usr/include/sys/mtio.h

const (
	MTIOCGET = 0x80306d02
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
