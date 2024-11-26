package wit

import "go.bytecodealliance.org/wit/clone"

// Stream represents a WIT [stream type], expected to be part of [WASI Preview 3].
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [stream type]: https://github.com/WebAssembly/WASI/blob/main/docs/WitInWasi.md#streams
// [WASI Preview 3]: https://bytecodealliance.org/articles/webassembly-the-updated-roadmap-for-developers
type Stream struct {
	_typeDefKind
	Element Type // optional associated Type (can be nil)
	End     Type // optional associated Type (can be nil)
}

// DeepClone implements [clone.DeepClonable].
func (s *Stream) DeepClone(state *clone.State) clone.DeepClonable {
	c := clone.Clone(state, s)
	c.Element = *clone.DeepClone(state, &s.Element)
	c.End = *clone.DeepClone(state, &s.End)
	return c
}

// Size returns the [ABI byte size] for a [Stream].
// TODO: what is the ABI size of a stream?
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (*Stream) Size() uintptr { return 0 }

// Align returns the [ABI byte alignment] a [Stream].
// TODO: what is the ABI alignment of a stream?
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (*Stream) Align() uintptr { return 0 }

// Flat returns the [flattened] ABI representation of [Stream].
// TODO: what is the ABI representation of a stream?
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (*Stream) Flat() []Type { return nil }

func (s *Stream) hasPointer() bool  { return HasPointer(s.Element) || HasPointer(s.End) }
func (s *Stream) hasBorrow() bool   { return HasBorrow(s.Element) || HasBorrow(s.End) }
func (s *Stream) hasResource() bool { return HasResource(s.Element) || HasResource(s.End) }
func (s *Stream) dependsOn(dep Node) bool {
	return dep == s || DependsOn(s.Element, dep) || DependsOn(s.End, dep)
}
