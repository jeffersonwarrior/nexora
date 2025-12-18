package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	initOnce     sync.Once
	initialized  atomic.Bool
	errorHandler slog.Handler
)

func Setup(logFile string, debug bool) {
	initOnce.Do(func() {
		logDir := filepath.Dir(logFile)
		errorLogFile := filepath.Join(logDir, "nexora-errors.log")

		// Main log rotator
		logRotator := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    10,   // Max size in MB
			MaxBackups: 5,    // Number of backups
			MaxAge:     30,   // Days
			Compress:   true, // Enable compression
		}

		// ERROR log rotator - separate file for all ERROR level messages
		errorRotator := &lumberjack.Logger{
			Filename:   errorLogFile,
			MaxSize:    50, // Larger for errors
			MaxBackups: 10, // More backups
			MaxAge:     90, // Longer retention
			Compress:   true,
		}

		level := slog.LevelInfo
		if debug {
			level = slog.LevelDebug
		}

		// Main logger for INFO/DEBUG/WARN
		mainLogger := slog.NewJSONHandler(logRotator, &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
		})

		// ERROR logger - captures only ERROR level
		errorHandler = slog.NewJSONHandler(errorRotator, &slog.HandlerOptions{
			Level:     slog.LevelError,
			AddSource: true,
		})

		// Multi-handler: main for non-errors, error file for errors
		slog.SetDefault(slog.New(multiHandler{
			main:   mainLogger,
			errors: errorHandler,
		}))

		initialized.Store(true)
	})
}

type multiHandler struct {
	main   slog.Handler
	errors slog.Handler
}

func (mh multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return mh.main.Enabled(ctx, level)
}

func (mh multiHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= slog.LevelError {
		return mh.errors.Handle(ctx, r)
	}
	return mh.main.Handle(ctx, r)
}

func (mh multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return multiHandler{
		main:   mh.main.WithAttrs(attrs),
		errors: mh.errors.WithAttrs(attrs),
	}
}

func (mh multiHandler) WithGroup(name string) slog.Handler {
	return multiHandler{
		main:   mh.main.WithGroup(name),
		errors: mh.errors.WithGroup(name),
	}
}

func Initialized() bool {
	return initialized.Load()
}

func RecoverPanic(name string, cleanup func()) {
	if r := recover(); r != nil {
		// Create a timestamped panic log file
		timestamp := time.Now().Format("20060102-150405")
		filename := fmt.Sprintf("nexora-panic-%s-%s.log", name, timestamp)

		file, err := os.Create(filename)
		if err == nil {
			defer file.Close()

			// Write panic information and stack trace
			fmt.Fprintf(file, "Panic in %s: %v\n\n", name, r)
			fmt.Fprintf(file, "Time: %s\n\n", time.Now().Format(time.RFC3339))
			fmt.Fprintf(file, "Stack Trace:\n%s\n", debug.Stack())

			// Execute cleanup function if provided
			if cleanup != nil {
				cleanup()
			}
		}
	}
}
