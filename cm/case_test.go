package cm

import (
	"strings"
	"testing"
)

func TestCaseUnmarshaler(t *testing.T) {
	tests := []struct {
		name  string
		cases []string
	}{
		{"nil", nil},
		{"empty slice", []string{}},
		{"a b c", strings.SplitAfter("abc", "")},
		{"a b c d e f g", strings.SplitAfter("abcdefg", "")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := CaseUnmarshaler[uint8](tt.cases)
			for want, c := range tt.cases {
				var got uint8
				err := f(&got, []byte(c))
				if err != nil {
					t.Error(err)
					return
				}
				if got != uint8(want) {
					t.Errorf("f(%q): got %d, expected %d", c, got, want)
				}
			}
		})
	}
}
