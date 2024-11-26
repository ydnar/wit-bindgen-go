package wit

import "go.bytecodealliance.org/wit/clone"

// Variant represents a WIT [variant type], a tagged/discriminated union.
// A variant type declares one or more cases. Each case has a name and, optionally,
// a type of data associated with that case.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [variant type]: https://component-model.bytecodealliance.org/design/wit.html#variants
type Variant struct {
	_typeDefKind
	Cases []Case
}

// DeepClone implements [clone.Clonable].
func (v *Variant) DeepClone(state *clone.State) clone.Clonable {
	c := clone.Shallow(state, v)
	c.Cases = clone.Slice(state, v.Cases)
	return c
}

// Enum attempts to represent [Variant] v as an [Enum].
// This will only succeed if v has no associated types. If v has
// associated types, then it will return nil.
func (v *Variant) Enum() *Enum {
	types := v.Types()
	if len(types) > 0 {
		return nil
	}
	e := &Enum{
		Cases: make([]EnumCase, len(v.Cases)),
	}
	for i := range v.Cases {
		e.Cases[i].Name = v.Cases[i].Name
		e.Cases[i].Docs = v.Cases[i].Docs
	}
	return e
}

// Types returns the unique associated types in [Variant] v.
func (v *Variant) Types() []Type {
	var types []Type
	typeMap := make(map[Type]bool)
	for i := range v.Cases {
		t := v.Cases[i].Type
		if t == nil || typeMap[t] {
			continue
		}
		types = append(types, t)
		typeMap[t] = true
	}
	return types
}

// Size returns the [ABI byte size] for [Variant] v.
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (v *Variant) Size() uintptr {
	s := Discriminant(len(v.Cases)).Size()
	s = Align(s, v.maxCaseAlign())
	s += v.maxCaseSize()
	return Align(s, v.Align())
}

// Align returns the [ABI byte alignment] for [Variant] v.
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (v *Variant) Align() uintptr {
	return max(Discriminant(len(v.Cases)).Align(), v.maxCaseAlign())
}

// Flat returns the [flattened] ABI representation of [Variant] v.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (v *Variant) Flat() []Type {
	var flat []Type
	for _, t := range v.Types() {
		for i, f := range t.Flat() {
			if i >= len(flat) {
				flat = append(flat, f)
			} else {
				flat[i] = flatJoin(flat[i], f)
			}
		}
	}
	return append(Discriminant(len(v.Cases)).Flat(), flat...)
}

func flatJoin(a, b Type) Type {
	if a == b {
		return a
	}
	if a.Size() == 4 && b.Size() == 4 {
		return U32{}
	}
	return U64{}
}

func (v *Variant) maxCaseSize() uintptr {
	var s uintptr
	for _, c := range v.Cases {
		if c.Type != nil {
			s = max(s, c.Type.Size())
		}
	}
	return s
}

func (v *Variant) maxCaseAlign() uintptr {
	var a uintptr = 1
	for _, c := range v.Cases {
		if c.Type != nil {
			a = max(a, c.Type.Align())
		}
	}
	return a
}

func (v *Variant) hasPointer() bool {
	for _, t := range v.Types() {
		if HasPointer(t) {
			return true
		}
	}
	return false
}

func (v *Variant) hasBorrow() bool {
	for _, t := range v.Types() {
		if HasBorrow(t) {
			return true
		}
	}
	return false
}

func (v *Variant) hasResource() bool {
	for _, t := range v.Types() {
		if HasResource(t) {
			return true
		}
	}
	return false
}

func (v *Variant) dependsOn(dep Node) bool {
	if dep == v {
		return true
	}
	for _, t := range v.Types() {
		if DependsOn(t, dep) {
			return true
		}
	}
	return false
}

// Case represents a single case in a [Variant].
// It implements the [Node] interface.
type Case struct {
	Name string
	Type Type // optional associated [Type] (can be nil)
	Docs Docs
}

// DeepClone implements [clone.Clonable].
func (c *Case) DeepClone(state *clone.State) clone.Clonable {
	cl := clone.Shallow(state, c)
	cl.Type = *clone.Clone(state, &c.Type)
	return cl
}

func (c *Case) dependsOn(dep Node) bool { return dep == c || DependsOn(c.Type, dep) }
