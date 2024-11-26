package clone

import (
	"unsafe"
)

// Shallow returns a shallow clone of pointer v, and
// records the cloned pointer in state.
// This is typically the first call in a Clone method.
func Shallow[T any](state *State, v *T) *T {
	if clone := state.load(v); clone != nil {
		return clone.(*T)
	}
	c := *v
	clone := &c
	state.store(v, clone)
	return clone
}

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
		clone := any(clonable.DeepClone(state)).(*T)
		return clone
	}

	// Check for underlying nil
	switch any(*v).(type) {
	case nil:
		var zero T
		return &zero
	}

	// Check if T was cloned
	if memoizable(*v) {
		if c := state.load(*v); c != nil {
			clone := c.(T)
			return &clone
		}
	}

	// Check if T implements Clonable
	if clonable, ok := any(*v).(Clonable); ok {
		clone := any(clonable.DeepClone(state)).(T)
		return &clone
	}

	// Otherwise make shallow copy
	return Shallow(state, v)
}

// Slice returns a clone of slice s.
// Elements of s will be passed to [Clone].
// The supplied [State] must not be nil.
func Slice[S ~[]T, T any](state *State, s S) S {
	if s == nil {
		return s
	}
	clone := make(S, len(s))
	for i := range s {
		clone[i] = *Clone(state, &s[i])
	}
	return clone
}

// Map returns a clone of map m.
// Both keys and values of m will be passed to [Clone].
// Note: if  K is not a value type (e.g. int, string), the cloned map may not have identical keys to m.
// The supplied [State] must not be nil.
func Map[M ~map[K]V, K comparable, V any](state *State, m M) M {
	if m == nil {
		return nil
	}
	clone := make(M, len(m))
	for k, v := range m {
		clone[*Clone(state, &k)] = *Clone(state, &v)
	}
	return clone
}

// Cloneable represents any type that can be cloned.
// The returned value must be identical to the receiver type.
//
//	func (*T) DeepClone(*State) Clonable // returns *T
//	func (T) DeepClone(*State) Clonable // returns T
type Clonable interface {
	DeepClone(*State) Clonable
}

// State tracks previously cloned values to enable cloning of circular data structures.
// The zero value is safe for use.
type State struct {
	clones map[any]any
}

func (state *State) load(v any) any {
	clone, ok := state.clones[v]
	if ok {
		// fmt.Printf("loaded clone of type %T (%p -> %p)\n", v, v, clone)
	}
	return clone
}

func (state *State) store(v, clone any) {
	if state.clones == nil {
		state.clones = make(map[any]any)
	}
	// fmt.Printf("stored clone of type %T (%p -> %p)\n", v, v, clone)
	state.clones[v] = clone
	state.clones[clone] = clone // handle circular data structures
}

func memoizable[T any](v T) bool {
	// Pointers and interface types are memoizable
	return unsafe.Sizeof(v) == unsafe.Sizeof(any(nil)) || unsafe.Sizeof(v) == unsafe.Sizeof((*byte)(nil))
}
