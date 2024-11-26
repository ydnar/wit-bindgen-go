package wit

import "go.bytecodealliance.org/wit/clone"

// PointerTo returns a [Pointer] to [Type] t.
func PointerTo(t Type) *TypeDef {
	return &TypeDef{Kind: &Pointer{Type: t}}
}

// Pointer represents a pointer to a WIT type.
// It is only used for ABI representation, e.g. pointers to function parameters or return values.
type Pointer struct {
	_typeDefKind
	Type Type
}

// Clone implements [clone.Clonable].
func (p *Pointer) Clone(state *clone.State) clone.Clonable {
	c := clone.Shallow(state, p)
	c.Type = *clone.Clone(state, &p.Type)
	return c
}

// Size returns the [ABI byte size] for [Pointer].
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (*Pointer) Size() uintptr { return 4 }

// Align returns the [ABI byte alignment] for [Pointer].
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (*Pointer) Align() uintptr { return 4 }

// Flat returns the [flattened] ABI representation of [Pointer].
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (p *Pointer) Flat() []Type { return []Type{PointerTo(p.Type)} }

func (*Pointer) hasPointer() bool          { return true }
func (p *Pointer) hasBorrow() bool         { return HasBorrow(p.Type) }
func (p *Pointer) hasResource() bool       { return HasResource(p.Type) }
func (p *Pointer) dependsOn(dep Node) bool { return dep == p || DependsOn(p.Type, dep) }
