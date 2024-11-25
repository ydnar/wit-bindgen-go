package wit

import (
	"go.bytecodealliance.org/wit/clone"
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

// Clone implements [clone.Clonable].
// The resulting [Resolve] and its contents may be freely modified.
func (r *Resolve) Clone(state *clone.State) clone.Clonable {
	c := clone.Shallow(state, r)
	c.Worlds = clone.Slice(state, r.Worlds)
	c.Interfaces = clone.Slice(state, r.Interfaces)
	c.TypeDefs = clone.Slice(state, r.TypeDefs)
	c.Packages = clone.Slice(state, r.Packages)
	return c
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
