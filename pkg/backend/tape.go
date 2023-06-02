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

	block := uint64(off) / b.blocksize

	if b.verbose {
		log.Printf("ReadAt() len(p)=%v off=%v block=%v", len(p), off, block)
	}

	b.lastOpWasRead = true

	location, err := b.index.GetLocation(block)
	if err != nil {
		if errors.Is(err, index.ErrNotExists) {
			return len(p), nil
		}

		return -1, err
	}

	if err := b.seekToBlock(b.drive, int32(location)); err != nil {
		return -1, err
	}

	return b.drive.Read(p)
}

func (b *TapeBackend) WriteAt(p []byte, off int64) (n int, err error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	block := uint64(off) / b.blocksize

	if b.verbose {
		log.Printf("WriteAt() len(p)=%v off=%v block=%v", len(p), off, block)
	}

	if b.lastOpWasRead {
		if err := b.seekToEOD(b.drive); err != nil {
			return -1, err
		}

		b.lastOpWasRead = false
	}

	curr, err := b.tell(b.drive)
	if err != nil {
		return -1, err
	}

	if err := b.index.SetLocation(block, curr); err != nil {
		return -1, err
	}

	if _, err := b.drive.Write(p); err != nil {
		return -1, err
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
