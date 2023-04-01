package index

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSetAndGetLocation(t *testing.T) {
	testCases := []struct {
		name     string
		block    uint64
		location uint64
	}{
		{
			name:     "Set and get location of block 1",
			block:    1,
			location: 100,
		},
		{
			name:     "Set and get location of block 2",
			block:    2,
			location: 200,
		},
		{
			name:     "Set and get location of block 3",
			block:    3,
			location: 300,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "bbolt_test")
			if err != nil {
				t.Fatal("Unable to create temporary directory:", err)
			}
			defer os.RemoveAll(tmpDir)

			file := filepath.Join(tmpDir, fmt.Sprintf("bbolt_test_%d.db", tc.block))

			index := NewBboltIndex(file, "test")
			err = index.Open()
			if err != nil {
				t.Fatal("Unable to open index:", err)
			}
			defer index.Close()

			err = index.SetLocation(tc.block, tc.location)
			if err != nil {
				t.Errorf("Unexpected error in SetLocation: %v", err)
			}

			location, err := index.GetLocation(tc.block)
			if err != nil {
				t.Errorf("Unexpected error in GetLocation: %v", err)
			}
			if location != tc.location {
				t.Errorf("Expected location to be %d, got %d", tc.location, location)
			}
		})
	}
}
