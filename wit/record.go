package wit

import "go.bytecodealliance.org/wit/clone"

// Record represents a WIT [record type], akin to a struct.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [record type]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#item-record-bag-of-named-fields
type Record struct {
	_typeDefKind
	Fields []Field
}

// DeepClone implements [clone.DeepClonable].
func (r *Record) DeepClone(state *clone.State) clone.DeepClonable {
	c := clone.Clone(state, r)
	c.Fields = clone.Slice(state, r.Fields)
	return c
}

// Size returns the [ABI byte size] for [Record] r.
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (r *Record) Size() uintptr {
	var s uintptr
	for _, f := range r.Fields {
		s = Align(s, f.Type.Align())
		s += f.Type.Size()
	}
	return s
}

// Align returns the [ABI byte alignment] for [Record] r.
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (r *Record) Align() uintptr {
	var a uintptr = 1
	for _, f := range r.Fields {
		a = max(a, f.Type.Align())
	}
	return a
}

// Flat returns the [flattened] ABI representation of [Record] r.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (r *Record) Flat() []Type {
	flat := make([]Type, 0, len(r.Fields))
	for _, f := range r.Fields {
		flat = append(flat, f.Type.Flat()...)
	}
	return flat
}

func (r *Record) hasPointer() bool {
	for _, f := range r.Fields {
		if HasPointer(f.Type) {
			return true
		}
	}
	return false
}

func (r *Record) hasBorrow() bool {
	for _, f := range r.Fields {
		if HasBorrow(f.Type) {
			return true
		}
	}
	return false
}

func (r *Record) hasResource() bool {
	for _, f := range r.Fields {
		if HasResource(f.Type) {
			return true
		}
	}
	return false
}

func (r *Record) dependsOn(dep Node) bool {
	if dep == r {
		return true
	}
	for _, f := range r.Fields {
		if DependsOn(f.Type, dep) {
			return true
		}
	}
	return false
}

// Field represents a field in a [Record].
type Field struct {
	Name string
	Type Type
	Docs Docs
}

// DeepClone implements [clone.DeepClonable].
func (f *Field) DeepClone(state *clone.State) clone.DeepClonable {
	c := clone.Clone(state, f)
	c.Type = *clone.DeepClone(state, &f.Type)
	return c
}
