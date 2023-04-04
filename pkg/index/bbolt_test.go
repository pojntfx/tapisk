package index

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSetAndGetLocation(t *testing.T) {
	testCases := []struct {
		name        string
		block       uint64
		location    uint64
		setLocation bool
		errExpect   error
	}{
		{
			name:        "Set and get location of block 1",
			block:       1,
			location:    100,
			setLocation: true,
			errExpect:   nil,
		},
		{
			name:        "Set and get location of block 2",
			block:       2,
			location:    200,
			setLocation: true,
			errExpect:   nil,
		},
		{
			name:        "Set and get location of block 3",
			block:       3,
			location:    300,
			setLocation: true,
			errExpect:   nil,
		},
		{
			name:        "Get location of block 3 that doesn't exist",
			block:       2,
			setLocation: false,
			errExpect:   ErrNotExists,
		},
		{
			name:        "Get location of block 3 that doesn't exist",
			block:       3,
			setLocation: false,
			errExpect:   ErrNotExists,
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

			if tc.setLocation {
				err = index.SetLocation(tc.block, tc.location)
				if err != nil {
					t.Errorf("Unexpected error in SetLocation: %v", err)
				}
			}

			location, err := index.GetLocation(tc.block)
			if !errors.Is(err, tc.errExpect) {
				t.Errorf("Unexpected error in GetLocation: %v", err)
			}

			if location != tc.location {
				t.Errorf("Expected location to be %d, got %d", tc.location, location)
			}
		})
	}
}
