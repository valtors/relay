package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "relay_state_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	oldCwd, _ := os.Getwd()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer os.Chdir(oldCwd)

	t.Run("WriteAndReadOutput", func(t *testing.T) {
		filename := "test_output.txt"
		content := "hello world"

		err := WriteOutput(filename, content)
		assert.NoError(t, err)

		exists := OutputExists(filename)
		assert.True(t, exists)

		read, err := ReadOutput(filename)
		assert.NoError(t, err)
		assert.Equal(t, content, read)
	})

	t.Run("ReadNonExistentOutput", func(t *testing.T) {
		_, err := ReadOutput("missing.txt")
		assert.Error(t, err)
	})

	t.Run("SessionManagement", func(t *testing.T) {
		brief := "a test brief"
		err := InitSession(brief)
		assert.NoError(t, err)

		meta, err := LoadSession()
		assert.NoError(t, err)
		assert.Equal(t, brief, meta.BriefPath)

		err = MarkStageComplete("research")
		assert.NoError(t, err)

		meta, err = LoadSession()
		assert.NoError(t, err)
		assert.Contains(t, meta.CompletedStages, StageResearch)

		count, err := IncrementIteration("research")
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		count, err = IncrementIteration("research")
		assert.NoError(t, err)
		assert.Equal(t, 2, count)

		err = SaveHumanNote("research", "some note")
		assert.NoError(t, err)

		meta, err = LoadSession()
		assert.NoError(t, err)
		assert.Equal(t, "some note", meta.HumanNotes["research"])
	})

	t.Run("Locking", func(t *testing.T) {
		filename := "test_lock"
		err := AcquireLock(filename)
		assert.NoError(t, err)

		err = AcquireLock(filename)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is locked by process")

		ReleaseLock(filename)

		err = AcquireLock(filename)
		assert.NoError(t, err)
		ReleaseLock(filename)
	})

	t.Run("BriefReading", func(t *testing.T) {
		briefPath := filepath.Join(tmpDir, "brief.md")
		err := os.WriteFile(briefPath, []byte("brief content"), 0644)
		require.NoError(t, err)

		read, err := ReadBrief(briefPath)
		assert.NoError(t, err)
		assert.Equal(t, "brief content", read)
	})
}
