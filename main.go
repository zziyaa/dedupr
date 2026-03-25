package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"runtime"

	"dedupr/internal/utils"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/rs/zerolog/log"
)

//go:embed all:frontend/dist
var assets embed.FS

// version and buildTime are set with ldflags during the build process.
var (
	version   string
	buildTime string
	buildType string
)

func main() {
	err := initLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	log.Info().
		Str("version", version).
		Str("build_time", buildTime).
		Str("os", runtime.GOOS).
		Str("arch", runtime.GOARCH).
		Int("num_cpu", runtime.NumCPU()).
		Str("go_version", runtime.Version()).
		Str("build_type", buildType).
		Str("log_level", log.Logger.GetLevel().String()).
		Msgf("%s started", utils.AppName)

	app := NewApp()

	if err = wails.Run(wailsOptions(app)); err != nil {
		log.Fatal().Err(err).Msg("Failed to run application")
	}
}

// handleStartup performs initialization tasks when the application starts up.
func handleStartup(app *App) func(ctx context.Context) {
	return func(ctx context.Context) {
		log.Info().Msg("Application starting up...")
		app.startup(ctx)
	}
}

// handleCleanup performs cleanup tasks when the application is shutting down.
func handleCleanup(app *App) func(ctx context.Context) {
	return func(ctx context.Context) {
		log.Info().Msg("Application shutting down...")

		// Save app state before closing
		app.shutdown()
	}
}

// wailsOptions returns the Wails application options configured for the App on macOS.
func wailsOptions(app *App) *options.App {
	macOptions := &mac.Options{
		TitleBar:             mac.TitleBarHidden(),
		Appearance:           mac.NSAppearanceNameDarkAqua,
		WebviewIsTransparent: true,
		WindowIsTranslucent:  true,
		About: &mac.AboutInfo{
			Title:   utils.AppName,
			Message: fmt.Sprintf("Version %s", version),
			Icon:    nil,
		},
	}

	return &options.App{
		Title:             utils.AppName,
		Width:             800,
		Height:            400,
		DisableResize:     true,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		AssetServer:       &assetserver.Options{Assets: assets},
		DragAndDrop:       &options.DragAndDrop{EnableFileDrop: true},
		BackgroundColour:  options.NewRGBA(255, 255, 255, 0),
		OnStartup:         handleStartup(app),
		OnShutdown:        handleCleanup(app),
		OnBeforeClose:     app.beforeClose,
		Bind:              []any{app},
		Logger:            &ZerologAdapter{}, // Use custom logger
		LogLevel:          3,                 // Set log level to info
		Menu:              createAppMenu(),
		Mac:               macOptions,
	}
}

// createAppMenu appends the Help menu to the default application menu for macOS
func createAppMenu() *menu.Menu {
	appMenu := menu.NewMenu()

	// Add the default application menu for macOS (required for proper menu bar behavior)
	if runtime.GOOS == "darwin" {
		appMenu.Append(menu.AppMenu())
		appMenu.Append(menu.EditMenu())
		appMenu.Append(menu.WindowMenu())
	}

	// Help menu
	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("Third-party Software", nil, func(_ *menu.CallbackData) {
		// Open help documentation or emit event to frontend
	})

	return appMenu
}
