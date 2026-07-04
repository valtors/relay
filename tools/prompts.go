// Package tools contains MCP tool handlers.
//
// During Phase 0 each handler is a stub that returns its target phase.
// Stubs exist so all 8 tools register and appear in MCP clients before
// implementations are written.
package tools

import (
	"relay/prompts"
)

// loadPrompt is a thin wrapper around the embedded prompts package so all
// tool handlers can share a single load call site.
func loadPrompt(name string) (string, error) {
	return prompts.Load(name)
}
