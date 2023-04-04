package backend

import (
	"errors"
	"os"
	"sync"

	"github.com/pojntfx/tapisk/pkg/index"
	"github.com/pojntfx/tapisk/pkg/mtio"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

type TapeBackend struct {
	drive *os.File
	index *index.BboltIndex

	size      int64
	blocksize uint64
	lock      sync.Mutex
}

func NewTapeBackend(
	drive *os.File,
	index *index.BboltIndex,

	size int64,
	blocksize uint64,
) *TapeBackend {
	return &TapeBackend{
		drive,
		index,

		size,
		blocksize,

		sync.Mutex{},
	}
}

func (b *TapeBackend) ReadAt(p []byte, off int64) (n int, err error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	startBlock := off / int64(b.blocksize)
	startOffset := off % int64(b.blocksize)
	endBlock := (off + int64(len(p))) / int64(b.blocksize)
	endOffset := (off + int64(len(p))) % int64(b.blocksize)

	out := make([]byte, (endBlock-startBlock)*int64(b.blocksize))

	for i := int64(0); i <= endBlock-startBlock; i++ {
		location, err := b.index.GetLocation(uint64(startBlock + i))
		if err != nil {
			return -1, err
		}

		if errors.Is(err, index.ErrNotExists) {
			continue
		}

		if err := mtio.SeekToBlock(b.drive, int32(location)); err != nil {
			return -1, err
		}

		if _, err = b.drive.Read(out[i*int64(b.blocksize):]); err != nil {
			return -1, err
		}
	}

	return copy(p, out[startOffset:len(out)-int(endOffset)]), nil
}

func (b *TapeBackend) WriteAt(p []byte, off int64) (n int, err error) {
	return -1, ErrNotImplemented
}

func (b *TapeBackend) Size() (int64, error) {
	return b.size, nil
}

func (b *TapeBackend) Sync() error {
	// nop, tapes are unbuffered
	return nil
}
