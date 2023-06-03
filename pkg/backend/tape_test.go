package backend

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/tapisk/pkg/index"
)

func TestTape(t *testing.T) {
	chunks.TestChunkedReadWriterAtGeneric(
		t,
		func(chunkSize, chunkCount int64) (chunks.ReadWriterAt, func() error, error) {
			size := int64(chunkCount) * int64(chunkSize)

			f, err := ioutil.TempFile("", "tape_backend_test_")
			if err != nil {
				return nil, nil, err
			}

			indexFile, err := ioutil.TempFile("", "index_test_")
			if err != nil {
				return nil, nil, err
			}

			index := index.NewBboltIndex(indexFile.Name(), "test")
			if err := index.Open(); err != nil {
				return nil, nil, err
			}

			return &TapeBackend{
					drive:     f,
					index:     index,
					size:      size,
					blocksize: uint64(chunkSize),
					seekToBlock: func(drive *os.File, block int32) error {
						_, err := drive.Seek(int64(block)*int64(chunkSize), io.SeekStart)

						return err
					},
					seekToEOD: func(drive *os.File) error {
						_, err := drive.Seek(0, io.SeekEnd)

						return err
					},
					tell: func(drive *os.File) (uint64, error) {
						curr, err := drive.Seek(0, io.SeekCurrent)
						if err != nil {
							return 0, err
						}

						return uint64(curr / int64(chunkSize)), nil
					},
				},
				func() error {
					if err := os.Remove(f.Name()); err != nil {
						return err
					}

					if err := os.Remove(indexFile.Name()); err != nil {
						return err
					}

					return index.Close()
				}, nil
		},
		[]int64{1, 2, 8, 64, 256, 512, 4096},
		[]int64{1, 10, 100},
	)
}
