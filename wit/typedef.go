package wit

import (
	"slices"

	"go.bytecodealliance.org/wit/clone"
)

// TypeDef represents a WIT type definition. A TypeDef may be named or anonymous,
// and optionally belong to a [World] or [Interface].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
type TypeDef struct {
	_type
	_worldItem
	Name      *string
	Kind      TypeDefKind
	Owner     TypeOwner
	Stability Stability // WIT @since or @unstable (nil if unknown)
	Docs      Docs
}

// DeepClone implements [clone.Clonable].
func (t *TypeDef) DeepClone(state *clone.State) clone.Clonable {
	c := clone.Shallow(state, t)
	c.Kind = *clone.Clone(state, &t.Kind)
	c.Owner = *clone.Clone(state, &t.Owner)
	c.Stability = *clone.Clone(state, &t.Stability)
	return c
}

// TypeName returns the [WIT] type name for t.
// Returns an empty string if t is anonymous.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (t *TypeDef) TypeName() string {
	if t.Name != nil {
		return *t.Name
	}
	return ""
}

// TypeDef returns the parent [TypeDef] of [TypeDef] t.
// If t is not a [type alias], TypeDef returns t.
//
// [type alias]: https://component-model.bytecodealliance.org/design/wit.html#type-aliases
func (t *TypeDef) TypeDef() *TypeDef {
	if t, ok := t.Kind.(*TypeDef); ok {
		return t
	}
	return t
}

// Root returns the root [TypeDef] of [TypeDef] t.
// If t is not a [type alias], Root returns t.
//
// [type alias]: https://component-model.bytecodealliance.org/design/wit.html#type-aliases
func (t *TypeDef) Root() *TypeDef {
	for {
		switch kind := t.Kind.(type) {
		case *TypeDef:
			t = kind
		default:
			return t
		}
	}
}

// Constructor returns the constructor for [TypeDef] t, or nil if none.
// Currently t must be a [Resource] to have a constructor.
func (t *TypeDef) Constructor() *Function {
	var constructor *Function
	t.Owner.AllFunctions()(func(f *Function) bool {
		if c, ok := f.Kind.(*Constructor); ok && c.Type == t {
			constructor = f
			return false
		}
		return true
	})
	return constructor
}

// StaticFunctions returns all static functions for [TypeDef] t.
// Currently t must be a [Resource] to have static functions.
func (t *TypeDef) StaticFunctions() []*Function {
	var statics []*Function
	t.Owner.AllFunctions()(func(f *Function) bool {
		if s, ok := f.Kind.(*Static); ok && s.Type == t {
			statics = append(statics, f)
		}
		return true
	})
	slices.SortFunc(statics, compareFunctions)
	return statics
}

// Methods returns all methods for [TypeDef] t.
// Currently t must be a [Resource] to have methods.
func (t *TypeDef) Methods() []*Function {
	var methods []*Function
	t.Owner.AllFunctions()(func(f *Function) bool {
		if m, ok := f.Kind.(*Method); ok && m.Type == t {
			methods = append(methods, f)
		}
		return true
	})
	slices.SortFunc(methods, compareFunctions)
	return methods
}

// Size returns the byte size for values of type t.
func (t *TypeDef) Size() uintptr {
	return t.Kind.Size()
}

// Align returns the byte alignment for values of type t.
func (t *TypeDef) Align() uintptr {
	return t.Kind.Align()
}

// Flat returns the [flattened] ABI representation of t.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (t *TypeDef) Flat() []Type {
	return t.Kind.Flat()
}

func (t *TypeDef) hasPointer() bool  { return HasPointer(t.Kind) }
func (t *TypeDef) hasBorrow() bool   { return HasBorrow(t.Kind) }
func (t *TypeDef) hasResource() bool { return HasResource(t.Kind) }
func (t *TypeDef) dependsOn(dep Node) bool {
	return dep == t || dep == t.Owner ||
		(t.Owner != nil && dep == t.Owner.WITPackage()) ||
		DependsOn(t.Kind, dep)
}

// TypeDefKind represents the underlying type in a [TypeDef], which can be one of
// [Record], [Resource], [Handle], [Flags], [Tuple], [Variant], [Enum],
// [Option], [Result], [List], [Future], [Stream], or [Type].
// It implements the [Node] and [ABI] interfaces.
type TypeDefKind interface {
	Node
	ABI
	isTypeDefKind()
}

// _typeDefKind is an embeddable type that conforms to the [TypeDefKind] interface.
type _typeDefKind struct{}

func (_typeDefKind) isTypeDefKind() {}

// KindOf probes [Type] t to determine if it is a [TypeDef] with [TypeDefKind] K.
// It returns the underlying Kind if present.
func KindOf[K TypeDefKind](t Type) (kind K) {
	if td, ok := t.(*TypeDef); ok {
		if kind, ok = td.Kind.(K); ok {
			return kind
		}
	}
	var zero K
	return zero
}
