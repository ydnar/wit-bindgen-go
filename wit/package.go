package wit

import (
	"strings"

	"go.bytecodealliance.org/wit/ordered"
)

// Package represents a [WIT package] within a [Resolve].
// It implements the [Node] interface.
//
// A Package is a collection of [Interface] and [World] values. Additionally,
// a Package contains a unique identifier that affects generated components and uniquely
// identifies this particular package.
//
// [WIT package]: https://component-model.bytecodealliance.org/design/wit.html#packages
type Package struct {
	Name       Ident
	Interfaces ordered.Map[string, *Interface]
	Worlds     ordered.Map[string, *World]
	Docs       Docs
}

// Clone returns a shallow clone of p.
func (p *Package) Clone() *Package {
	c := *p
	c.Interfaces = *p.Interfaces.Clone()
	c.Worlds = *p.Worlds.Clone()
	return &c
}

func (p *Package) dependsOn(dep Node) bool {
	if dep == p {
		return true
	}
	var done bool
	p.Interfaces.All()(func(_ string, i *Interface) bool {
		done = DependsOn(i, dep)
		return !done
	})
	if done {
		return true
	}
	p.Worlds.All()(func(_ string, w *World) bool {
		done = DependsOn(w, dep)
		return !done
	})
	return done
}

func comparePackages(a, b *Package) int {
	// fmt.Fprintln(os.Stderr, "comparing "+b.Name.String()+" to "+a.Name.String())
	switch {
	case a == b:
		return 0
	case DependsOn(a, b):
		// fmt.Fprintln(os.Stderr, a.Name.String()+" depends on "+b.Name.String())
		return 1
	case DependsOn(b, a):
		// fmt.Fprintln(os.Stderr, b.Name.String()+" depends on "+a.Name.String())
		return -1
	}
	// fmt.Fprintln(os.Stderr, a.Name.String()+" does not depend on "+b.Name.String())
	return -1 * strings.Compare(a.Name.String(), b.Name.String())
}
