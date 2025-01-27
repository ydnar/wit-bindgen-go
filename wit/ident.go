package wit

import (
	"errors"
	"strings"

	"github.com/coreos/go-semver/semver"
)

// Ident represents a [Component Model] identifier for a [Package], [World], or [Interface],
// such as [wasi:clocks@0.2.0] or [wasi:clocks/wall-clock@0.2.0].
//
// A Ident contains a namespace and package name, along with an optional extension and [SemVer] version.
//
// [Component Model]: https://component-model.bytecodealliance.org/introduction.html
// [wasi:clocks@0.2.0]: https://github.com/WebAssembly/wasi-clocks
// [wasi:clocks/wall-clock@0.2.0]: https://github.com/WebAssembly/wasi-clocks/blob/main/wit/wall-clock.wit
// [SemVer]: https://semver.org/
type Ident struct {
	// Namespace specifies the package namespace, such as "wasi" in "wasi:foo/bar".
	Namespace string

	// Package specifies the name of the package.
	Package string

	// Extension optionally specifies a world or interface name.
	Extension string

	// Version optionally specifies version information.
	Version *semver.Version
}

// ParseIdent parses a WIT identifier string into an [Ident],
// returning any errors encountered. The resulting Ident
// may not be valid.
func ParseIdent(s string) (Ident, error) {
	s = strings.ReplaceAll(s, "%", "")
	var id Ident
	name, ver, hasVer := strings.Cut(s, "@")
	base, ext, hasExt := strings.Cut(name, "/")
	id.Namespace, id.Package, _ = strings.Cut(base, ":")
	if hasVer {
		var err error
		id.Version, err = semver.NewVersion(ver)
		if err != nil {
			return id, err
		}
	}
	if hasExt {
		id.Extension = ext
	}
	return id, id.Validate()
}

// Validate validates id, returning any errors.
func (id *Ident) Validate() error {
	switch {
	case id.Namespace == "":
		return errors.New("missing package namespace")
	case id.Package == "":
		return errors.New("missing package name")
	}
	return nil
}

// String implements [fmt.Stringer], returning the canonical string representation of an [Ident].
func (id *Ident) String() string {
	if id.Version == nil {
		return id.UnversionedString()
	}
	if id.Extension == "" {
		return escape(id.Namespace) + ":" + escape(id.Package) + "@" + id.Version.String()
	}
	return escape(id.Namespace) + ":" + escape(id.Package) + "/" + escape(id.Extension) + "@" + id.Version.String()
}

// UnversionedString returns a string representation of an [Ident] without version information.
func (id *Ident) UnversionedString() string {
	if id.Extension == "" {
		return escape(id.Namespace) + ":" + escape(id.Package)
	}
	return escape(id.Namespace) + ":" + escape(id.Package) + "/" + escape(id.Extension)
}
