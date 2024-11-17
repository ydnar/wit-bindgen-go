package clone

import (
	"slices"
	"unsafe"
)

// Clone returns a copy of v.
// If v was previously cloned, the earlier copy will be returned.
// If v implements [Clonable], the value of CloneWith will be returned.
// Otherwise it returns a shallow copy, or nil if v is nil.
// The supplied State must not be nil.
func Clone[T any](state *State, v *T) *T {
	if v == nil {
		return nil
	}
	p := unsafe.Pointer(v)
	if c, ok := state.clones[p]; ok {
		return (*T)(c)
	}
	var clone *T
	if clonable, ok := any(v).(Clonable[T]); ok {
		clone = clonable.CloneWith(state)
	} else {
		shallow := *v
		clone = &shallow
	}
	if state.clones == nil {
		state.clones = make(map[unsafe.Pointer]unsafe.Pointer)
	}
	state.clones[p] = unsafe.Pointer(clone)
	return clone
}

// Slice returns a copy of slice s.
// The supplied State must not be nil.
func Slice[S ~[]T, T any](state *State, s S) S {
	if s == nil {
		return s
	}
	clone := slices.Clone(s)
	// for i := range clone {
	// 	e, ok := any(clone[i]).(Clonable[T])
	// 	if !ok {
	// 		break
	// 	}

	// }
	return clone
}

// Cloneable represents any type that can be cloned.
type Clonable[T any] interface {
	CloneWith(*State) *T
}

// State keeps track of previously cloned pointers, so circular data structures may be cloned.
// The zero value is safe for use.
type State struct {
	clones map[unsafe.Pointer]unsafe.Pointer
}
