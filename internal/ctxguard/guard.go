package ctxguard

import "fmt"

const (
	maxChars     = 120_000
	summaryChars = 3_000
)

func Guard(content, label string) string {
	if len(content) <= maxChars {
		return content
	}
	return content[:summaryChars] + fmt.Sprintf(
		"\n\n[TRUNCATED: %s was %d chars. Showing first %d chars to stay within context window.]\n\n",
		label, len(content), summaryChars,
	)
}

type Part struct {
	Label    string
	Content  string
	Required bool
}

func Build(parts []Part) string {
	format := func(p Part) string {
		return "## " + p.Label + "\n" + Guard(p.Content, p.Label)
	}

	var required, optional string
	for _, p := range parts {
		if p.Required {
			if required != "" {
				required += "\n\n---\n\n"
			}
			required += format(p)
		} else {
			if optional != "" {
				optional += "\n\n---\n\n"
			}
			optional += format(p)
		}
	}

	full := required + "\n\n---\n\n" + optional
	if len(full) <= maxChars {
		return full
	}
	return required
}
