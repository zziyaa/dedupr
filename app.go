package main

import (
	"context"
	"dedupr/internal/dedup"
	"errors"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context

	findMu     sync.Mutex         // guards findCancel
	findCancel context.CancelFunc // nil when no scan is running

	emitter func(string, ...any) // defaults to runtime.EventsEmit
}

// NewApp creates a new App application struct.
func NewApp() *App {
	// nil fields will be populated in *startup* method,
	// that's where we obtain the context from Wails runtime.
	return &App{
		ctx: nil,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.emitter = func(name string, data ...any) {
		runtime.EventsEmit(ctx, name, data...)
	}
}

func (a *App) shutdown() {}

// beforeClose is called when the user attempts to close the application window.
// It returns true to prevent the close (e.g. when there's an ongoing operation, and the user cancels),
// or false to allow the application to exit.
func (a *App) beforeClose(ctx context.Context) bool {
	a.findMu.Lock()
	scanning := a.findCancel != nil
	a.findMu.Unlock()

	if !scanning {
		return false
	}

	result, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "Scan in progress",
		Message:       "A duplicate scan is currently running. Quitting will cancel it. Are you sure you want to quit?",
		Buttons:       []string{"Quit", "Cancel"},
		DefaultButton: "Cancel",
		CancelButton:  "Cancel",
	})
	if err != nil {
		// On error, allow close.
		log.Error().Err(err).Msg("Failed to show before-close dialog")
		return false
	}

	// Prevent close if the user did not confirm.
	return result != "Quit"
}

// FindDuplicates starts an async scan of the given paths. Results are delivered
// via Wails events: "dedup:complete", "dedup:error", or "dedup:cancelled".
// Returns an error immediately if a scan is already in progress.
func (a *App) FindDuplicates(paths []string) error {
	if len(paths) < 2 {
		return nil
	}

	a.findMu.Lock()
	if a.findCancel != nil {
		a.findMu.Unlock()
		return fmt.Errorf("a scan is already in progress")
	}
	findCtx, cancel := context.WithCancel(a.ctx)
	a.findCancel = cancel
	a.findMu.Unlock()

	log.Info().Int("paths", len(paths)).Msg("dedup scan started")

	go func() {
		groups, err := dedup.Find(findCtx, paths)

		a.findMu.Lock()
		a.findCancel = nil
		a.findMu.Unlock()

		switch {
		case err == nil:
			log.Info().Int("groups", len(groups)).Msg("dedup scan complete")
			a.emitter("dedup:complete", groups)
		case errors.Is(err, context.Canceled):
			log.Info().Msg("dedup scan cancelled")
			a.emitter("dedup:cancelled")
		default:
			log.Error().Err(err).Msg("dedup.Find failed")
			a.emitter("dedup:error", map[string]any{"error": err.Error()})
		}
	}()

	return nil
}

// CancelFindDuplicates cancels any in-progress scan. The frontend will receive
// a "dedup:cancelled" event when the goroutine exits.
func (a *App) CancelFindDuplicates() {
	a.findMu.Lock()
	cancel := a.findCancel
	a.findMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// DisplayFileDialog opens file selection dialog (allows multiple file selection),
// returns selected file paths or error if the dialog fails to open.
func (a *App) DisplayFileDialog() ([]string, error) {
	opts := runtime.OpenDialogOptions{
		Title:                "",
		CanCreateDirectories: false,
		ShowHiddenFiles:      false,
	}

	selectedFilePaths, err := runtime.OpenMultipleFilesDialog(a.ctx, opts)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open file dialog")
		return []string{}, fmt.Errorf("failed to open file dialog")
	}

	return selectedFilePaths, nil
}
