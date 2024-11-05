package cm

import (
	"reflect"
	"strings"
	"unsafe"

	"go.bytecodealliance.org/internal/tinyunsafe"
)

func typeName(v any) string {
	var name string
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		name = "*" + t.Elem().String()
	} else {
		name = t.String()
	}
	return strings.ReplaceAll(name, " ", "")
}

func sizePlusAlignOf[T any]() uintptr {
	var v T
	return unsafe.Sizeof(v) + unsafe.Alignof(v)
}

func alignOf[T any]() uintptr {
	var v T
	return unsafe.Alignof(v)
}

func zeroPtr[T any]() *T {
	var zero T
	return &zero
}

// VariantDebug is an interface used in tests to validate layout of variant types.
type VariantDebug interface {
	Size() uintptr
	DataAlign() uintptr
	DataOffset() uintptr
}

func (v variant[Disc, Shape, Align]) Size() uintptr       { return unsafe.Sizeof(v) }
func (v variant[Disc, Shape, Align]) DataAlign() uintptr  { return unsafe.Alignof(v.data) }
func (v variant[Disc, Shape, Align]) DataOffset() uintptr { return tinyunsafe.OffsetOf(&v, &v.data) }

// ResultDebug is an interface used in tests to validate layout of result types.
type ResultDebug interface {
	VariantDebug
}

func (r BoolResult) Size() uintptr       { return unsafe.Sizeof(r) }
func (r BoolResult) DataAlign() uintptr  { return 0 }
func (r BoolResult) DataOffset() uintptr { return 0 }

func (r result[Shape, OK, Err]) Size() uintptr       { return unsafe.Sizeof(r) }
func (r result[Shape, OK, Err]) DataAlign() uintptr  { return unsafe.Alignof(r.data) }
func (r result[Shape, OK, Err]) DataOffset() uintptr { return tinyunsafe.OffsetOf(&r, &r.data) }
