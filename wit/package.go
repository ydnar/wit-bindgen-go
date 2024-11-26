package wit

import (
	"strings"

	"go.bytecodealliance.org/wit/clone"
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

// DeepClone implements [clone.Clonable].
func (p *Package) DeepClone(state *clone.State) clone.Clonable {
	c := clone.Shallow(state, p)
	c.Interfaces = *clone.Clone(state, &p.Interfaces)
	c.Worlds = *clone.Clone(state, &p.Worlds)
	return c
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

// constrainTo destructively constrains p to node.
func (p *Package) constrainTo(node Node) {
	p.Worlds.All()(func(name string, w *World) bool {
		if !DependsOn(w, node) {
			p.Worlds.Delete(name)
			return true
		}
		w.constrainTo(node)
		return true
	})
	p.Interfaces.All()(func(name string, i *Interface) bool {
		if !DependsOn(node, i) {
			p.Interfaces.Delete(name)
		}
		return true
	})
}

func comparePackages(a, b *Package) int {
	switch {
	case a == b:
		return 0
	case DependsOn(b, a):
		// println(b.Name.String() + " depends on " + a.Name.String())
		return 1
	case DependsOn(a, b):
		// println(a.Name.String() + " depends on " + b.Name.String())
		return -1
	}
	// println(a.Name.String() + " does not depend on " + b.Name.String())
	return strings.Compare(a.Name.String(), b.Name.String())
}
