package dedup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"
)

// TestComputeHashes_ContextCancelled verifies that computeHashes stops launching
// new work and fills remaining results with context.Canceled when the context is
// cancelled while the semaphore is full.
//
// Strategy: create NumCPU+1 paths so that after NumCPU goroutines fill the
// semaphore, the main loop blocks on the select. We wait until all NumCPU
// goroutines have started, then cancel — guaranteeing the select sees ctx.Done()
// before any remaining path is processed.
func TestComputeHashes_ContextCancelled(t *testing.T) {
	numWorkers := runtime.NumCPU()
	// Need at least one path beyond the semaphore capacity so the loop reaches
	// the select while all slots are occupied.
	n := numWorkers + 1
	paths := make([]string, n)
	for i := range paths {
		paths[i] = fmt.Sprintf("fake-%d", i)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// started is signalled by each goroutine once it has entered the hash function.
	started := make(chan struct{}, numWorkers)

	blockingHash := func(path string) (string, error) {
		started <- struct{}{} // signal: this goroutine is now occupying a semaphore slot
		<-ctx.Done()          // block until cancelled
		return "", ctx.Err()
	}

	// Cancel once all worker slots are occupied, so the next iteration of the
	// main loop is guaranteed to hit the ctx.Done() branch.
	go func() {
		for range numWorkers {
			<-started
		}
		cancel()
	}()

	results := computeHashes(ctx, paths, blockingHash)

	// Every result must carry context.Canceled — either from the blocking hash
	// function (in-flight goroutines) or from the ctx.Done() fill (unstarted paths).
	for i, r := range results {
		if !errors.Is(r.err, context.Canceled) {
			t.Errorf("results[%d]: expected context.Canceled, got %v", i, r.err)
		}
	}
}

func TestHashContent_Deterministic_Partial(t *testing.T) {
	data := strings.NewReader("hello world")
	h1, err := hashContent(data, true)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	h2, err := hashContent(strings.NewReader("hello world"), true)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if h1 != h2 {
		t.Errorf("non-deterministic: %q != %q", h1, h2)
	}
	if len(h1) != 32 {
		t.Errorf("expected 32-char hex string, got %d chars: %q", len(h1), h1)
	}
}

func TestHashContent_Deterministic_Full(t *testing.T) {
	h1, err := hashContent(strings.NewReader("hello world"), false)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	h2, err := hashContent(strings.NewReader("hello world"), false)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if h1 != h2 {
		t.Errorf("non-deterministic: %q != %q", h1, h2)
	}
	if len(h1) != 32 {
		t.Errorf("expected 32-char hex string, got %d chars: %q", len(h1), h1)
	}
}

func TestHashContent_DifferentContents_Differ(t *testing.T) {
	ha, err := hashContent(strings.NewReader("content A"), false)
	if err != nil {
		t.Fatalf("hash a: %v", err)
	}
	hb, err := hashContent(strings.NewReader("content B"), false)
	if err != nil {
		t.Fatalf("hash b: %v", err)
	}
	if ha == hb {
		t.Errorf("different contents produced identical hash: %q", ha)
	}
}

// TestHashContent_PartialVsFull_SmallReader verifies that a reader shorter than
// partialHashSize produces the same digest in both modes, since both consume
// the same bytes.
func TestHashContent_PartialVsFull_SmallReader(t *testing.T) {
	data := []byte("small content")
	hp, err := hashContent(bytes.NewReader(data), true)
	if err != nil {
		t.Fatalf("partial: %v", err)
	}
	hf, err := hashContent(bytes.NewReader(data), false)
	if err != nil {
		t.Fatalf("full: %v", err)
	}
	if hp != hf {
		t.Errorf("small reader: partial %q != full %q", hp, hf)
	}
}

// TestHashContent_PartialVsFull_LargeReader verifies that a reader larger than
// partialHashSize produces different digests in partial vs full mode.
func TestHashContent_PartialVsFull_LargeReader(t *testing.T) {
	// Build a buffer larger than partialHashSize with distinct content in the
	// bytes beyond the limit so the two modes cannot accidentally agree.
	data := make([]byte, partialHashSize+512)
	for i := range data {
		data[i] = byte(i)
	}
	hp, err := hashContent(bytes.NewReader(data), true)
	if err != nil {
		t.Fatalf("partial: %v", err)
	}
	hf, err := hashContent(bytes.NewReader(data), false)
	if err != nil {
		t.Fatalf("full: %v", err)
	}
	if hp == hf {
		t.Errorf("large reader: partial and full produced identical hash %q", hp)
	}
}

// TestHashContent_ReadError verifies that a read error is propagated.
func TestHashContent_ReadError(t *testing.T) {
	boom := fmt.Errorf("disk exploded")
	r := &errorReader{err: boom}
	_, err := hashContent(r, false)
	if !errors.Is(err, boom) {
		t.Errorf("expected wrapped disk error, got: %v", err)
	}
}

type errorReader struct{ err error }

func (e *errorReader) Read(_ []byte) (int, error) { return 0, e.err }

// Ensure errorReader satisfies io.Reader at compile time.
var _ io.Reader = (*errorReader)(nil)
