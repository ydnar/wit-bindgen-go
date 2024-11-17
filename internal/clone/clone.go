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
func Clone[T any](state *State, v T) T {
	i := any(v)
	switch i.(type) {
	case nil:
		return v
	}
	var clone T
	if memoizable(v) {
		if c, ok := state.clones[i]; ok {
			return c.(T)
		}
	}
	if clonable, ok := i.(Clonable); ok {
		clone = clonable.CloneWith(state).(T)
	} else {
		clone = v
	}
	if memoizable(v) {
		if state.clones == nil {
			state.clones = make(map[any]any)
		}
		state.clones[i] = clone
	}
	return clone
}

func memoizable[T any](v T) bool {
	return unsafe.Sizeof(v) == unsafe.Sizeof(any(nil)) || unsafe.Sizeof(v) == unsafe.Sizeof((*byte)(nil))
}

// Slice returns a copy of slice s.
// The supplied State must not be nil.
func Slice[S ~[]T, T any](state *State, s S) S {
	if s == nil {
		return s
	}
	clone := slices.Clone(s)
	for i, e := range clone {
		clone[i] = Clone(state, e)
	}
	return clone
}

// Cloneable represents any type that can be cloned.
type Clonable interface {
	CloneWith(*State) any
}

// State keeps track of previously cloned pointers, so circular data structures may be cloned.
// The zero value is safe for use.
type State struct {
	clones map[any]any
}
