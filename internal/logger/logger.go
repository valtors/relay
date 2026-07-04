// Package logger writes exclusively to stderr.
// stdout is reserved for the MCP JSON protocol — writing anything there
// breaks the client connection silently and is extremely hard to debug.
package logger

import (
	"log/slog"
	"os"
)

var L *slog.Logger

func init() {
	L = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func Info(msg string, args ...any)  { L.Info(msg, args...) }
func Warn(msg string, args ...any)  { L.Warn(msg, args...) }
func Error(msg string, args ...any) { L.Error(msg, args...) }
func Debug(msg string, args ...any) { L.Debug(msg, args...) }

// Raw writes directly to stderr — used for interactive checkpoint UI.
func Raw(s string) { os.Stderr.WriteString(s) }
