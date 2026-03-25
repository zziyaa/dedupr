package dedup_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"dedupr/internal/dedup"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return path
}

func TestFind_EmptyInput(t *testing.T) {
	groups, err := dedup.Find(context.Background(), []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(groups))
	}
}

func TestFind_SingleFile(t *testing.T) {
	dir := t.TempDir()
	p := writeTempFile(t, dir, "a.txt", "hello")
	groups, err := dedup.Find(context.Background(), []string{p})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(groups))
	}
}

func TestFind_AllUnique_DifferentSizes(t *testing.T) {
	dir := t.TempDir()
	a := writeTempFile(t, dir, "a.txt", "short")
	b := writeTempFile(t, dir, "b.txt", "a longer string")
	c := writeTempFile(t, dir, "c.txt", "the longest string of all three")
	groups, err := dedup.Find(context.Background(), []string{a, b, c})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(groups))
	}
}

func TestFind_AllUnique_SameSize(t *testing.T) {
	dir := t.TempDir()
	a := writeTempFile(t, dir, "a.txt", "aaaa")
	b := writeTempFile(t, dir, "b.txt", "bbbb")
	groups, err := dedup.Find(context.Background(), []string{a, b})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(groups))
	}
}

func TestFind_TwoDuplicates(t *testing.T) {
	dir := t.TempDir()
	a := writeTempFile(t, dir, "a.txt", "duplicate content")
	b := writeTempFile(t, dir, "b.txt", "duplicate content")
	groups, err := dedup.Find(context.Background(), []string{a, b})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(groups[0].Paths))
	}
}

func TestFind_ThreeFiles_TwoDuplicates(t *testing.T) {
	dir := t.TempDir()
	a := writeTempFile(t, dir, "a.txt", "same")
	b := writeTempFile(t, dir, "b.txt", "same")
	c := writeTempFile(t, dir, "c.txt", "different")
	groups, err := dedup.Find(context.Background(), []string{a, b, c})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Paths) != 2 {
		t.Fatalf("expected 2 paths in group, got %d", len(groups[0].Paths))
	}
}

func TestFind_MultipleGroups(t *testing.T) {
	dir := t.TempDir()
	a1 := writeTempFile(t, dir, "a1.txt", "group A content")
	a2 := writeTempFile(t, dir, "a2.txt", "group A content")
	b1 := writeTempFile(t, dir, "b1.txt", "group B content!!")
	b2 := writeTempFile(t, dir, "b2.txt", "group B content!!")
	groups, err := dedup.Find(context.Background(), []string{a1, a2, b1, b2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
}

func TestFind_EmptyFiles(t *testing.T) {
	dir := t.TempDir()
	a := writeTempFile(t, dir, "a.txt", "")
	b := writeTempFile(t, dir, "b.txt", "")
	groups, err := dedup.Find(context.Background(), []string{a, b})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group for empty files, got %d", len(groups))
	}
	if len(groups[0].Paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(groups[0].Paths))
	}
}

func TestFind_NonexistentPath(t *testing.T) {
	_, err := dedup.Find(context.Background(), []string{"/nonexistent/path/file.txt"})
	if err == nil {
		t.Fatal("expected error for nonexistent path, got nil")
	}
}

func TestFind_ResultOrdering(t *testing.T) {
	dir := t.TempDir()
	// Create files with names that would be out of order alphabetically
	z := writeTempFile(t, dir, "z.txt", "same content")
	a := writeTempFile(t, dir, "a.txt", "same content")
	m := writeTempFile(t, dir, "m.txt", "same content")
	groups, err := dedup.Find(context.Background(), []string{z, a, m})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	paths := groups[0].Paths
	if len(paths) != 3 {
		t.Fatalf("expected 3 paths, got %d", len(paths))
	}
	// Paths must be sorted alphabetically
	for i := 1; i < len(paths); i++ {
		if paths[i] < paths[i-1] {
			t.Fatalf("paths not sorted: %v", paths)
		}
	}
}

func TestFind_CancelledContext(t *testing.T) {
	dir := t.TempDir()
	a := writeTempFile(t, dir, "a.txt", "duplicate content")
	b := writeTempFile(t, dir, "b.txt", "duplicate content")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := dedup.Find(ctx, []string{a, b})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
