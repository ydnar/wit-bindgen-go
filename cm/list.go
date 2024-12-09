package cm

import (
	"bytes"
	"encoding/json"
	"unsafe"
)

// nullLiteral is the JSON representation of a null literal.
// https://pkg.go.dev/encoding/json#Unmarshaler
// By convention, to approximate the behavior of Unmarshal itself, Unmarshalers implement UnmarshalJSON([]byte("null")) as a no-op.
var nullLiteral = []byte("null")

// List represents a Component Model list.
// The binary representation of list<T> is similar to a Go slice minus the cap field.
type List[T any] struct {
	_ HostLayout
	list[T]
}

// AnyList is a type constraint for generic functions that accept any [List] type.
type AnyList[T any] interface {
	~struct {
		_ HostLayout
		list[T]
	}
}

// NewList returns a List[T] from data and len.
func NewList[T any, Len AnyInteger](data *T, len Len) List[T] {
	return List[T]{
		list: list[T]{
			data: data,
			len:  uintptr(len),
		},
	}
}

// ToList returns a List[T] equivalent to the Go slice s.
// The underlying slice data is not copied, and the resulting List points at the
// same array storage as the slice.
func ToList[S ~[]T, T any](s S) List[T] {
	return NewList[T](unsafe.SliceData([]T(s)), uintptr(len(s)))
}

// list represents the internal representation of a Component Model list.
// It is intended to be embedded in a [List], so embedding types maintain
// the methods defined on this type.
type list[T any] struct {
	_    HostLayout
	data *T
	len  uintptr
}

// Slice returns a Go slice representing the List.
func (l list[T]) Slice() []T {
	return unsafe.Slice(l.data, l.len)
}

// Data returns the data pointer for the list.
func (l list[T]) Data() *T {
	return l.data
}

// Len returns the length of the list.
// TODO: should this return an int instead of a uintptr?
func (l list[T]) Len() uintptr {
	return l.len
}

// MarshalJSON implements json.Marshaler.
func (l list[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Slice())
}

// UnmarshalJSON implements json.Unmarshaler.
func (l *list[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullLiteral) {
		return nil
	}

	var s []T
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	l.data = unsafe.SliceData([]T(s))
	l.len = uintptr(len(s))

	return nil
}
