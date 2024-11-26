package wit

import "go.bytecodealliance.org/wit/clone"

// Option represents a WIT [option type], a special case of [Variant]. An Option can
// contain a value of a single type, either build-in or user defined, or no value.
// The equivalent in Go for an option<string> could be represented as *string.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [option type]: https://component-model.bytecodealliance.org/design/wit.html#options
type Option struct {
	_typeDefKind
	Type Type
}

// DeepClone implements [clone.DeepClonable].
func (o *Option) DeepClone(state *clone.State) clone.DeepClonable {
	c := clone.Clone(state, o)
	c.Type = *clone.DeepClone(state, &o.Type)
	return c
}

// Despecialize despecializes [Option] o into a [Variant] with two cases, "none" and "some".
// See the [canonical ABI documentation] for more information.
//
// [canonical ABI documentation]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (o *Option) Despecialize() TypeDefKind {
	return &Variant{
		Cases: []Case{
			{Name: "none"},
			{Name: "some", Type: o.Type},
		},
	}
}

// Size returns the [ABI byte size] for [Option] o.
// It is first [despecialized] into a [Variant] with two cases, "none" and "some(T)", then sized.
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
// [despecialized]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (o *Option) Size() uintptr {
	return o.Despecialize().Size()
}

// Align returns the [ABI byte alignment] for [Option] o.
// It is first [despecialized] into a [Variant] with two cases, "none" and "some(T)", then aligned.
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
// [despecialized]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (o *Option) Align() uintptr {
	return o.Despecialize().Align()
}

// Flat returns the [flattened] ABI representation of [Option] o.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (o *Option) Flat() []Type {
	return o.Despecialize().Flat()
}
