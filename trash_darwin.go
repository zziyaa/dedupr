//go:build darwin

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// TrashResult holds the result of a trash operation for a single path.
type TrashResult struct {
	Path  string `json:"path"`
	Error string `json:"error,omitempty"`
}

// moveAllToTrash moves all given paths to the macOS Trash in a single
// AppleScript call via Finder.
func moveAllToTrash(paths []string) error {
	var sb strings.Builder
	sb.WriteString(`tell application "Finder" to delete {`)
	for i, p := range paths {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "POSIX file %q", p)
	}
	sb.WriteString("}")
	out, err := exec.Command("osascript", "-e", sb.String()).CombinedOutput()
	if err != nil {
		return fmt.Errorf("osascript: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// MoveToTrash moves the given paths to macOS Trash via Finder (non-destructive).
// It first attempts all paths in one batch. If the batch fails, it falls back
// to one-by-one calls to identify which individual paths succeeded and which failed.
func (a *App) MoveToTrash(paths []string) ([]TrashResult, error) {
	if len(paths) == 0 {
		return nil, nil
	}

	results := make([]TrashResult, len(paths))
	for i, p := range paths {
		results[i] = TrashResult{Path: p}
	}

	// Happy path: batch all in one Finder call.
	if err := moveAllToTrash(paths); err == nil {
		return results, nil
	}

	// Fallback: trash one at a time to surface per-file errors.
	for i, p := range paths {
		if err := moveAllToTrash([]string{p}); err != nil {
			results[i].Error = err.Error()
		}
	}
	return results, nil
}
