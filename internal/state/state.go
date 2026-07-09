package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"relay/internal/logger"
)

type StageKey string

const (
	StagePMPlan    StageKey = "pm_plan"
	StageResearch  StageKey = "research"
	StageBrand     StageKey = "brand"
	StageUX        StageKey = "ux"
	StageGTM       StageKey = "gtm"
	StageAssembled StageKey = "assembled"
)

type SessionMeta struct {
	StartedAt       string            `json:"startedAt"`
	BriefPath       string            `json:"briefPath"`
	CompletedStages []StageKey        `json:"completedStages"`
	IterationCounts map[StageKey]int  `json:"iterationCounts"`
	HumanNotes      map[string]string `json:"humanNotes"`
}

const metaFile = ".session.meta.json"

func OutputDir() (string, error) {
	dir := filepath.Join(cwd(), "output")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}
	return dir, nil
}

func cwd() string {
	d, err := os.Getwd()
	if err != nil {
		return "."
	}
	return d
}

func WriteOutput(filename, content string) error {
	dir, err := OutputDir()
	if err != nil {
		return err
	}
	finalPath := filepath.Join(dir, filename)
	tmpPath := finalPath + ".tmp"

	if err := os.WriteFile(tmpPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write tmp %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename %s → %s: %w", tmpPath, finalPath, err)
	}

	logger.Info("written", "file", finalPath)
	return nil
}

func ReadOutput(filename string) (string, error) {
	dir, err := OutputDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, filename)
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stage output %q not found — run the prior stage first", filename)
	}
	if err != nil {
		return "", fmt.Errorf("read %s: %w", p, err)
	}
	return string(b), nil
}

func ReadBrief(briefPath string) (string, error) {
	if briefPath == "" {
		briefPath = filepath.Join(cwd(), "product_brief.md")
	}
	b, err := os.ReadFile(briefPath)
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("product_brief.md not found at %s — create this file in your project root", briefPath)
	}
	if err != nil {
		return "", fmt.Errorf("read brief: %w", err)
	}
	return string(b), nil
}

func OutputExists(filename string) bool {
	dir, err := OutputDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, filename))
	return err == nil
}

func lockPath(filename string) (string, error) {
	dir, err := OutputDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filename+".lock"), nil
}

func AcquireLock(filename string) error {
	lp, err := lockPath(filename)
	if err != nil {
		return err
	}

	currentPid := os.Getpid()
	if b, err := os.ReadFile(lp); err == nil {
		pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
		if err == nil {
			if pid == currentPid {
				return fmt.Errorf("%s is locked by process %d — retry in a moment", filename, pid)
			}
			proc, err := os.FindProcess(pid)
			if err == nil {
				if err := proc.Signal(os.Signal(nil)); err == nil {
					return fmt.Errorf("%s is locked by process %d — retry in a moment", filename, pid)
				}
			}
			logger.Warn("clearing stale lock", "file", filename, "pid", pid)
			os.Remove(lp)
		}
	}

	return os.WriteFile(lp, []byte(strconv.Itoa(currentPid)), 0o644)
}

func ReleaseLock(filename string) {
	if lp, err := lockPath(filename); err == nil {
		os.Remove(lp)
	}
}

func InitSession(briefPath string) error {
	return writeJSON(metaFile, newSessionMeta(briefPath))
}

func newSessionMeta(briefPath string) SessionMeta {
	return SessionMeta{
		StartedAt:       nowISO(),
		BriefPath:       briefPath,
		CompletedStages: []StageKey{},
		IterationCounts: map[StageKey]int{},
		HumanNotes:      map[string]string{},
	}
}

func LoadSession() (*SessionMeta, error) {
	raw, err := ReadOutput(metaFile)
	if err != nil {
		return nil, err
	}
	var meta SessionMeta
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return nil, fmt.Errorf("session meta is corrupted — delete ./output/.session.meta.json and restart")
	}
	if meta.IterationCounts == nil {
		meta.IterationCounts = map[StageKey]int{}
	}
	if meta.HumanNotes == nil {
		meta.HumanNotes = map[string]string{}
	}
	return &meta, nil
}

func MarkStageComplete(stage StageKey) error {
	meta, err := ensureSession()
	if err != nil {
		return err
	}
	for _, s := range meta.CompletedStages {
		if s == stage {
			return nil
		}
	}
	meta.CompletedStages = append(meta.CompletedStages, stage)
	return writeJSON(metaFile, meta)
}

func IsStageComplete(stage StageKey) bool {
	meta, err := LoadSession()
	if err != nil {
		return false
	}
	for _, s := range meta.CompletedStages {
		if s == stage {
			return true
		}
	}
	return false
}

func IncrementIteration(stage StageKey) (int, error) {
	meta, err := ensureSession()
	if err != nil {
		return 0, err
	}
	meta.IterationCounts[stage]++
	n := meta.IterationCounts[stage]
	if err := writeJSON(metaFile, meta); err != nil {
		return 0, err
	}
	return n, nil
}

func SaveHumanNote(checkpoint, notes string) error {
	meta, err := ensureSession()
	if err != nil {
		return err
	}
	meta.HumanNotes[checkpoint] = notes
	return writeJSON(metaFile, meta)
}

func ensureSession() (*SessionMeta, error) {
	meta, err := LoadSession()
	if err == nil {
		return meta, nil
	}
	if OutputExists(metaFile) {
		return nil, err
	}
	meta = &SessionMeta{
		StartedAt:       nowISO(),
		BriefPath:       "",
		CompletedStages: []StageKey{},
		IterationCounts: map[StageKey]int{},
		HumanNotes:      map[string]string{},
	}
	return meta, nil
}

func writeJSON(filename string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	return WriteOutput(filename, string(b))
}

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
