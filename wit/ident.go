package wit

import (
	"errors"
	"strconv"
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
	var id Ident
	name, ver, hasVer := strings.Cut(s, "@")
	if hasVer {
		var err error
		id.Version, err = semver.NewVersion(ver)
		if err != nil {
			return id, err
		}
	}
	base, ext, hasExt := strings.Cut(name, "/")
	ns, pkg, _ := strings.Cut(base, ":")
	id.Namespace = trimPercent(ns)
	id.Package = trimPercent(pkg)
	if hasExt {
		id.Extension = trimPercent(ext)
	}
	return id, id.Validate()
}

func trimPercent(s string) string {
	if len(s) > 0 && s[0] == '%' {
		return s[1:]
	}
	return s
}

// Validate validates id, returning any errors.
func (id *Ident) Validate() error {
	switch {
	case id.Namespace == "":
		return errors.New("missing package namespace")
	case id.Package == "":
		return errors.New("missing package name")
	}
	if err := validateName(id.Namespace); err != nil {
		return err
	}
	if err := validateName(id.Package); err != nil {
		return err
	}
	return validateName(id.Extension)
}

func validateName(s string) error {
	if len(s) == 0 {
		return nil
	}
	var prev rune
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z':
			switch {
			case prev >= 'A' && prev <= 'Z':
				return errors.New("invalid character " + strconv.Quote(string(c)))
			}
		case c >= 'A' && c <= 'Z':
			switch {
			case prev == 0: // start of string
			case prev >= 'A' && prev <= 'Z':
			case prev >= '0' && prev <= '9':
			case prev == '-':
			default:
				return errors.New("invalid character " + strconv.Quote(string(c)))
			}
		case c >= '0' && c <= '9':
			switch {
			case prev >= 'a' && prev <= 'z':
			case prev >= 'A' && prev <= 'Z':
			case prev >= '0' && prev <= '9':
			default:
				return errors.New("invalid character " + strconv.Quote(string(c)))
			}
		case c == '-':
			switch {
			case prev == 0: // start of string
				return errors.New("invalid leading -")
			case prev == '-':
				return errors.New("invalid double --")
			}
		default:
			return errors.New("invalid character " + strconv.Quote(string(c)))
		}
		prev = c
	}
	if prev == '-' {
		return errors.New("invalid trailing -")
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
