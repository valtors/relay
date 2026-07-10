package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	L = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	t.Run("Info", func(t *testing.T) {
		buf.Reset()
		Quiet = false
		Info("test info", "key", "value")
		assert.Contains(t, buf.String(), "level=INFO")
		assert.Contains(t, buf.String(), "msg=\"test info\"")
		assert.Contains(t, buf.String(), "key=value")
	})

	t.Run("Quiet", func(t *testing.T) {
		buf.Reset()
		Quiet = true
		Info("should be quiet")
		assert.Empty(t, buf.String())
	})

	t.Run("LogLevels", func(t *testing.T) {
		buf.Reset()
		Quiet = false
		Warn("test warn")
		assert.Contains(t, buf.String(), "level=WARN")

		buf.Reset()
		Error("test error")
		assert.Contains(t, buf.String(), "level=ERROR")

		buf.Reset()
		Debug("test debug")
		assert.Contains(t, buf.String(), "level=DEBUG")
	})
}
