//go:build integration && darwin

package main

import (
	"errors"
	"io/fs"
	"os"
	"testing"
)

func TestMoveToTrash(t *testing.T) {
	a, _ := newTestApp()

	f, err := os.CreateTemp(t.TempDir(), "dup*")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	path := f.Name()
	f.Close()

	results, err := a.MoveToTrash([]string{path})
	if err != nil {
		t.Fatalf("MoveToTrash error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error != "" {
		t.Fatalf("expected no error, got: %s", results[0].Error)
	}
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected file to be gone after trash, got stat err: %v", err)
	}
}
