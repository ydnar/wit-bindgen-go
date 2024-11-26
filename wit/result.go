package wit

import "go.bytecodealliance.org/wit/clone"

// Result represents a WIT [result type], which is the result of a function call,
// returning an optional value and/or an optional error. It is roughly equivalent to
// the Go pattern of returning (T, error).
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [result type]: https://component-model.bytecodealliance.org/design/wit.html#results
type Result struct {
	_typeDefKind
	OK  Type // optional associated [Type] (can be nil)
	Err Type // optional associated [Type] (can be nil)
}

// DeepClone implements [clone.Clonable].
func (r *Result) DeepClone(state *clone.State) clone.Clonable {
	c := clone.Shallow(state, r)
	c.OK = *clone.Clone(state, &r.OK)
	c.Err = *clone.Clone(state, &r.Err)
	return c
}

// Despecialize despecializes [Result] o into a [Variant] with two cases, "ok" and "error".
// See the [canonical ABI documentation] for more information.
//
// [canonical ABI documentation]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (r *Result) Despecialize() TypeDefKind {
	return &Variant{
		Cases: []Case{
			{Name: "ok", Type: r.OK},
			{Name: "error", Type: r.Err},
		},
	}
}

// Types returns the unique associated types in [Result] r.
func (r *Result) Types() []Type {
	var types []Type
	if r.OK != nil {
		types = append(types, r.OK)
	}
	if r.Err != nil && r.Err != r.OK {
		types = append(types, r.Err)
	}
	return types
}

// Size returns the [ABI byte size] for [Result] r.
// It is first [despecialized] into a [Variant] with two cases "ok" and "error", then sized.
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
// [despecialized]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (r *Result) Size() uintptr {
	return r.Despecialize().Size()
}

// Align returns the [ABI byte alignment] for [Result] r.
// It is first [despecialized] into a [Variant] with two cases "ok" and "error", then aligned.
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
// [despecialized]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (r *Result) Align() uintptr {
	return r.Despecialize().Align()
}

// Flat returns the [flattened] ABI representation of [Result] r.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (r *Result) Flat() []Type {
	return r.Despecialize().Flat()
}
