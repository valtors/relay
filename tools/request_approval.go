// Package tools — request_approval implements the human-in-the-loop checkpoint.
//
// Behaviour:
//   - Writes a checkpoint_<NAME>.md file before blocking on input so the human
//     always has a persistent artifact to review.
//   - Prints the prompt UI to stderr (stdout is reserved for the MCP wire).
//   - Reads one line from stdin: empty/"approve" → approve; "iterate <notes>" →
//     iterate with notes; anything else → approve treating the line as a note.
//   - Non-interactive stdin (CI / piped input) auto-approves so the pipeline can
//     still complete in unattended environments.
//   - Optional CHECKPOINT_TIMEOUT_MINUTES env var auto-approves on timeout.
//   - Returns a JSON-encoded CheckpointResult so run_workflow can parse it.
package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"relay/internal/logger"
	"relay/internal/state"
)

const (
	border  = "════════════════════════════════════════════════════════════"
	divider = "────────────────────────────────────────────────────────────"

	maxIterateNotesLen = 2000
)

// CheckpointResult holds the human's decision and notes.
type CheckpointResult struct {
	Decision string `json:"decision"` // "approve" | "iterate"
	Notes    string `json:"notes"`
}

func RequestApproval(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	checkpoint := strings.TrimSpace(req.GetString("checkpoint", ""))
	if checkpoint == "" {
		return mcp.NewToolResultError("checkpoint is required (H1, H2, H3, or H4)"), nil
	}

	summary := req.GetString("summary", "")

	questions := extractQuestions(req)
	if len(questions) == 0 {
		questions = []string{
			"Does this output match your expectations?",
			"Ready to proceed?",
		}
	}

	checkpointFile := fmt.Sprintf("checkpoint_%s.md", checkpoint)

	// Write checkpoint to disk before blocking on input — guarantees a
	// reviewable artifact even if the user closes the terminal.
	doc := buildCheckpointDoc(checkpoint, summary, questions)
	if err := state.WriteOutput(checkpointFile, doc); err != nil {
		logger.Warn("could not write checkpoint file", "err", err)
	}

	// Render the prompt to stderr — stdout is the MCP JSON-RPC wire.
	logger.Raw("\n" + border + "\n")
	logger.Raw(fmt.Sprintf("  CHECKPOINT %s — REVIEW REQUIRED\n", checkpoint))
	logger.Raw(border + "\n\n")
	if summary != "" {
		logger.Raw(summary + "\n\n")
	}
	logger.Raw("Questions:\n")
	for i, q := range questions {
		logger.Raw(fmt.Sprintf("  %d. %s\n", i+1, q))
	}
	logger.Raw("\n" + divider + "\n")
	logger.Raw("  \"approve\" to continue\n")
	logger.Raw("  \"iterate <your notes>\" to redo this stage\n")
	logger.Raw(divider + "\n> ")

	result, err := waitForDecision()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("checkpoint input error: %v", err)), nil
	}

	logger.Raw("\n" + divider + "\n")
	logger.Raw(fmt.Sprintf("  Decision: %s\n", strings.ToUpper(result.Decision)))
	if result.Notes != "" {
		logger.Raw(fmt.Sprintf("  Notes:    %s\n", result.Notes))
	}
	logger.Raw(divider + "\n\n")

	withDecision := doc + fmt.Sprintf(
		"\n\n---\n\n## Decision\n\n**%s**\n\nNotes: %s\n",
		strings.ToUpper(result.Decision),
		notesOrNone(result.Notes),
	)
	if err := state.WriteOutput(checkpointFile, withDecision); err != nil {
		logger.Warn("could not update checkpoint file", "err", err)
	}

	if err := state.SaveHumanNote(checkpoint, result.Notes); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("save human note: %v", err)), nil
	}

	b, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(b)), nil
}

// extractQuestions pulls the questions array from the request, accepting
// either a []string (when the SDK already typed it) or a []any of strings.
func extractQuestions(req mcp.CallToolRequest) []string {
	if qs := req.GetStringSlice("questions", nil); len(qs) > 0 {
		return qs
	}
	args := req.GetArguments()
	if args == nil {
		return nil
	}
	raw, ok := args["questions"]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, q := range v {
			if s, ok := q.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func waitForDecision() (*CheckpointResult, error) {
	// Non-interactive stdin (CI, piped input) — auto-approve so the pipeline
	// completes in unattended runs instead of hanging forever.
	fi, err := os.Stdin.Stat()
	if err == nil && (fi.Mode()&os.ModeCharDevice) == 0 {
		logger.Warn("non-interactive stdin — auto-approving checkpoint")
		return &CheckpointResult{
			Decision: "approve",
			Notes:    "(auto-approved: non-interactive)",
		}, nil
	}

	timeoutMin := getEnvInt("CHECKPOINT_TIMEOUT_MINUTES", 0)

	type readResult struct {
		r   *CheckpointResult
		err error
	}
	ch := make(chan readResult, 1)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		// Allow long iterate notes (default scanner buffer is 64KB which is
		// fine, but be explicit so the cap matches maxIterateNotesLen+slack).
		scanner.Buffer(make([]byte, 0, 8*1024), 64*1024)

		if !scanner.Scan() {
			if serr := scanner.Err(); serr != nil {
				ch <- readResult{err: serr}
				return
			}
			// EOF — stdin closed by the caller mid-prompt.
			ch <- readResult{r: &CheckpointResult{
				Decision: "approve",
				Notes:    "(stdin closed)",
			}}
			return
		}

		line := strings.TrimSpace(scanner.Text())
		ch <- readResult{r: parseDecisionLine(line)}
	}()

	if timeoutMin > 0 {
		select {
		case res := <-ch:
			return res.r, res.err
		case <-time.After(time.Duration(timeoutMin) * time.Minute):
			logger.Warn("checkpoint timeout — auto-approving", "minutes", timeoutMin)
			return &CheckpointResult{Decision: "approve", Notes: "(timed out)"}, nil
		}
	}

	res := <-ch
	return res.r, res.err
}

func buildCheckpointDoc(checkpoint, summary string, questions []string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# Checkpoint %s\n\n## PM Agent Summary\n\n%s\n\n## Questions\n\n",
		checkpoint, summary)
	for i, q := range questions {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, q)
	}
	return sb.String()
}

func notesOrNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

// parseDecisionLine maps one trimmed user input line to a CheckpointResult.
//
// Rules (matches spec):
//   - "" or "approve" (case-insensitive)        → approve, no notes
//   - "iterate" alone                            → iterate, no notes
//   - "iterate <text>"                           → iterate, notes = <text> (capped)
//   - anything else                              → approve, notes = original line
func parseDecisionLine(line string) *CheckpointResult {
	lower := strings.ToLower(line)
	switch {
	case lower == "" || lower == "approve":
		return &CheckpointResult{Decision: "approve", Notes: ""}

	case lower == "iterate" || strings.HasPrefix(lower, "iterate "):
		notes := ""
		if len(line) > len("iterate") {
			notes = strings.TrimSpace(line[len("iterate"):])
		}
		if len(notes) > maxIterateNotesLen {
			notes = notes[:maxIterateNotesLen]
		}
		return &CheckpointResult{Decision: "iterate", Notes: notes}

	default:
		return &CheckpointResult{Decision: "approve", Notes: line}
	}
}

// getEnvInt parses an int env var, returning the default on missing/invalid.
func getEnvInt(key string, def int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return def
	}
	return n
}
