package wit

import (
	"strings"

	"go.bytecodealliance.org/wit/clone"
)

// Function represents a WIT [function].
// Functions can be freestanding, methods, constructors or static.
// It implements the [Node] and [WorldItem] interfaces.
//
// [function]: https://component-model.bytecodealliance.org/design/wit.html#functions
type Function struct {
	_worldItem
	Name      string
	Kind      FunctionKind
	Params    []Param   // arguments to the function
	Results   []Param   // a function can have a single anonymous result, or > 1 named results
	Stability Stability // WIT @since or @unstable (nil if unknown)
	Docs      Docs
}

// DeepClone implements [clone.DeepClonable].
func (f *Function) DeepClone(state *clone.State) clone.DeepClonable {
	c := clone.Clone(state, f)
	c.Kind = *clone.DeepClone(state, &f.Kind)
	c.Params = clone.Slice(state, f.Params)
	c.Results = clone.Slice(state, f.Results)
	c.Stability = *clone.DeepClone(state, &f.Stability)
	return c
}

// BaseName returns the base name of [Function] f.
// For static functions, this returns the function name unchanged.
// For constructors, this removes the [constructor] and type prefix.
// For static functions, this removes the [static] and type prefix.
// For methods, this removes the [method] and type prefix.
// For special functions like [resource-drop], it will return a well-known value.
func (f *Function) BaseName() string {
	switch {
	case strings.HasPrefix(f.Name, "[constructor]"):
		return "constructor"
	case strings.HasPrefix(f.Name, "[resource-new]"):
		return "resource-new"
	case strings.HasPrefix(f.Name, "[resource-rep]"):
		return "resource-rep"
	case strings.HasPrefix(f.Name, "[resource-drop]"):
		return "resource-drop"
	case strings.HasPrefix(f.Name, "[dtor]"):
		return "destructor"
	}
	name, after, found := strings.Cut(f.Name, ".")
	if found {
		name = after
	}
	after, found = strings.CutPrefix(f.Name, "cabi_post_")
	if found {
		name = after + "-post-return"
	}
	return name
}

// Type returns the associated (self) [Type] for [Function] f, if f is a constructor, method, or static function.
// If f is a freestanding function, this returns nil.
func (f *Function) Type() Type {
	switch kind := f.Kind.(type) {
	case *Constructor:
		return kind.Type
	case *Static:
		return kind.Type
	case *Method:
		return kind.Type
	default:
		return nil
	}
}

// IsAdmin returns true if [Function] f is an administrative function in the Canonical ABI.
func (f *Function) IsAdmin() bool {
	switch {
	// Imported
	case f.IsStatic() && strings.HasPrefix(f.Name, "[resource-new]"):
		return true
	case f.IsMethod() && strings.HasPrefix(f.Name, "[resource-rep]"):
		return true
	case f.IsMethod() && strings.HasPrefix(f.Name, "[resource-drop]"):
		return true

	// Exported
	case f.IsMethod() && strings.HasPrefix(f.Name, "[dtor]"):
		return true
	case strings.HasPrefix(f.Name, "cabi_post_"):
		return true
	}
	return false
}

// IsFreestanding returns true if [Function] f is a freestanding function,
// and not a constructor, method, or static function.
func (f *Function) IsFreestanding() bool {
	_, ok := f.Kind.(*Freestanding)
	return ok
}

// IsConstructor returns true if [Function] f is a constructor.
// To qualify, it must have a *[Constructor] Kind with a non-nil type.
func (f *Function) IsConstructor() bool {
	kind, ok := f.Kind.(*Constructor)
	return ok && kind.Type != nil
}

// IsMethod returns true if [Function] f is a method.
// To qualify, it must have a *[Method] Kind with a non-nil [Type] which matches borrow<t> of its first param.
func (f *Function) IsMethod() bool {
	if len(f.Params) == 0 {
		return false
	}
	kind, ok := f.Kind.(*Method)
	if !ok {
		return false
	}
	t := f.Params[0].Type
	h := KindOf[*Borrow](t)
	return t == kind.Type || (h != nil && h.Type == kind.Type)
}

// IsStatic returns true if [Function] f is a static function.
// To qualify, it must have a *[Static] Kind with a non-nil type.
func (f *Function) IsStatic() bool {
	kind, ok := f.Kind.(*Static)
	return ok && kind.Type != nil
}

func (f *Function) dependsOn(dep Node) bool {
	if dep == f {
		return true
	}
	for _, p := range f.Params {
		if DependsOn(p.Type, dep) {
			return true
		}
	}
	for _, r := range f.Results {
		if DependsOn(r.Type, dep) {
			return true
		}
	}
	return false
}

func compareFunctions(a, b *Function) int {
	return strings.Compare(a.Name, b.Name)
}

// Param represents a parameter to or the result of a [Function].
// A Param can be unnamed.
type Param struct {
	Name string
	Type Type
}

// DeepClone implements [clone.DeepClonable].
func (p *Param) DeepClone(state *clone.State) clone.DeepClonable {
	c := clone.Clone(state, p)
	c.Type = *clone.DeepClone(state, &p.Type)
	return c
}

// FunctionKind represents the kind of a WIT [function], which can be one of
// [Freestanding], [Method], [Static], or [Constructor].
//
// [function]: https://component-model.bytecodealliance.org/design/wit.html#functions
type FunctionKind interface {
	isFunctionKind()
}

// _functionKind is an embeddable type that conforms to the [FunctionKind] interface.
type _functionKind struct{}

func (_functionKind) isFunctionKind() {}

// Freestanding represents a free-standing function that is not a method, static, or a constructor.
type Freestanding struct{ _functionKind }

// Method represents a function that is a method on its associated [Type].
// The first argument to the function is self, an instance of [Type].
type Method struct {
	_functionKind
	Type Type
}

// DeepClone implements [clone.DeepClonable].
func (m *Method) DeepClone(state *clone.State) clone.DeepClonable {
	c := clone.Clone(state, m)
	c.Type = *clone.DeepClone(state, &m.Type)
	return c
}

// Static represents a function that is a static method of its associated [Type].
type Static struct {
	_functionKind
	Type Type
}

// DeepClone implements [clone.DeepClonable].
func (s *Static) DeepClone(state *clone.State) clone.DeepClonable {
	c := clone.Clone(state, s)
	c.Type = *clone.DeepClone(state, &s.Type)
	return c
}

// Constructor represents a function that is a constructor for its associated [Type].
type Constructor struct {
	_functionKind
	Type Type
}

// DeepClone implements [clone.DeepClonable].
func (c *Constructor) DeepClone(state *clone.State) clone.DeepClonable {
	cl := clone.Clone(state, c)
	cl.Type = *clone.DeepClone(state, &c.Type)
	return cl
}
