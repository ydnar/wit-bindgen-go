package cm

import (
	"errors"
	"unsafe"
)

// IndexFunc is a function that returns an integer index of v.
// Used to reverse a string into a [variant] or [enum] case.
//
// [enum]: https://component-model.bytecodealliance.org/design/wit.html#enums
// [variant]: https://component-model.bytecodealliance.org/design/wit.html#variants
type IndexFunc[I Discriminant, V comparable] func(v V) (i I, ok bool)

// Index returns an [IndexFunc] that indexes the values slice.
// Panics on error.
func MustIndex[I Discriminant, V comparable](values []V) IndexFunc[I, V] {
	f, err := Index[I, V](values)
	if err != nil {
		panic(err)
	}
	return f
}

// Index returns an [IndexFunc] that indexes the values slice.
// Return an error if len(values) is too large to be indexed by I.
func Index[I Discriminant, V comparable](values []V) (IndexFunc[I, V], error) {
	if len(values) == 0 {
		return nil, errors.New("zero-length index")
	}
	if len(values) <= linearScanThreshold {
		return linearIndex[I, V](values).indexOf, nil
	}
	max := 1<<(unsafe.Sizeof(I(0))*8) - 1
	if len(values) > max {
		return nil, errors.New("len(values) exceeded index type")
	}
	m := make(mapIndex[I, V], len(values))
	for i, v := range values {
		m[v] = I(i)
	}
	return m.indexOf, nil
}

const linearScanThreshold = 16

type linearIndex[I Discriminant, V comparable] []V

func (idx linearIndex[I, V]) indexOf(v V) (i I, ok bool) {
	for i := 0; i < len(idx); i++ {
		if idx[i] == v {
			return I(i), true
		}
	}
	return 0, false
}

type mapIndex[I Discriminant, V comparable] map[V]I

func (idx mapIndex[I, V]) indexOf(v V) (i I, ok bool) {
	i, ok = idx[v]
	return
}
