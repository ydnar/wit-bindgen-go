package bindgen

import (
	"strings"

	"go.bytecodealliance.org/internal/go/gen"
)

func formatDocComments(s string, indent bool) string {
	return gen.FormatDocComments(processMarkdown(s), indent)
}

func processMarkdown(s string) string {
	var lines []string
	var indent bool
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(line, "```") {
			indent = !indent
			line = ""
		} else if indent {
			line = "\t" + line
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
