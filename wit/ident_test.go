package wit

import (
	"reflect"
	"testing"

	"github.com/coreos/go-semver/semver"
)

func TestIdent(t *testing.T) {
	tests := []struct {
		s       string
		want    Ident
		wantErr bool
	}{
		{"wasi:io", Ident{Namespace: "wasi", Package: "io"}, false},
		{"wasi:io@0.2.0", Ident{Namespace: "wasi", Package: "io", Version: semver.New("0.2.0")}, false},
		{"wasi:io/streams", Ident{Namespace: "wasi", Package: "io", Extension: "streams"}, false},
		{"wasi:io/streams@0.2.0", Ident{Namespace: "wasi", Package: "io", Extension: "streams", Version: semver.New("0.2.0")}, false},

		// Escaping
		{"%use:%own", Ident{Namespace: "use", Package: "own"}, false},
		{"%use:%own@0.2.0", Ident{Namespace: "use", Package: "own", Version: semver.New("0.2.0")}, false},
		{"%use:%own/%type", Ident{Namespace: "use", Package: "own", Extension: "type"}, false},
		{"%use:%own/%type@0.2.0", Ident{Namespace: "use", Package: "own", Extension: "type", Version: semver.New("0.2.0")}, false},

		// Mixed-case
		{"ABC:def-GHI", Ident{Namespace: "ABC", Package: "def-GHI"}, false},
		{"ABC1:def2-GHI3", Ident{Namespace: "ABC1", Package: "def2-GHI3"}, false},

		// Errors
		{"", Ident{}, true},
		{":", Ident{}, true},
		{":/", Ident{}, true},
		{":/@", Ident{}, true},
		{"wasi", Ident{Namespace: "wasi"}, true},
		{"wasi:", Ident{Namespace: "wasi"}, true},
		{"wasi:/", Ident{}, true},
		{"wasi:clocks@", Ident{}, true},
		{"wasi:clocks/wall-clock@", Ident{}, true},
		{"foo%:bar%baz", Ident{Namespace: "foo%", Package: "bar%baz"}, true},
		{"-foo:bar", Ident{Namespace: "-foo", Package: "bar"}, true},
		{"foo-:bar", Ident{Namespace: "foo-", Package: "bar"}, true},
		{"foo--foo:bar", Ident{Namespace: "foo--foo", Package: "bar"}, true},
		{"aBc:bar", Ident{Namespace: "aBc", Package: "bar"}, true},
		{"1:2", Ident{Namespace: "1", Package: "2"}, true},
		{"1a:2b", Ident{Namespace: "1a", Package: "2b"}, true},
		{"foo-1:bar", Ident{Namespace: "foo-1", Package: "bar"}, true},
		{"foo:bar-1", Ident{Namespace: "foo", Package: "bar-2"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got, err := ParseIdent(tt.s)
			if tt.wantErr && err == nil {
				t.Errorf("ParseIdent(%q): expected error, got nil error", tt.s)
			} else if !tt.wantErr && err != nil {
				t.Errorf("ParseIdent(%q): expected no error, got error: %v", tt.s, err)
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseIdent(%q): %v, expected %v", tt.s, got, tt.want)
			}
		})
	}
}
