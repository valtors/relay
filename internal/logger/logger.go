package logger

import (
	"log/slog"
	"os"
)

var L *slog.Logger

// Quiet suppresses Info-level startup messages (banner already shown).
var Quiet bool

func init() {
	L = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func Info(msg string, args ...any) {
	if Quiet {
		return
	}
	L.Info(msg, args...)
}
func Warn(msg string, args ...any)  { L.Warn(msg, args...) }
func Error(msg string, args ...any) { L.Error(msg, args...) }
func Debug(msg string, args ...any) { L.Debug(msg, args...) }

func Raw(s string) { os.Stderr.WriteString(s) }
