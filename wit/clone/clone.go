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
	// First, check for nil
	if v == nil {
		return nil
	}

	// Check previous clones
	if c, ok := state.clones[v]; ok {
		return c.(*T)
	}

	// Check if *T implements Clonable
	if clonable, ok := any(v).(Clonable); ok {
		clone := any(clonable.CloneWith(state)).(*T)
		if state.clones == nil {
			state.clones = make(map[any]any)
		}
		state.clones[v] = clone
		return clone
	}

	// Check for underlying nil
	switch any(*v).(type) {
	case nil:
		var zero T
		return &zero
	}

	// Check if T was cloned
	if memoizable(v) {
		if c, ok := state.clones[*v]; ok {
			clone := c.(T)
			return &clone
		}
	}

	// Check if T implements Clonable
	var clone T
	if clonable, ok := any(*v).(Clonable); ok {
		clone = clonable.CloneWith(state).(T)
	} else {
		// Otherwise make shallow copy
		clone = *v
	}

	if memoizable(v) {
		if state.clones == nil {
			state.clones = make(map[any]any)
		}
		state.clones[*v] = clone
	}

	return &clone
}

// Pointers and interface types are memoizable
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
	for i := range clone {
		clone[i] = *Clone(state, &clone[i])
	}
	return clone
}

// Cloneable represents any type that can be cloned.
type Clonable interface {
	CloneWith(*State) Clonable
}

// State keeps track of previously cloned pointers, so circular data structures may be cloned.
// The zero value is safe for use.
type State struct {
	clones map[any]any
}
