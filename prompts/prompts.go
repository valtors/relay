package prompts

import (
	"embed"
	"fmt"
)

var promptFS embed.FS

func Load(name string) (string, error) {
	b, err := promptFS.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("load prompt %s: %w", name, err)
	}
	return string(b), nil
}
