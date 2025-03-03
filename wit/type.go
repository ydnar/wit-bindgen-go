package wit

import (
	"fmt"
	"unsafe"

	"go.bytecodealliance.org/wit/iterate"
)

// TypeOwner is the interface implemented by any type that can own a TypeDef,
// currently [World] and [Interface].
type TypeOwner interface {
	Node
	AllFunctions() iterate.Seq[*Function]
	WITPackage() *Package
	isTypeOwner()
}

type _typeOwner struct{}

func (_typeOwner) isTypeOwner() {}

// Type is the interface implemented by any type definition. This can be a
// [primitive type] or a user-defined type in a [TypeDef].
// It also conforms to the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
type Type interface {
	TypeDefKind
	TypeName() string
	isType()
}

// _type is an embeddable struct that conforms to the [Type] interface.
// It also implements the [Node], [ABI], and [TypeDefKind] interfaces.
type _type struct{ _typeDefKind }

func (_type) TypeName() string { return "" }
func (_type) isType()          {}

// ParseType parses a WIT [primitive type] string into
// the associated Type implementation from this package.
// It returns an error if the type string is not recognized.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
func ParseType(s string) (Type, error) {
	switch s {
	case "bool":
		return Bool{}, nil
	case "s8":
		return S8{}, nil
	case "u8":
		return U8{}, nil
	case "s16":
		return S16{}, nil
	case "u16":
		return U16{}, nil
	case "s32":
		return S32{}, nil
	case "u32":
		return U32{}, nil
	case "s64":
		return S64{}, nil
	case "u64":
		return U64{}, nil
	case "f32", "float32": // TODO: remove float32 at some point
		return F32{}, nil
	case "f64", "float64": // TODO: remove float64 at some point
		return F64{}, nil
	case "char":
		return Char{}, nil
	case "string":
		return String{}, nil
	case "error-context":
		return &ErrorContext{}, nil
	}
	return nil, fmt.Errorf("unknown primitive type %q", s)
}

// primitive is a type constraint of the Go equivalents of WIT [primitive types].
//
// [primitive types]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
type primitive interface {
	bool | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64 | float32 | float64 | char | string | errorContext
}

// char is defined because [rune] is an alias of [int32]
type char rune

// errorContext is defined because WIT error-context is a primitive type in the Component Model.
type errorContext uint32

// Primitive is the interface implemented by WIT [primitive types].
// It also conforms to the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive types]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
type Primitive interface {
	Type
	isPrimitive()
}

// _primitive is an embeddable struct that conforms to the [PrimitiveType] interface.
// It represents a WebAssembly Component Model [primitive type] mapped to its equivalent Go type.
// It also conforms to the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
type _primitive[T primitive] struct{ _type }

// isPrimitive conforms to the [Primitive] interface.
func (_primitive[T]) isPrimitive() {}

// Size returns the byte size for values of this type.
func (_primitive[T]) Size() uintptr {
	var v T
	switch any(v).(type) {
	case string:
		return 8 // [2]int32
	default:
		return unsafe.Sizeof(v)
	}
}

// Align returns the byte alignment for values of this type.
func (_primitive[T]) Align() uintptr {
	var v T
	switch any(v).(type) {
	case string:
		return 4 // int32
	default:
		return unsafe.Alignof(v)
	}
}

// hasPointer returns whether the ABI representation of this type contains a pointer.
// This will only return true for [String].
func (_primitive[T]) hasPointer() bool {
	var v T
	switch any(v).(type) {
	case string:
		return true
	default:
		return false
	}
}

// Flat returns the [flattened] ABI representation of this type.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (_primitive[T]) Flat() []Type {
	var v T
	switch any(v).(type) {
	case bool, int8, uint8, int16, uint16, int32, uint32, char, errorContext:
		return []Type{U32{}}
	case int64, uint64:
		return []Type{U64{}}
	case float32:
		return []Type{F32{}}
	case float64:
		return []Type{F64{}}
	case string:
		return []Type{PointerTo(U8{}), U32{}}
	default:
		panic(fmt.Sprintf("BUG: unknown primitive type %T", v)) // should never reach here
	}
}

// Bool represents the WIT [primitive type] bool, a boolean value either true or false.
// It is equivalent to the Go type [bool].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [bool]: https://pkg.go.dev/builtin#bool
type Bool struct{ _primitive[bool] }

// S8 represents the WIT [primitive type] s8, a signed 8-bit integer.
// It is equivalent to the Go type [int8].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [int8]: https://pkg.go.dev/builtin#int8
type S8 struct{ _primitive[int8] }

// U8 represents the WIT [primitive type] u8, an unsigned 8-bit integer.
// It is equivalent to the Go type [uint8].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [uint8]: https://pkg.go.dev/builtin#uint8
type U8 struct{ _primitive[uint8] }

// S16 represents the WIT [primitive type] s16, a signed 16-bit integer.
// It is equivalent to the Go type [int16].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [int16]: https://pkg.go.dev/builtin#int16
type S16 struct{ _primitive[int16] }

// U16 represents the WIT [primitive type] u16, an unsigned 16-bit integer.
// It is equivalent to the Go type [uint16].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [uint16]: https://pkg.go.dev/builtin#uint16
type U16 struct{ _primitive[uint16] }

// S32 represents the WIT [primitive type] s32, a signed 32-bit integer.
// It is equivalent to the Go type [int32].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [int32]: https://pkg.go.dev/builtin#int32
type S32 struct{ _primitive[int32] }

// U32 represents the WIT [primitive type] u32, an unsigned 32-bit integer.
// It is equivalent to the Go type [uint32].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [uint32]: https://pkg.go.dev/builtin#uint32
type U32 struct{ _primitive[uint32] }

// S64 represents the WIT [primitive type] s64, a signed 64-bit integer.
// It is equivalent to the Go type [int64].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [int64]: https://pkg.go.dev/builtin#int64
type S64 struct{ _primitive[int64] }

// U64 represents the WIT [primitive type] u64, an unsigned 64-bit integer.
// It is equivalent to the Go type [uint64].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [uint64]: https://pkg.go.dev/builtin#uint64
type U64 struct{ _primitive[uint64] }

// F32 represents the WIT [primitive type] f32, a 32-bit floating point value.
// It is equivalent to the Go type [float32].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [float32]: https://pkg.go.dev/builtin#float32
type F32 struct{ _primitive[float32] }

// F64 represents the WIT [primitive type] f64, a 64-bit floating point value.
// It is equivalent to the Go type [float64].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [float64]: https://pkg.go.dev/builtin#float64
type F64 struct{ _primitive[float64] }

// Char represents the WIT [primitive type] char, a single Unicode character,
// specifically a [Unicode scalar value]. It is equivalent to the Go type [rune].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [Unicode scalar value]: https://unicode.org/glossary/#unicode_scalar_value
// [rune]: https://pkg.go.dev/builtin#rune
type Char struct{ _primitive[char] }

// String represents the WIT [primitive type] string, a finite string of Unicode characters.
// It is equivalent to the Go type [string].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [string]: https://pkg.go.dev/builtin#string
type String struct{ _primitive[string] }

// ErrorContext represents a WIT [error-context] type.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [resource type]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/Explainer.md#error-context-type
type ErrorContext struct{ _primitive[errorContext] }
