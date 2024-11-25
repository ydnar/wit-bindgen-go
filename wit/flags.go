package wit

import "go.bytecodealliance.org/wit/clone"

// Flags represents a WIT [flags type], stored as a bitfield.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [flags type]: https://component-model.bytecodealliance.org/design/wit.html#flags
type Flags struct {
	_typeDefKind
	Flags []Flag
}

// Clone implements [clone.Clonable].
func (f *Flags) Clone(state *clone.State) clone.Clonable {
	c := clone.Shallow(state, f)
	c.Flags = clone.Slice(state, f.Flags)
	return c
}

// Size returns the [ABI byte size] of [Flags] f.
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (f *Flags) Size() uintptr {
	n := len(f.Flags)
	switch {
	case n <= 8:
		return 1
	case n <= 16:
		return 2
	}
	return 4 * uintptr((n+31)>>5)
}

// Align returns the [ABI byte alignment] of [Flags] f.
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (f *Flags) Align() uintptr {
	n := len(f.Flags)
	switch {
	case n <= 8:
		return 1
	case n <= 16:
		return 2
	}
	return 4
}

// Flat returns the [flattened] ABI representation of [Flags] f.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (f *Flags) Flat() []Type {
	flat := make([]Type, (len(f.Flags)+31)>>5)
	for i := range flat {
		flat[i] = U32{}
	}
	return flat
}

// Flag represents a single flag value in a [Flags] type.
// It implements the [Node] interface.
type Flag struct {
	Name string
	Docs Docs
}
