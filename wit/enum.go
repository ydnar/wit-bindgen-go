package wit

import "go.bytecodealliance.org/wit/clone"

// Enum represents a WIT [enum type], which is a [Variant] without associated data.
// The equivalent in Go is a set of const identifiers declared with iota.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [enum type]: https://component-model.bytecodealliance.org/design/wit.html#enums
type Enum struct {
	_typeDefKind
	Cases []EnumCase
}

// Clone implements [clone.Clonable].
func (e *Enum) Clone(state *clone.State) clone.Clonable {
	c := clone.Shallow(state, e)
	c.Cases = clone.Slice(state, e.Cases)
	return c
}

// Despecialize despecializes [Enum] e into a [Variant] with no associated types.
// See the [canonical ABI documentation] for more information.
//
// [canonical ABI documentation]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (e *Enum) Despecialize() TypeDefKind {
	v := &Variant{
		Cases: make([]Case, len(e.Cases)),
	}
	for i := range e.Cases {
		v.Cases[i].Name = e.Cases[i].Name
		v.Cases[i].Docs = e.Cases[i].Docs
	}
	return v
}

// Size returns the [ABI byte size] for [Enum] e, the smallest integer
// type that can represent 0...len(e.Cases).
// It is first [despecialized] into a [Variant] with no associated types, then sized.
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
// [despecialized]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (e *Enum) Size() uintptr {
	return e.Despecialize().Size()
}

// Align returns the [ABI byte alignment] for [Enum] e.
// It is first [despecialized] into a [Variant] with no associated types, then aligned.
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
// [despecialized]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (e *Enum) Align() uintptr {
	return e.Despecialize().Align()
}

// Flat returns the [flattened] ABI representation of [Enum] e.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (v *Enum) Flat() []Type {
	return Discriminant(len(v.Cases)).Flat()
}

// EnumCase represents a single case in an [Enum].
// It implements the [Node] interface.
type EnumCase struct {
	Name string
	Docs Docs
}
