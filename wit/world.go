package wit

import (
	"go.bytecodealliance.org/wit/iterate"
	"go.bytecodealliance.org/wit/ordered"
)

// A World represents all of the imports and exports of a [WebAssembly component].
// It implements the [Node] and [TypeOwner] interfaces.
//
// [WebAssembly component]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#wit-worlds
type World struct {
	_typeOwner

	Name      string
	Imports   ordered.Map[string, WorldItem]
	Exports   ordered.Map[string, WorldItem]
	Package   *Package  // the Package this World belongs to (must be non-nil)
	Stability Stability // WIT @since or @unstable (nil if unknown)
	Docs      Docs
}

// Clone returns a shallow clone of w.
func (w *World) Clone() *World {
	c := *w
	c.Imports = *w.Imports.Clone()
	c.Exports = *w.Exports.Clone()
	return &c
}

// WITPackage returns the [Package] that [World] w belongs to.
func (w *World) WITPackage() *Package {
	return w.Package
}

// Match returns true if [World] w matches pattern, which can be one of:
// "name", "namespace:package/name" (qualified), or "namespace:package/name@1.0.0" (versioned).
func (w *World) Match(pattern string) bool {
	if pattern == w.Name {
		return true
	}
	id := w.Package.Name
	id.Extension = w.Name
	if pattern == id.String() {
		return true
	}
	id.Version = nil
	return pattern == id.String()
}

// HasInterface returns true if [World] w references [Interface] i.
func (w *World) HasInterface(i *Interface) bool {
	var found bool
	w.AllInterfaces()(func(_ string, face *Interface) bool {
		found = face == i
		return !found
	})
	return found
}

// AllInterfaces returns a [sequence] that yields each [Interface] in a [World].
// The sequence stops if yield returns false.
//
// [sequence]: https://github.com/golang/go/issues/61897
func (w *World) AllInterfaces() iterate.Seq2[string, *Interface] {
	return func(yield func(string, *Interface) bool) {
		w.AllItems()(func(name string, i WorldItem) bool {
			if ref, ok := i.(*InterfaceRef); ok {
				return yield(name, ref.Interface)
			}
			return true
		})
	}
}

// AllTypeDefs returns a [sequence] that yields each [TypeDef] in a [World].
// The sequence stops if yield returns false.
//
// [sequence]: https://github.com/golang/go/issues/61897
func (w *World) AllTypeDefs() iterate.Seq2[string, *TypeDef] {
	return func(yield func(string, *TypeDef) bool) {
		w.AllItems()(func(name string, i WorldItem) bool {
			if t, ok := i.(*TypeDef); ok {
				return yield(name, t)
			}
			return true
		})
	}
}

// AllFunctions returns a [sequence] that yields each [Function] in a [World].
// The sequence stops if yield returns false.
//
// [sequence]: https://github.com/golang/go/issues/61897
func (w *World) AllFunctions() iterate.Seq[*Function] {
	return func(yield func(*Function) bool) {
		w.AllItems()(func(_ string, i WorldItem) bool {
			if f, ok := i.(*Function); ok {
				return yield(f)
			}
			return true
		})
	}
}

// AllItems returns a [sequence] that yields each [WorldItem] in a [World].
// The sequence stops if yield returns false.
//
// [sequence]: https://github.com/golang/go/issues/61897
func (w *World) AllItems() iterate.Seq2[string, WorldItem] {
	return func(yield func(string, WorldItem) bool) {
		var done bool
		yield = iterate.Done2(iterate.Once2(yield), func() { done = true })
		f := func(name string, i WorldItem) bool {
			return yield(name, i)
		}
		w.Imports.All()(f)
		if done {
			return
		}
		w.Exports.All()(f)
	}
}

func (w *World) dependsOn(dep Node) bool {
	if dep == w || dep == w.Package {
		return true
	}
	var done bool
	w.AllItems()(func(_ string, i WorldItem) bool {
		done = DependsOn(i, dep)
		return !done
	})
	return done
}

// A WorldItem is any item that can be exported from or imported into a [World],
// currently either an [InterfaceRef], [TypeDef], or [Function].
// Any WorldItem is also a [Node].
type WorldItem interface {
	Node
	isWorldItem()
}

// _worldItem is an embeddable type that conforms to the [WorldItem] interface.
type _worldItem struct{}

func (_worldItem) isWorldItem() {}
