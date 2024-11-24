package clone

import (
	"slices"
	"unsafe"
)

// Clone returns a clone of pointer v.
// If v was previously cloned, the earlier copy will be returned.
// If *T implements [Clonable], the value of Clone will be returned.
// Otherwise it returns a shallow copy, or nil if v is nil.
// To clone interface values, pass a pointer to the interface.
// The supplied [State] must not be nil.
func Clone[T any](state *State, v *T) *T {
	// First, check for nil
	if v == nil {
		return nil
	}

	// Check previous clones
	if clone := state.load(v); clone != nil {
		return clone.(*T)
	}

	// Check if *T implements Clonable
	if clonable, ok := any(v).(Clonable); ok {
		clone := any(clonable.Clone(state)).(*T)
		state.store(v, clone)
		return clone
	}

	// Check for underlying nil
	switch any(*v).(type) {
	case nil:
		var zero T
		return &zero
	}

	// Check if T was cloned
	if clone := state.load(*v); clone != nil {
		clone := clone.(T)
		return &clone
	}

	// Check if T implements Clonable
	var clone T
	if clonable, ok := any(*v).(Clonable); ok {
		clone = clonable.Clone(state).(T)
	} else {
		// Otherwise make shallow copy
		clone = *v
	}

	state.store(*v, clone)

	return &clone
}

// Slice returns a clone of slice s.
// Elements of s will be passed to [Clone].
// The supplied [State] must not be nil.
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
// The returned value must be identical to the receiver type.
//
//	func (*T) Clone(*State) Clonable // returns *T
//	func (T) Clone(*State) Clonable // returns T
type Clonable interface {
	Clone(*State) Clonable
}

// State tracks previously cloned values to enable cloning of circular data structures.
// The zero value is safe for use.
type State struct {
	clones map[any]any
}

func (state *State) load(v any) any {
	if !memoizable(v) {
		return nil
	}
	return state.clones[v]
}

func (state *State) store(v, clone any) {
	if !memoizable(v) {
		return
	}
	if state.clones == nil {
		state.clones = make(map[any]any)
	}
	state.clones[v] = clone
}

func memoizable[T any](v T) bool {
	// Pointers and interface types are memoizable
	return unsafe.Sizeof(v) == unsafe.Sizeof(any(nil)) || unsafe.Sizeof(v) == unsafe.Sizeof((*byte)(nil))
}
