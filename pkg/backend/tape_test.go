package backend

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/pojntfx/tapisk/pkg/index"
)

func TestReadAtTable(t *testing.T) {
	// Initialize tape
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

func TestReadAtOverwrites(t *testing.T) {
	// Initialize tape
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

	// Add overwrites
	expect := []byte("Overwrite")
	{
		// Read second block
		block := make([]byte, blockSize)
		if _, err := f.ReadAt(block, int64(blockSize)*2); err != nil {
			t.Fatal(err)
		}

		// Change block 2 bytes in
		copy(block[2:], expect)

		// Seek to end
		if _, err := f.Seek(0, io.SeekEnd); err != nil {
			t.Fatal(err)
		}

		curr, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			t.Fatal(err)
		}

		// Add new location to index
		if err := index.SetLocation(2, uint64(curr)/blockSize); err != nil {
			t.Fatal(err)
		}

		// Write back block
		if _, err := f.Write(block); err != nil {
			t.Fatal(err)
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
		name string
		run  func(tb *TapeBackend, t *testing.T) error
	}{
		{
			name: "Read and write less than a block",
			run: func(tb *TapeBackend, t *testing.T) error {
				// Read back (2nd block + 2 bytes in)
				{
					got := make([]byte, len(expect))
					if _, err := tb.ReadAt(got, (int64(blockSize)*2)+2); err != nil {
						return err
					}

					if !reflect.DeepEqual(got, expect) {
						return fmt.Errorf("ReadAt = %v, want %v", got, expect)
					}
				}

				// Read back (2nd block, first two bytes - should contains 2s)
				{
					got := make([]byte, 2)
					if _, err := tb.ReadAt(got, (int64(blockSize) * 2)); err != nil {
						return err
					}

					expect := bytes.Repeat([]byte{2}, 2)
					if !reflect.DeepEqual(got, expect) {
						return fmt.Errorf("ReadAt = %v, want %v", got, expect)
					}
				}

				// Read back (2nd block, after the overwrite - should contains 2s)
				{
					got := make([]byte, blockSize-2-uint64(len(expect)))
					if _, err := tb.ReadAt(got, (int64(blockSize)*2)+2+int64(len(expect))); err != nil {
						return err
					}

					expect := bytes.Repeat([]byte{2}, len(got))
					if !reflect.DeepEqual(got, expect) {
						return fmt.Errorf("ReadAt = %v, want %v", got, expect)
					}
				}

				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(tb, t); err != nil {
				t.Errorf("Unexpected error: %v", err)

				return
			}
		})
	}
}

func TestWriteAt(t *testing.T) {
	// Initialize tape
	f, err := ioutil.TempFile("", "tape_backend_test_")
	if err != nil {
		t.Fatal("Failed to create temporary file for tape drive:", err)
	}
	defer os.Remove(f.Name())

	blockSize := uint64(4096)
	numBlocks := uint64(3)
	size := int64(numBlocks) * int64(blockSize)

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

	tb := &TapeBackend{
		drive:     f,
		index:     index,
		size:      size,
		blocksize: blockSize,
		seekToBlock: func(drive *os.File, block int32) error {
			_, err := drive.Seek(int64(block)*int64(blockSize), io.SeekStart)

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

			return uint64(curr / int64(blockSize)), nil
		},
	}

	testCases := []struct {
		name string
		run  func(tb *TapeBackend, t *testing.T) error
	}{
		{
			name: "Read and write less than a block",
			run: func(tb *TapeBackend, t *testing.T) error {
				// Write and read initial contents
				{
					expect := []byte("Hello, world!")
					if _, err := tb.WriteAt(expect, 0); err != nil {
						return err
					}

					got := make([]byte, len(expect))
					if _, err := tb.ReadAt(got, 0); err != nil {
						return err
					}

					if !reflect.DeepEqual(got, expect) {
						return fmt.Errorf("ReadAt = %v, want %v", got, expect)
					}
				}

				// Overwrite and read part of it 2 bytes in
				{
					expect := []byte("ovrw")
					if _, err := tb.WriteAt(expect, 2); err != nil {
						return err
					}

					got := make([]byte, len(expect))
					if _, err := tb.ReadAt(got, 2); err != nil {
						return err
					}

					if !reflect.DeepEqual(got, expect) {
						return fmt.Errorf("ReadAt = %v, want %v", got, expect)
					}
				}

				// Read back part from before the overwrite
				{
					expect := []byte("He")

					got := make([]byte, len(expect))
					if _, err := tb.ReadAt(got, 0); err != nil {
						return err
					}

					if !reflect.DeepEqual(got, expect) {
						return fmt.Errorf("ReadAt = %v, want %v", got, expect)
					}
				}

				// Read back part from after the overwrite
				{
					expect := []byte(" world!")

					got := make([]byte, len(expect))
					if _, err := tb.ReadAt(got, 6); err != nil {
						return err
					}

					if !reflect.DeepEqual(got, expect) {
						return fmt.Errorf("ReadAt = %v, want %v", got, expect)
					}
				}

				return nil
			},
		},
		{
			name: "Read and write one block",
			run: func(tb *TapeBackend, t *testing.T) error {
				// Write and read exactly one block
				{
					expect := bytes.Repeat([]byte{5}, int(blockSize))
					if _, err := tb.WriteAt(expect, 0); err != nil {
						return err
					}

					got := make([]byte, len(expect))
					if _, err := tb.ReadAt(got, 0); err != nil {
						return err
					}

					if !reflect.DeepEqual(got, expect) {
						return fmt.Errorf("ReadAt = %v, want %v", got, expect)
					}
				}

				// Write and read exactly one block 5 bytes in
				{
					expect := bytes.Repeat([]byte{5}, int(blockSize))
					if _, err := tb.WriteAt(expect, 5); err != nil {
						return err
					}

					got := make([]byte, len(expect))
					if _, err := tb.ReadAt(got, 5); err != nil {
						return err
					}

					if !reflect.DeepEqual(got, expect) {
						return fmt.Errorf("ReadAt = %v, want %v", got, expect)
					}
				}

				return nil
			},
		},
		{
			name: "Read and write more than one block",
			run: func(tb *TapeBackend, t *testing.T) error {
				// Write and read more than one block
				{
					expect := append(bytes.Repeat([]byte{5}, int(blockSize)), bytes.Repeat([]byte{6}, int(blockSize))...)
					if _, err := tb.WriteAt(expect, 0); err != nil {
						return err
					}

					got := make([]byte, len(expect))
					if _, err := tb.ReadAt(got, 0); err != nil {
						return err
					}

					if !reflect.DeepEqual(got, expect) {
						return fmt.Errorf("ReadAt = %v, want %v", got, expect)
					}
				}

				// Write and read more than one block 5 bytes in
				{
					expect := append(bytes.Repeat([]byte{5}, int(blockSize)), bytes.Repeat([]byte{6}, int(blockSize))...)
					if _, err := tb.WriteAt(expect, 5); err != nil {
						return err
					}

					got := make([]byte, len(expect))
					if _, err := tb.ReadAt(got, 5); err != nil {
						return err
					}

					if !reflect.DeepEqual(got, expect) {
						return fmt.Errorf("ReadAt = %v, want %v", got, expect)
					}
				}

				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(tb, t); err != nil {
				t.Errorf("Unexpected error: %v", err)

				return
			}
		})
	}
}
