package tools

import (
	"github.com/valtors/relay/prompts"
)

func loadPrompt(name string) (string, error) {
	return prompts.Load(name)
}
