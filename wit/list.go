package wit

// List represents a WIT [list type], which is an ordered vector of an arbitrary type.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [list type]: https://component-model.bytecodealliance.org/design/wit.html#lists
type List struct {
	_typeDefKind
	Type Type
}

// Size returns the [ABI byte size] for a [List].
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (*List) Size() uintptr { return 8 } // [2]int32

// Align returns the [ABI byte alignment] a [List].
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (*List) Align() uintptr { return 4 }

// Flat returns the [flattened] ABI representation of [List].
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (l *List) Flat() []Type { return []Type{PointerTo(l.Type), U32{}} }

func (*List) hasPointer() bool          { return true }
func (l *List) hasBorrow() bool         { return HasBorrow(l.Type) }
func (l *List) hasResource() bool       { return HasResource(l.Type) }
func (l *List) dependsOn(dep Node) bool { return dep == l || DependsOn(l.Type, dep) }
