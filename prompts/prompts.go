// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Tamish Max
package prompts

import (
	"embed"
	"fmt"
)

//go:embed *.md
var promptFS embed.FS

func Load(name string) (string, error) {
	b, err := promptFS.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("load prompt %s: %w", name, err)
	}
	return string(b), nil
}
