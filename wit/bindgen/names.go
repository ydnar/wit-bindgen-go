package bindgen

import (
	"strings"
	"unicode"

	"go.bytecodealliance.org/internal/go/gen"
)

// GoPackageName generates a Go local package name (e.g. "json").
func GoPackageName(name string) string {
	return strings.Map(func(c rune) rune {
		if notLetterDigit(c) {
			return -1
		}
		return c
	}, strings.ToLower(name))
}

// GoName returns an idiomatic (exported CamelCase) Go name for a WIT name.
func GoName(name string, export bool) string {
	var b strings.Builder
	for i, segment := range segments(name) {
		if i == 0 && !export {
			segment = strings.ToLower(segment)
			if s, ok := Segments[segment]; ok {
				// Use opinionated segment
				b.WriteString(s)
			} else {
				// Default to lowercase segment
				b.WriteString(segment)
			}
		} else {
			if segment == strings.ToUpper(segment) {
				// Preserve all UPPERCASE
				b.WriteString(segment)
			} else if s, ok := ExportedSegments[segment]; ok {
				// Use opinionated segment
				b.WriteString(s)
			} else if gen.Initialisms[segment] {
				// Use opinionated segment from initialisms
				b.WriteString(strings.ToUpper(segment))
			} else {
				// Title-case the segment
				runes := []rune(segment)
				runes[0] = unicode.ToUpper(runes[0])
				b.WriteString(string(runes))
			}
		}
	}
	return b.String()
}

// SnakeName returns a snake_case equivalent of a WIT name.
// It may conflict with a Go keyword or predeclared identifier.
func SnakeName(name string) string {
	return strings.Join(segments(strings.ToLower(name)), "_")
}

// FlatName returns a flat equivalent of a WIT name, where the segments are joined together with no delimiter.
// It may conflict with a Go keyword or predeclared identifier.
func FlatName(name string) string {
	return strings.Join(segments(strings.ToLower(name)), "")
}

// segments splits a kebab-case WIT name into its constituent segments.
// For example: "hello-world" splits into "hello", "world".
func segments(name string) []string {
	return strings.FieldsFunc(name, notLetterDigit)
}

func notLetterDigit(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsDigit(c)
}

// Segments maps common WASI identifier segments to opinionated non-exported Go equivalents.
var Segments = map[string]string{
	"datetime": "dateTime",
	"filesize": "fileSize",
	"readlink": "readLink",
}

// ExportedSegments maps common WASI identifier segments to opinionated exported Go equivalents.
var ExportedSegments = map[string]string{
	"datetime": "DateTime",
	"filesize": "FileSize",
	"ipv4":     "IPv4",
	"ipv6":     "IPv6",
	"readlink": "ReadLink",
}
