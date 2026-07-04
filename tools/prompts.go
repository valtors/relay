package tools

import (
	"relay/prompts"
)

func loadPrompt(name string) (string, error) {
	return prompts.Load(name)
}
