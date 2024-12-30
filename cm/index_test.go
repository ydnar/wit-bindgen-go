package cm

import (
	"strings"
	"testing"
)

func TestIndex(t *testing.T) {
	tests := []struct {
		name    string
		strings []string
	}{
		{"nil", nil},
		{"empty slice", []string{}},
		{"a b c", strings.SplitAfter("abc", "")},
		{"a b c d e f g", strings.SplitAfter("abcdefg", "")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Index(tt.strings)
			for want, s := range tt.strings {
				got := f(s)
				if got != want {
					t.Errorf("f(%q): got %d, expected %d", s, got, want)
				}
			}
		})
	}
}
