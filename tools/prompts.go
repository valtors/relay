// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Tamish Max
package tools

import (
	"github.com/valtors/relay/prompts"
)

func loadPrompt(name string) (string, error) {
	return prompts.Load(name)
}
