package wit

import (
	"slices"

	"go.bytecodealliance.org/wit/clone"
)

// PruneToWorld returns a copy of [Resolve] r containing only nodes referenced by [World] w.
func PruneToWorld(r *Resolve, w *World) *Resolve {
	return pruneToNode(r, w)
}

// PruneToInterface returns a copy of [Resolve] r containing only nodes referenced by [Interface] i.
// The resulting Resolve may include unique [World] node(s) that are a subset of the originals in r.
func PruneToInterface(r *Resolve, i *Interface) *Resolve {
	return pruneToNode(r, i)
}

func pruneToNode(r *Resolve, target Node) *Resolve {
	state := &clone.State{}
	return pruneResolve(r, state, func(node Node) bool {
		switch node := node.(type) {
		case *Package:
			return !DependsOn(node, target) && !DependsOn(target, node)
		case *World:
			return !DependsOn(node, target)
		case *Interface:
			return !DependsOn(target, node)
		case *Function:
			return !DependsOn(target, node)
		}
		return false // keep
	})
}

func pruneResolve(r *Resolve, state *clone.State, del func(Node) bool) *Resolve {
	// del = memoize.Function(del)
	r = clone.Clone(state, r)
	r.Packages = slices.DeleteFunc(r.Packages, func(p *Package) bool { return del(p) })
	for i, p := range r.Packages {
		r.Packages[i] = prunePackage(p, state, del)
	}
	r.Worlds = slices.DeleteFunc(r.Worlds, func(w *World) bool { return del(w) })
	for i, w := range r.Worlds {
		r.Worlds[i] = pruneWorld(w, state, del)
	}
	r.Interfaces = slices.DeleteFunc(r.Interfaces, func(i *Interface) bool { return del(i) })
	r.TypeDefs = slices.DeleteFunc(r.TypeDefs, func(t *TypeDef) bool { return del(t) })
	return r
}

func prunePackage(p *Package, state *clone.State, del func(Node) bool) *Package {
	p.Worlds.All()(func(name string, w *World) bool {
		if del(w) {
			p = clone.Clone(state, p) // lazily clone
			p.Worlds.Delete(name)
			return true
		}
		w2 := pruneWorld(w, state, del)
		if w2 != w {
			p = clone.Clone(state, p) // lazily clone
			p.Worlds.Set(name, w2)
		}
		return true
	})
	p.Interfaces.All()(func(name string, i *Interface) bool {
		if del(i) {
			p = clone.Clone(state, p) // lazily clone
			p.Interfaces.Delete(name)
		}
		return true
	})
	return p
}

func pruneWorld(w *World, state *clone.State, del func(Node) bool) *World {
	w.Imports.All()(func(name string, i WorldItem) bool {
		var node Node = i
		if ref, ok := i.(*InterfaceRef); ok {
			node = ref.Interface
		}
		if del(node) {
			w = clone.Clone(state, w) // lazily clone
			w.Imports.Delete(name)
		}
		return true
	})
	w.Exports.All()(func(name string, i WorldItem) bool {
		var node Node = i
		if ref, ok := i.(*InterfaceRef); ok {
			node = ref.Interface
		}
		if del(node) {
			w = clone.Clone(state, w) // lazily clone
			w.Exports.Delete(name)
		}
		return true
	})
	return w
}
