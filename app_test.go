package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// testEmitter records emitted events for test assertions.
type testEmitter struct {
	mu     sync.Mutex
	events []emittedEvent
}

type emittedEvent struct {
	name string
	data []any
}

func (e *testEmitter) emit(name string, data ...any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = append(e.events, emittedEvent{name: name, data: data})
}

func (e *testEmitter) waitFor(t *testing.T, eventName string, timeout time.Duration) emittedEvent {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		e.mu.Lock()
		for _, ev := range e.events {
			if ev.name == eventName {
				e.mu.Unlock()
				return ev
			}
		}
		e.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for event %q", eventName)
	return emittedEvent{}
}

func newTestApp() (*App, *testEmitter) {
	em := &testEmitter{}
	a := &App{
		ctx:     context.Background(),
		emitter: em.emit,
	}
	return a, em
}

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return p
}

func TestFindDuplicates_EmitsComplete(t *testing.T) {
	a, em := newTestApp()
	dir := t.TempDir()
	f1 := writeTempFile(t, dir, "a.txt", "hello")
	f2 := writeTempFile(t, dir, "b.txt", "world")

	err := a.FindDuplicates([]string{f1, f2})
	if err != nil {
		t.Fatalf("FindDuplicates returned unexpected error: %v", err)
	}

	ev := em.waitFor(t, "dedup:complete", 5*time.Second)
	if ev.name != "dedup:complete" {
		t.Fatalf("expected dedup:complete, got %q", ev.name)
	}
}

func TestFindDuplicates_ErrorWhenAlreadyRunning(t *testing.T) {
	a, _ := newTestApp()

	// Manually set findCancel to simulate an in-progress scan.
	a.findMu.Lock()
	_, cancel := context.WithCancel(context.Background())
	a.findCancel = cancel
	a.findMu.Unlock()
	defer cancel()

	dir := t.TempDir()
	f1 := writeTempFile(t, dir, "a.txt", "hello")
	f2 := writeTempFile(t, dir, "b.txt", "world")

	err := a.FindDuplicates([]string{f1, f2})
	if err == nil {
		t.Fatal("expected error when scan already in progress, got nil")
	}
}

func TestCancelFindDuplicates_EmitsCancelled(t *testing.T) {
	a, em := newTestApp()
	dir := t.TempDir()
	f1 := writeTempFile(t, dir, "a.txt", "hello")
	f2 := writeTempFile(t, dir, "b.txt", "world")

	err := a.FindDuplicates([]string{f1, f2})
	if err != nil {
		t.Fatalf("FindDuplicates returned unexpected error: %v", err)
	}

	a.CancelFindDuplicates()

	em.waitFor(t, "dedup:cancelled", 5*time.Second)
}

func TestFindDuplicates_CanRunAgainAfterCompletion(t *testing.T) {
	a, em := newTestApp()
	dir := t.TempDir()
	f1 := writeTempFile(t, dir, "a.txt", "hello")
	f2 := writeTempFile(t, dir, "b.txt", "world")

	if err := a.FindDuplicates([]string{f1, f2}); err != nil {
		t.Fatalf("first FindDuplicates: %v", err)
	}
	em.waitFor(t, "dedup:complete", 5*time.Second)

	// Reset recorded events.
	em.mu.Lock()
	em.events = nil
	em.mu.Unlock()

	if err := a.FindDuplicates([]string{f1, f2}); err != nil {
		t.Fatalf("second FindDuplicates: %v", err)
	}
	em.waitFor(t, "dedup:complete", 5*time.Second)
}
