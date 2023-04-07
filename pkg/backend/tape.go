package backend

import (
	"errors"
	"os"
	"sync"

	"github.com/pojntfx/tapisk/pkg/index"
	"github.com/pojntfx/tapisk/pkg/mtio"
)

func getBlockBuffer(blocksize int64, plength int64, off int64) (
	startBlock,
	endBlock int64,
	out []byte,
	lowerBound,
	upperBound int64,
) {
	startBlock = off / blocksize
	lowerBound = off % blocksize
	endBlock = (off + plength) / blocksize
	endOffset := (off + plength) % blocksize

	if uint64(plength) < uint64(blocksize) && startBlock == endBlock {
		endOffset = lowerBound + plength
	}

	out = make([]byte, (endBlock-startBlock+1)*blocksize)

	upperBound = int64(len(out)) - endOffset + lowerBound
	if upperBound > int64(len(out)) {
		upperBound = int64(len(out))
	}

	return
}

type TapeBackend struct {
	drive *os.File
	index *index.BboltIndex

	size      int64
	blocksize uint64

	seekToBlock func(drive *os.File, block int32) error
	seekToEOD   func(drive *os.File) error
	tell        func(drive *os.File) (uint64, error)

	lock sync.Mutex
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

		mtio.SeekToBlock,
		mtio.SeekToEOD,
		mtio.Tell,

		sync.Mutex{},
	}
}

func (b *TapeBackend) ReadAt(p []byte, off int64) (n int, err error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	startBlock, endBlock, out, lowerBound, upperBound := getBlockBuffer(int64(b.blocksize), int64(len(p)), off)

	for i := int64(0); i <= endBlock-startBlock; i++ {
		location, err := b.index.GetLocation(uint64(startBlock + i))
		if err != nil {
			if errors.Is(err, index.ErrNotExists) {
				continue
			}

			return -1, err
		}

		if err := b.seekToBlock(b.drive, int32(location)); err != nil {
			return -1, err
		}

		if _, err = b.drive.Read(out[i*int64(b.blocksize) : (i+1)*int64(b.blocksize)]); err != nil {
			return -1, err
		}
	}

	return copy(p, out[lowerBound:upperBound]), nil
}

func (b *TapeBackend) WriteAt(p []byte, off int64) (n int, err error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	startBlock, endBlock, out, lowerBound, upperBound := getBlockBuffer(int64(b.blocksize), int64(len(p)), off)

	if _, err := b.ReadAt(out[:lowerBound], off); err != nil {
		return -1, err
	}

	if _, err := b.ReadAt(out[upperBound:], off+upperBound); err != nil {
		return -1, err
	}

	copy(out[lowerBound:upperBound], p)

	for i := int64(0); i <= endBlock-startBlock; i++ {
		if err := b.seekToEOD(b.drive); err != nil {
			return -1, err
		}

		location, err := b.tell(b.drive)
		if err != nil {
			return -1, err
		}

		if err := b.index.SetLocation(uint64(startBlock+i), location); err != nil {
			return -1, err
		}

		if _, err = b.drive.Write(out[i*int64(b.blocksize) : (i+1)*int64(b.blocksize)]); err != nil {
			return -1, err
		}
	}

	return len(p), nil
}

func (b *TapeBackend) Size() (int64, error) {
	return b.size, nil
}

func (b *TapeBackend) Sync() error {
	// nop, tapes are unbuffered
	return nil
}
