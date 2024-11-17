package wit

// Future represents a WIT [future type], expected to be part of [WASI Preview 3].
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [future type]: https://github.com/bytecodealliance/wit-bindgen/issues/270
// [WASI Preview 3]: https://bytecodealliance.org/articles/webassembly-the-updated-roadmap-for-developers
type Future struct {
	_typeDefKind
	Type Type // optional associated Type (can be nil)
}

// Size returns the [ABI byte size] for a [Future].
// TODO: what is the ABI size of a future?
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (*Future) Size() uintptr { return 0 }

// Align returns the [ABI byte alignment] a [Future].
// TODO: what is the ABI alignment of a future?
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (*Future) Align() uintptr { return 0 }

// Flat returns the [flattened] ABI representation of [Future].
// TODO: what is the ABI representation of a stream?
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (*Future) Flat() []Type { return nil }

func (f *Future) hasPointer() bool        { return HasPointer(f.Type) }
func (f *Future) hasBorrow() bool         { return HasBorrow(f.Type) }
func (f *Future) hasResource() bool       { return HasResource(f.Type) }
func (f *Future) dependsOn(dep Node) bool { return dep == f || DependsOn(f.Type, dep) }
