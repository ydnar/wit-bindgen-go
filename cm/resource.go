package cm

import "unsafe"

// Resource represents an opaque Component Model [resource handle].
// It is represented in the [Canonical ABI] as an 32-bit integer.
//
// [resource handle]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/Explainer.md#handle-types
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
type Resource uint32

// ResourceNone is a sentinel value indicating a null or uninitialized resource.
// This is a reserved value specified in the [Canonical ABI runtime state].
//
// [Canonical ABI runtime state]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#runtime-state
const ResourceNone = 0

// Rep represents an opaque resource representation, typically a pointer.
type Rep uint32

func AsRep[T any, Rep RepTypes[T]](rep Rep) T {
	return *(*T)(unsafe.Pointer(&rep))
}

// RepTypes is a type constraint for a concrete resource representation,
// currently represented in the [Canonical ABI] as a 32-bit integer value.
//
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
type RepTypes[T any] interface {
	~int32 | ~uint32 | ~uintptr | *T
}

type Own[T Resourcer] struct {
	handle uint32
}

var _ [unsafe.Sizeof(Own[Resourcer]{})]byte = [unsafe.Sizeof(uint32(0))]byte{}

type Resourcer interface {
	ResourceRep() Rep
	ResourceDestructor()
}
