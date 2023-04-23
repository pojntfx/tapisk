package backend

import (
	"errors"
	"log"
	"os"
	"sync"

	"github.com/pojntfx/tapisk/pkg/index"
	"github.com/pojntfx/tapisk/pkg/mtio"
)

type TapeBackend struct {
	drive *os.File
	index *index.BboltIndex

	size      int64
	blocksize uint64

	seekToBlock func(drive *os.File, block int32) error
	seekToEOD   func(drive *os.File) error
	tell        func(drive *os.File) (uint64, error)

	lock sync.Mutex

	verbose bool

	lastOpWasRead bool
}

func NewTapeBackend(
	drive *os.File,
	index *index.BboltIndex,

	size int64,
	blocksize uint64,

	verbose bool,
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

		verbose,

		true,
	}
}

func (b *TapeBackend) ReadAt(p []byte, off int64) (n int, err error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.verbose {
		log.Printf("ReadAt() len(p)=%v off=%v", len(p), off)
	}

	return b.readAt(p, off)
}

func (b *TapeBackend) readAt(p []byte, off int64) (n int, err error) {
	startBlock := uint64(off) / b.blocksize
	lowerBound := uint64(off) % b.blocksize
	endBlock := (uint64(off) + uint64(len(p))) / b.blocksize
	endOffset := (uint64(off) + uint64(len(p))) % b.blocksize

	if b.verbose {
		log.Printf("readAt() len(p)=%v off=%v startBlock=%v endBlock=%v", len(p), off, startBlock, endBlock)
	}

	b.lastOpWasRead = true

	if uint64(len(p)) < uint64(b.blocksize) && startBlock == endBlock {
		endOffset = lowerBound + uint64(len(p))
	}

	out := make([]byte, (endBlock-startBlock+1)*b.blocksize)

	upperBound := uint64(len(out)) - endOffset + lowerBound
	if upperBound > uint64(len(out)) {
		upperBound = uint64(len(out))
	}

	for i := uint64(0); i <= endBlock-startBlock; i++ {
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

		if _, err = b.drive.Read(out[uint64(i)*b.blocksize : (uint64(i)+1)*b.blocksize]); err != nil {
			return -1, err
		}
	}

	return copy(p, out[lowerBound:upperBound]), nil
}

func (b *TapeBackend) WriteAt(p []byte, off int64) (n int, err error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	startBlock := uint64(off) / b.blocksize
	lowerBound := uint64(off) % b.blocksize
	endBlock := (uint64(off) + uint64(len(p))) / b.blocksize
	upperBound := (uint64(off) + uint64(len(p))) % b.blocksize

	if b.verbose {
		log.Printf("WriteAt() len(p)=%v off=%v startBlock=%v endBlock=%v", len(p), off, startBlock, endBlock)
	}

	needsUpdating := lowerBound != 0 || (upperBound != 0 && upperBound != b.blocksize)

	var buf []byte
	if needsUpdating {
		buf = make([]byte, ((endBlock-startBlock)+1)*b.blocksize)
		if _, err := b.readAt(buf, int64(startBlock)*int64(b.blocksize)); err != nil {
			return -1, err
		}
		copy(buf[lowerBound:], p)
		p = buf
	} else {
		if uint64(len(p))%b.blocksize != 0 {
			buf = make([]byte, ((endBlock - startBlock + 1) * b.blocksize))
			copy(buf, p)
			p = buf
		}
	}

	if b.lastOpWasRead {
		if err := b.seekToEOD(b.drive); err != nil {
			return -1, err
		}

		b.lastOpWasRead = false
	}

	numBlocksToWrite := uint64(len(p)) / b.blocksize
	for i := uint64(0); i < numBlocksToWrite; i++ {
		curr, err := b.tell(b.drive)
		if err != nil {
			return -1, err
		}

		if err := b.index.SetLocation(startBlock+i, curr); err != nil {
			return -1, err
		}

		if _, err := b.drive.Write(p[i*b.blocksize : (i+1)*b.blocksize]); err != nil {
			return -1, err
		}
	}

	return len(p), nil
}

func (b *TapeBackend) Size() (int64, error) {
	if b.verbose {
		log.Println("Size()")
	}

	return b.size, nil
}

func (b *TapeBackend) Sync() error {
	if b.verbose {
		log.Println("Sync()")
	}

	// nop, tapes are unbuffered
	return nil
}
