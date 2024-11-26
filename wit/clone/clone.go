package clone

import (
	"unsafe"
)

// Clonable represents any type that can be cloned.
type Clonable[T any] interface {
	Clone() *T
}

// DeepClonable represents any type that can be deeply cloned.
// The returned value must be identical to the receiver type.
//
//	func (*T) DeepClone(*State) DeepClonable // returns *T
//	func (T) DeepClone(*State) DeepClonable // returns T
type DeepClonable interface {
	DeepClone(*State) DeepClonable
}

// Clone returns a shallow clone of pointer v, and
// records the cloned pointer in state.
func Clone[T any](state *State, v *T) *T {
	if clone := state.load(v); clone != nil {
		return clone.(*T)
	}
	var clone *T
	if clonable, ok := any(v).(Clonable[T]); ok {
		clone = clonable.Clone()
	} else {
		c := *v // clone by value
		clone = &c
	}
	state.store(v, clone)
	return clone
}

// DeepClone returns a deep clone of pointer v.
// If v was previously cloned, the earlier copy will be returned.
// If *T implements [DeepClonable], the value of DeepClone will be returned.
// Otherwise it returns a shallow copy, or nil if v is nil.
// To clone interface values, pass a pointer to the interface.
// The supplied [State] must not be nil.
func DeepClone[T any](state *State, v *T) *T {
	// First, check for nil
	if v == nil {
		return nil
	}

	// Check previous clones
	if clone := state.load(v); clone != nil {
		return clone.(*T)
	}

	// Check if *T implements DeepClonable
	if clonable, ok := any(v).(DeepClonable); ok {
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

	// Check if T implements DeepClonable
	if clonable, ok := any(*v).(DeepClonable); ok {
		clone := any(clonable.DeepClone(state)).(T)
		return &clone
	}

	// Otherwise make shallow copy
	return Clone(state, v)
}

// Slice returns a deep clone of slice s.
// Elements of s will be passed to [DeepClone].
// The supplied [State] must not be nil.
func Slice[S ~[]T, T any](state *State, s S) S {
	if s == nil {
		return s
	}
	clone := make(S, len(s))
	for i := range s {
		clone[i] = *DeepClone(state, &s[i])
	}
	return clone
}

// Map returns a deep clone of map m.
// Both keys and values of m will be passed to [DeepClone].
// Note: if  K is not a value type (e.g. int, string), the cloned map may not have identical keys to m.
// The supplied [State] must not be nil.
func Map[M ~map[K]V, K comparable, V any](state *State, m M) M {
	if m == nil {
		return nil
	}
	clone := make(M, len(m))
	for k, v := range m {
		clone[*DeepClone(state, &k)] = *DeepClone(state, &v)
	}
	return clone
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
