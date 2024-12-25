package wit

// ErrorContext represents a WIT [error-context] type.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [resource type]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/Explainer.md#error-context-type
type ErrorContext struct{ _typeDefKind }

// Size returns the [ABI byte size] for [ErrorContext].
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (*ErrorContext) Size() uintptr { return 4 }

// Align returns the [ABI byte alignment] for [ErrorContext].
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (*ErrorContext) Align() uintptr { return 4 }

// Flat returns the [flattened] ABI representation of [ErrorContext].
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (*ErrorContext) Flat() []Type { return []Type{U32{}} }

// hasResource always returns false.
// TODO: is error-context a resource? Does it have direction?
func (*ErrorContext) hasResource() bool { return false }
