package wit

import (
	"slices"

	"go.bytecodealliance.org/wit/iterate"
)

// Resolve represents a fully resolved set of WIT ([WebAssembly Interface Type])
// packages and worlds. It implements the [Node] interface.
//
// This structure contains a graph of WIT packages and their contents
// merged together into slices organized by type. Items are sorted
// topologically and everything is fully resolved.
//
// Each [World], [Interface], [TypeDef], or [Package] in a Resolve must be non-nil.
//
// [WebAssembly Interface Type]: https://component-model.bytecodealliance.org/design/wit.html
type Resolve struct {
	Worlds     []*World
	Interfaces []*Interface
	TypeDefs   []*TypeDef
	Packages   []*Package
}

// Clone returns a shallow clone of r.
func (r *Resolve) Clone() *Resolve {
	c := *r
	c.Worlds = slices.Clone(r.Worlds)
	c.Interfaces = slices.Clone(r.Interfaces)
	c.TypeDefs = slices.Clone(r.TypeDefs)
	c.Packages = slices.Clone(r.Packages)
	return &c
}

// AllFunctions returns a [sequence] that yields each [Function] in a [Resolve].
// The sequence stops if yield returns false.
//
// [sequence]: https://github.com/golang/go/issues/61897
func (r *Resolve) AllFunctions() iterate.Seq[*Function] {
	return func(yield func(*Function) bool) {
		var done bool
		yield = iterate.Done(iterate.Once(yield), func() { done = true })
		for i := 0; i < len(r.Worlds) && !done; i++ {
			r.Worlds[i].AllFunctions()(yield)
		}
		for i := 0; i < len(r.Interfaces) && !done; i++ {
			r.Interfaces[i].AllFunctions()(yield)
		}
	}
}

func (r *Resolve) dependsOn(dep Node) bool {
	for _, w := range r.Worlds {
		if DependsOn(w, dep) {
			return true
		}
	}
	for _, i := range r.Interfaces {
		if DependsOn(i, dep) {
			return true
		}
	}
	for _, t := range r.TypeDefs {
		if DependsOn(t, dep) {
			return true
		}
	}
	for _, p := range r.Packages {
		if DependsOn(p, dep) {
			return true
		}
	}
	return false
}
