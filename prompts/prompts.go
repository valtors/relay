// Package prompts embeds agent prompt files into the binary.
//
// NOTE: The spec showed `tools/prompts.go` with `//go:embed ../prompts/*.md`,
// but Go's embed directive forbids parent-directory paths. We achieve the same
// goal by placing the embed loader in this package alongside the .md files.
package prompts

import (
	"embed"
	"fmt"
)

//go:embed *.md
var promptFS embed.FS

// Load reads a prompt file from the embedded filesystem.
// Compiled into the binary — no loose files at runtime.
func Load(name string) (string, error) {
	b, err := promptFS.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("load prompt %s: %w", name, err)
	}
	return string(b), nil
}
