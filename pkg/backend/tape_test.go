package backend

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pojntfx/tapisk/pkg/index"
)

func TestReadAt(t *testing.T) {
	f, err := ioutil.TempFile("", "tape_backend_test_")
	if err != nil {
		t.Fatal("Failed to create temporary file for tape drive:", err)
	}
	defer os.Remove(f.Name())

	blockSize := uint64(4096)
	numBlocks := uint64(3)
	size := int64(numBlocks) * int64(blockSize)

	for i := uint64(0); i < numBlocks; i++ {
		data := make([]byte, blockSize)
		for j := range data {
			data[j] = byte(i)
		}

		if _, err := f.Write(data); err != nil {
			t.Fatal("Failed to write data to tape drive:", err)
		}
	}

	indexFile, err := ioutil.TempFile("", "index_test_")
	if err != nil {
		t.Fatal("Failed to create temporary file for index:", err)
	}
	defer os.Remove(indexFile.Name())

	index := index.NewBboltIndex(indexFile.Name(), "test")
	if err := index.Open(); err != nil {
		t.Fatal("Failed to open index:", err)
	}
	defer index.Close()

	for i := uint64(0); i < numBlocks; i++ {
		if err := index.SetLocation(i, i); err != nil {
			t.Fatal("Failed to set location in index:", err)
		}
	}

	tb := &TapeBackend{
		drive:     f,
		index:     index,
		size:      size,
		blocksize: blockSize,
		seekToBlock: func(drive *os.File, block int32) error {
			_, err := drive.Seek(int64(block)*int64(blockSize), io.SeekStart)

			return err
		},
	}

	testCases := []struct {
		name   string
		offset int64
		length int
		want   []byte
	}{
		{
			name:   "Read whole tape",
			offset: 0,
			length: int(size),
			want:   bytes.Join([][]byte{bytes.Repeat([]byte{0}, int(blockSize)), bytes.Repeat([]byte{1}, int(blockSize)), bytes.Repeat([]byte{2}, int(blockSize))}, []byte{}),
		},
		{
			name:   "Read first block",
			offset: 0,
			length: int(blockSize),
			want:   bytes.Repeat([]byte{0}, int(blockSize)),
		},
		{
			name:   "Read middle block",
			offset: int64(blockSize),
			length: int(blockSize),
			want:   bytes.Repeat([]byte{1}, int(blockSize)),
		},
		{
			name:   "Read last block",
			offset: int64(size) - int64(blockSize),
			length: int(blockSize),
			want:   bytes.Repeat([]byte{2}, int(blockSize)),
		},
		{
			name:   "Read first half of the first block",
			offset: 0,
			length: int(blockSize) / 2,
			want:   bytes.Repeat([]byte{0}, int(blockSize)/2),
		},
		{
			name:   "Read second half of the first block",
			offset: int64(blockSize) / 2,
			length: int(blockSize) / 2,
			want:   bytes.Repeat([]byte{0}, int(blockSize)/2),
		},
		{
			name:   "Read second half of the first block and first half of second block",
			offset: int64(blockSize) / 2,
			length: int(blockSize),
			want:   append(bytes.Repeat([]byte{0}, int(blockSize)/2), bytes.Repeat([]byte{1}, int(blockSize)/2)...),
		},
		{
			name:   "Read second half of the first block and first half of second block plus some bytes",
			offset: int64(blockSize) / 2,
			length: int(blockSize) + 12,
			want:   append(bytes.Repeat([]byte{0}, int(blockSize)/2), bytes.Repeat([]byte{1}, (int(blockSize)/2)+12)...),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := make([]byte, tc.length)
			if _, err := tb.ReadAt(got, tc.offset); err != nil {
				t.Errorf("Unexpected error: %v", err)

				return
			}

			if !bytes.Equal(got, tc.want) {
				t.Errorf("ReadAt(%d, %d) = %v, want %v", tc.offset, tc.length, got, tc.want)

				return
			}
		})
	}
}
