package wit

import "go.bytecodealliance.org/wit/clone"

// Resource represents a WIT [resource type].
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [resource type]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#item-resource
type Resource struct{ _typeDefKind }

// Size returns the [ABI byte size] for [Resource].
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (*Resource) Size() uintptr { return 4 }

// Align returns the [ABI byte alignment] for [Resource].
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (*Resource) Align() uintptr { return 4 }

// Flat returns the [flattened] ABI representation of [Resource].
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (*Resource) Flat() []Type { return []Type{U32{}} }

// hasResource always returns true.
func (*Resource) hasResource() bool { return true }

// Handle represents a WIT [handle type].
// It conforms to the [Node], [ABI], and [TypeDefKind] interfaces.
// Handles represent the passing of unique ownership of a resource between
// two components. When the owner of an owned handle drops that handle,
// the resource is destroyed. In contrast, a borrowed handle represents
// a temporary loan of a handle from the caller to the callee for the
// duration of the call.
//
// [handle type]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#handles
type Handle interface {
	TypeDefKind
	isHandle()
}

// _handle is an embeddable type that conforms to the [Handle] interface.
type _handle struct{ _typeDefKind }

func (_handle) isHandle() {}

// Size returns the [ABI byte size] for this [Handle].
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (_handle) Size() uintptr { return 4 }

// Align returns the [ABI byte alignment] for this [Handle].
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (_handle) Align() uintptr { return 4 }

// Flat returns the [flattened] ABI representation of this type.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (_handle) Flat() []Type { return []Type{U32{}} }

// Own represents an WIT [owned handle].
// It implements the [Handle], [Node], [ABI], and [TypeDefKind] interfaces.
//
// [owned handle]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#handles
type Own struct {
	_handle
	Type *TypeDef
}

// DeepClone implements [clone.DeepClonable].
func (o *Own) DeepClone(state *clone.State) clone.DeepClonable {
	c := clone.Clone(state, o)
	c.Type = *clone.DeepClone(state, &o.Type)
	return c
}

func (o *Own) hasResource() bool       { return HasResource(o.Type) }
func (o *Own) dependsOn(dep Node) bool { return dep == o || DependsOn(o.Type, dep) }

// Borrow represents a WIT [borrowed handle].
// It implements the [Handle], [Node], [ABI], and [TypeDefKind] interfaces.
//
// [borrowed handle]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#handles
type Borrow struct {
	_handle
	Type *TypeDef
}

// DeepClone implements [clone.DeepClonable].
func (b *Borrow) DeepClone(state *clone.State) clone.DeepClonable {
	c := clone.Clone(state, b)
	c.Type = *clone.DeepClone(state, &b.Type)
	return c
}

func (*Borrow) hasBorrow() bool           { return true }
func (b *Borrow) hasResource() bool       { return HasResource(b.Type) }
func (b *Borrow) dependsOn(dep Node) bool { return dep == b || DependsOn(b.Type, dep) }
