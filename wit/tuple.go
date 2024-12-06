package wit

import (
	"strconv"
)

// Tuple represents a WIT [tuple type].
// A tuple type is an ordered fixed length sequence of values of specified types.
// It is similar to a [Record], except that the fields are identified by their order instead of by names.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [tuple type]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple struct {
	_typeDefKind
	Types []Type
}

// Type returns a non-nil [Type] if all types in t
// are the same. Returns nil if t contains more than one type.
func (t *Tuple) Type() Type {
	if len(t.Types) == 0 {
		return nil
	}
	typ := t.Types[0]
	for i := 0; i < len(t.Types); i++ {
		if t.Types[i] != typ {
			return nil
		}
	}
	return typ
}

// Despecialize despecializes [Tuple] e into a [Record] with 0-based integer field names.
// See the [canonical ABI documentation] for more information.
//
// [canonical ABI documentation]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (t *Tuple) Despecialize() TypeDefKind {
	r := &Record{
		Fields: make([]Field, len(t.Types)),
	}
	for i := range t.Types {
		r.Fields[i].Name = strconv.Itoa(i)
		r.Fields[i].Type = t.Types[i]
	}
	return r
}

// Size returns the [ABI byte size] for [Tuple] t.
// It is first [despecialized] into a [Record] with 0-based integer field names, then sized.
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
// [despecialized]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (t *Tuple) Size() uintptr {
	return t.Despecialize().Size()
}

// Align returns the [ABI byte alignment] for [Tuple] t.
// It is first [despecialized] into a [Record] with 0-based integer field names, then aligned.
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
// [despecialized]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (t *Tuple) Align() uintptr {
	return t.Despecialize().Align()
}

// Flat returns the [flattened] ABI representation of [Tuple] t.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (t *Tuple) Flat() []Type {
	return t.Despecialize().Flat()
}
