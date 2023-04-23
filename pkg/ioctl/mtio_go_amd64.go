//go:build linux && !cgo && amd64

package ioctl

// See /usr/include/sys/mtio.h

const (
	MTIOCGET = 0x80306d02
	MTIOCTOP = 0x40086d01
	MTIOCPOS = 0x80086d03

	MT_ST_BLKSIZE_MASK  = 0xffffff
	MT_ST_BLKSIZE_SHIFT = 0x0

	MTSEEK = 0x16
	MTEOM  = 0xc
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

type Mtop struct {
	mt_op    int16
	mt_count int32
}

func (m *Mtop) SetOp(v int16) {
	m.mt_op = v
}

func (m *Mtop) SetCount(v int32) {
	m.mt_count = v
}

type Mtpos struct {
	mt_blkno int64
}

func (m Mtpos) Blkno() uint64 {
	return uint64(m.mt_blkno)
}
