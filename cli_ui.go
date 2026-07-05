package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type cliUI struct {
	color bool
}

type commandSummary struct {
	name        string
	usage       string
	description string
}

type namedValue struct {
	label string
	value string
}

func newCLIUI(w io.Writer, noColor bool) cliUI {
	return cliUI{
		color: !noColor && os.Getenv("NO_COLOR") == "" && isTerminalWriter(w),
	}
}

func isTerminalWriter(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

func extractGlobalNoColorFlag(args []string) ([]string, bool) {
	filtered := make([]string, 0, len(args))
	noColor := false
	for _, arg := range args {
		if arg == "--no-color" {
			noColor = true
			continue
		}
		filtered = append(filtered, arg)
	}
	return filtered, noColor
}

func (ui cliUI) bold(text string) string {
	return ui.wrap(text, "1")
}

func (ui cliUI) green(text string) string {
	return ui.wrap(text, "32")
}

func (ui cliUI) cyan(text string) string {
	return ui.wrap(text, "36")
}

func (ui cliUI) dim(text string) string {
	return ui.wrap(text, "2")
}

func (ui cliUI) wrap(text, code string) string {
	if !ui.color || text == "" {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func renderBox(rawLines, styledLines []string) string {
	width := 0
	for _, line := range rawLines {
		if l := runeLen(line); l > width {
			width = l
		}
	}

	var b strings.Builder
	b.WriteString("  ╭")
	b.WriteString(strings.Repeat("─", width+4))
	b.WriteString("╮\n")
	for i, line := range styledLines {
		padding := strings.Repeat(" ", width-runeLen(rawLines[i]))
		b.WriteString("  │  ")
		b.WriteString(line)
		b.WriteString(padding)
		b.WriteString("  │\n")
	}
	b.WriteString("  ╰")
	b.WriteString(strings.Repeat("─", width+4))
	b.WriteString("╯\n")
	return b.String()
}

func formatStatusBox(ui cliUI, version string, toolCount, categories int) string {
	rawLines := []string{
		"relay v" + version,
		"",
		fmt.Sprintf("%-11s %d (%d categories)", "Tools:", toolCount, categories),
		fmt.Sprintf("%-11s stdio | http", "Transport:"),
		fmt.Sprintf("%-11s ready", "Status:"),
	}

	styledLines := []string{
		ui.bold(rawLines[0]),
		rawLines[1],
		ui.bold("Tools:") + rawLines[2][len("Tools:"):],
		ui.bold("Transport:") + rawLines[3][len("Transport:"):],
		ui.bold("Status:") + rawLines[4][len("Status:"):len(rawLines[4])-len("ready")] + ui.green("ready"),
	}

	return renderBox(rawLines, styledLines)
}

func printCommandGroup(w io.Writer, ui cliUI, title string, commands []commandSummary) {
	fmt.Fprintln(w, ui.bold(title))
	nameWidth := 0
	for _, cmd := range commands {
		if l := runeLen(cmd.usage); l > nameWidth {
			nameWidth = l
		}
	}
	for _, cmd := range commands {
		fmt.Fprintf(w, "  %-*s  %s\n", nameWidth, cmd.usage, cmd.description)
	}
	fmt.Fprintln(w)
}

func runeLen(value string) int {
	return len([]rune(value))
}

func categoryMeta(category string) (icon, label string) {
	switch category {
	case "data":
		return ">", "Data"
	case "file":
		return ">", "File"
	case "image":
		return ">", "Image"
	case "pdf":
		return ">", "PDF"
	case "text":
		return ">", "Text"
	case "web":
		return ">", "Web"
	case "workflow":
		return ">", "Workflow"
	default:
		return ">", titleCaseWord(category)
	}
}

func shortToolName(entry toolInfo) string {
	prefix := entry.Category + "_"
	name := entry.Name
	if strings.HasPrefix(name, prefix) {
		name = strings.TrimPrefix(name, prefix)
	}

	switch entry.Name {
	case "data_csv_to_json":
		return "csv→json"
	case "data_json_to_csv":
		return "json→csv"
	case "data_json_format":
		return "format json"
	case "data_json_query":
		return "query json"
	case "pdf_extract_text":
		return "extract text"
	case "pdf_page_count":
		return "page count"
	case "pdf_extract_pages":
		return "extract pages"
	case "run_workflow":
		return "run workflow"
	case "pm_plan":
		return "PM plan"
	case "run_research":
		return "research"
	case "run_brand":
		return "brand"
	case "run_ux":
		return "UX"
	case "run_gtm":
		return "GTM"
	case "request_approval":
		return "approval"
	case "assemble_plan":
		return "assemble plan"
	}

	return strings.ReplaceAll(name, "_", " ")
}

func sortToolsForDisplay(entries []toolInfo) []toolInfo {
	sorted := make([]toolInfo, len(entries))
	copy(sorted, entries)

	sort.SliceStable(sorted, func(i, j int) bool {
		leftRank := toolDisplayRank(sorted[i])
		rightRank := toolDisplayRank(sorted[j])
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		return shortToolName(sorted[i]) < shortToolName(sorted[j])
	})

	return sorted
}

func toolDisplayRank(entry toolInfo) int {
	order := map[string]map[string]int{
		"data": {
			"data_csv_to_json": 0,
			"data_json_to_csv": 1,
			"data_json_format": 2,
			"data_json_query":  3,
		},
		"file": {
			"file_read":  0,
			"file_write": 1,
			"file_list":  2,
			"file_size":  3,
			"file_hash":  4,
			"file_zip":   5,
			"file_unzip": 6,
		},
		"image": {
			"image_info":      0,
			"image_resize":    1,
			"image_crop":      2,
			"image_convert":   3,
			"image_rotate":    4,
			"image_grayscale": 5,
			"image_flip":      6,
		},
		"pdf": {
			"pdf_info":          0,
			"pdf_page_count":    1,
			"pdf_extract_text":  2,
			"pdf_extract_pages": 3,
			"pdf_merge":         4,
			"pdf_split":         5,
		},
		"text": {
			"text_word_count":    0,
			"text_replace":       1,
			"text_extract_regex": 2,
			"text_base64_encode": 3,
			"text_base64_decode": 4,
			"text_md_to_html":    5,
		},
		"web": {
			"web_fetch":  0,
			"web_status": 1,
		},
		"workflow": {
			"run_workflow":     0,
			"pm_plan":          1,
			"run_research":     2,
			"run_brand":        3,
			"run_ux":           4,
			"run_gtm":          5,
			"request_approval": 6,
			"assemble_plan":    7,
		},
	}

	if categoryOrder, ok := order[entry.Category]; ok {
		if rank, ok := categoryOrder[entry.Name]; ok {
			return rank
		}
	}
	return 1_000
}

func wrapJoinedItems(items []string, width int, indent string) string {
	if len(items) == 0 {
		return indent
	}

	var lines []string
	current := indent
	for _, item := range items {
		if current == indent {
			current += item
			continue
		}

		candidate := current + " · " + item
		if runeLen(candidate) > width {
			lines = append(lines, current)
			current = indent + item
			continue
		}
		current = candidate
	}
	lines = append(lines, current)
	return strings.Join(lines, "\n")
}

func titleCaseWord(value string) string {
	if value == "" {
		return value
	}
	runes := []rune(value)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}
