package wit

// Stream represents a WIT [stream type], expected to be part of [WASI Preview 3].
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [stream type]: https://github.com/WebAssembly/WASI/blob/main/docs/WitInWasi.md#streams
// [WASI Preview 3]: https://bytecodealliance.org/articles/webassembly-the-updated-roadmap-for-developers
type Stream struct {
	_typeDefKind
	Type Type // associated Type (must not be nil)
}

// Size returns the [ABI byte size] for a [Stream].
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (*Stream) Size() uintptr { return 4 }

// Align returns the [ABI byte alignment] a [Stream].
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (*Stream) Align() uintptr { return 4 }

// Flat returns the [flattened] ABI representation of [Stream].
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (*Stream) Flat() []Type { return []Type{U32{}} }

func (s *Stream) hasPointer() bool  { return HasPointer(s.Type) }
func (s *Stream) hasBorrow() bool   { return HasBorrow(s.Type) }
func (s *Stream) hasResource() bool { return HasResource(s.Type) }
func (s *Stream) dependsOn(dep Node) bool {
	return dep == s || DependsOn(s.Type, dep)
}
