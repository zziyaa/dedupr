package main

import (
	"fmt"
	stdlog "log" // Alias standard logger
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"dedupr/internal/utils"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var initLoggerOnce sync.Once

// initLogger initializes the zerolog's global logger.
func initLogger() error {
	var initErr error

	initLoggerOnce.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
		logLevel := zerolog.InfoLevel // Default log level

		// Check for environment variable for log level, which takes precedence over the default.
		envLogLevel := os.Getenv(fmt.Sprintf("%s_LOG_LEVEL", strings.ToUpper(utils.AppName)))
		if envLogLevel != "" {
			if l, err := zerolog.ParseLevel(envLogLevel); err == nil {
				logLevel = l
			}
		}

		appDataDir, err := utils.AppDataDir()
		if err != nil {
			initErr = err
			return
		}

		// Create logs directory if it doesn't exist
		logsDir := filepath.Join(appDataDir, "logs")
		err = os.MkdirAll(logsDir, 0755)
		if err != nil {
			initErr = fmt.Errorf("failed to create log directory '%s': %v", logsDir, err)
			return
		}

		logFile := filepath.Join(logsDir, fmt.Sprintf("%s.log", utils.AppName))

		fileLogger := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    10,   // megabytes
			MaxBackups: 3,    // number of old log files to keep
			MaxAge:     28,   // days
			Compress:   true, // compress rotated files
		}

		// Configure Console Writer (pretty-printed for development)
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

		// Combine console and file writers using MultiLevelWriter
		multiWriter := zerolog.MultiLevelWriter(consoleWriter, fileLogger)

		// Configure Zerolog global logger
		log.Logger = zerolog.New(multiWriter).Level(zerolog.Level(logLevel)).With().Timestamp().Logger()

		// Note: zerolog's global log level doesn't affect Wails logger level.
		zerolog.SetGlobalLevel(zerolog.Level(logLevel))

		// Redirect standard Go log output to zerolog
		stdlog.SetFlags(0)
		stdlog.SetOutput(log.Logger)
	})

	return initErr
}

// --- Custom Zerolog Adapter for Wails ---

// ZerologAdapter wraps zerolog to satisfy the wails.Logger interface
// It doesn't need changes as it uses the global zerolog instance configured above.
type ZerologAdapter struct{}

// Print implements logger.Logger. Maps to Debug level.
func (l *ZerologAdapter) Print(message string)   { log.Debug().Msg(message) }
func (l *ZerologAdapter) Trace(message string)   { log.Trace().Msg(message) }
func (l *ZerologAdapter) Debug(message string)   { log.Debug().Msg(message) }
func (l *ZerologAdapter) Info(message string)    { log.Info().Msg(message) }
func (l *ZerologAdapter) Warning(message string) { log.Warn().Msg(message) }
func (l *ZerologAdapter) Error(message string)   { log.Error().Msg(message) }
func (l *ZerologAdapter) Fatal(message string)   { log.Fatal().Msg(message) }
