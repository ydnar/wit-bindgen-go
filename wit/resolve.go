package wit

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unsafe"

	"github.com/coreos/go-semver/semver"
	"go.bytecodealliance.org/wit/iterate"
	"go.bytecodealliance.org/wit/ordered"
)

// Resolve represents a fully resolved set of WIT ([WebAssembly Interface Type])
// packages and worlds. It implements the [Node] interface.
//
// This structure contains a graph of WIT packages and their contents
// merged together into slices organized by type. Items are sorted
// topologically and everything is fully resolved.
//
// Each [World], [Interface], [TypeDef], or [Package] in a Resolve must be non-nil.
//
// [WebAssembly Interface Type]: https://component-model.bytecodealliance.org/design/wit.html
type Resolve struct {
	Worlds     []*World
	Interfaces []*Interface
	TypeDefs   []*TypeDef
	Packages   []*Package
}

// AllFunctions returns a [sequence] that yields each [Function] in a [Resolve].
// The sequence stops if yield returns false.
//
// [sequence]: https://github.com/golang/go/issues/61897
func (r *Resolve) AllFunctions() iterate.Seq[*Function] {
	return func(yield func(*Function) bool) {
		var done bool
		yield = iterate.Done(iterate.Once(yield), func() { done = true })
		for i := 0; i < len(r.Worlds) && !done; i++ {
			r.Worlds[i].AllFunctions()(yield)
		}
		for i := 0; i < len(r.Interfaces) && !done; i++ {
			r.Interfaces[i].AllFunctions()(yield)
		}
	}
}

// A World represents all of the imports and exports of a [WebAssembly component].
// It implements the [Node] and [TypeOwner] interfaces.
//
// [WebAssembly component]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#wit-worlds
type World struct {
	_typeOwner

	Name      string
	Imports   ordered.Map[string, WorldItem]
	Exports   ordered.Map[string, WorldItem]
	Package   *Package  // the Package this World belongs to (must be non-nil)
	Stability Stability // WIT @since or @unstable (nil if unknown)
	Docs      Docs
}

func (w *World) dependsOn(pkg *Package) bool {
	if w.Package == pkg {
		return true
	}
	var done bool
	w.AllItems()(func(_ string, i WorldItem) bool {
		done = DependsOn(i, pkg)
		return !done
	})
	return done
}

// WITPackage returns the [Package] that [World] w belongs to.
func (w *World) WITPackage() *Package {
	return w.Package
}

// Match returns true if [World] w matches pattern, which can be one of:
// "name", "namespace:package/name" (qualified), or "namespace:package/name@1.0.0" (versioned).
func (w *World) Match(pattern string) bool {
	if pattern == w.Name {
		return true
	}
	id := w.Package.Name
	id.Extension = w.Name
	if pattern == id.String() {
		return true
	}
	id.Version = nil
	return pattern == id.String()
}

// HasInterface returns true if [World] w references [Interface] i.
func (w *World) HasInterface(i *Interface) bool {
	var found bool
	w.AllInterfaces()(func(_ string, face *Interface) bool {
		found = face == i
		return !found
	})
	return found
}

// AllInterfaces returns a [sequence] that yields each [Interface] in a [World].
// The sequence stops if yield returns false.
//
// [sequence]: https://github.com/golang/go/issues/61897
func (w *World) AllInterfaces() iterate.Seq2[string, *Interface] {
	return func(yield func(string, *Interface) bool) {
		w.AllItems()(func(name string, i WorldItem) bool {
			if ref, ok := i.(*InterfaceRef); ok {
				return yield(name, ref.Interface)
			}
			return true
		})
	}
}

// AllTypeDefs returns a [sequence] that yields each [TypeDef] in a [World].
// The sequence stops if yield returns false.
//
// [sequence]: https://github.com/golang/go/issues/61897
func (w *World) AllTypeDefs() iterate.Seq2[string, *TypeDef] {
	return func(yield func(string, *TypeDef) bool) {
		w.AllItems()(func(name string, i WorldItem) bool {
			if t, ok := i.(*TypeDef); ok {
				return yield(name, t)
			}
			return true
		})
	}
}

// AllFunctions returns a [sequence] that yields each [Function] in a [World].
// The sequence stops if yield returns false.
//
// [sequence]: https://github.com/golang/go/issues/61897
func (w *World) AllFunctions() iterate.Seq[*Function] {
	return func(yield func(*Function) bool) {
		w.AllItems()(func(_ string, i WorldItem) bool {
			if f, ok := i.(*Function); ok {
				return yield(f)
			}
			return true
		})
	}
}

// AllItems returns a [sequence] that yields each [WorldItem] in a [World].
// The sequence stops if yield returns false.
//
// [sequence]: https://github.com/golang/go/issues/61897
func (w *World) AllItems() iterate.Seq2[string, WorldItem] {
	return func(yield func(string, WorldItem) bool) {
		var done bool
		yield = iterate.Done2(iterate.Once2(yield), func() { done = true })
		f := func(name string, i WorldItem) bool {
			return yield(name, i)
		}
		w.Imports.All()(f)
		if done {
			return
		}
		w.Exports.All()(f)
	}
}

// A WorldItem is any item that can be exported from or imported into a [World],
// currently either an [InterfaceRef], [TypeDef], or [Function].
// Any WorldItem is also a [Node].
type WorldItem interface {
	Node
	isWorldItem()
}

// _worldItem is an embeddable type that conforms to the [WorldItem] interface.
type _worldItem struct{}

func (_worldItem) isWorldItem() {}

// An InterfaceRef represents a reference to an [Interface] with a [Stability] attribute.
// It implements the [Node] and [WorldItem] interfaces.
type InterfaceRef struct {
	_worldItem

	Interface *Interface
	Stability Stability
}

func (ref *InterfaceRef) dependsOn(pkg *Package) bool {
	return DependsOn(ref.Interface, pkg)
}

// An Interface represents a [collection of types and functions], which are imported into
// or exported from a [WebAssembly component].
// It implements the [Node], and [TypeOwner] interfaces.
//
// [collection of types and functions]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#wit-interfaces.
// [WebAssembly component]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#wit-worlds
type Interface struct {
	_typeOwner

	Name      *string
	TypeDefs  ordered.Map[string, *TypeDef]
	Functions ordered.Map[string, *Function]
	Package   *Package  // the Package this Interface belongs to
	Stability Stability // WIT @since or @unstable (nil if unknown)
	Docs      Docs
}

func (i *Interface) dependsOn(pkg *Package) bool {
	if i.Package == pkg {
		return true
	}
	var done bool
	i.TypeDefs.All()(func(_ string, t *TypeDef) bool {
		done = DependsOn(t, pkg)
		return !done
	})
	if done {
		return true
	}
	i.Functions.All()(func(_ string, f *Function) bool {
		done = DependsOn(f, pkg)
		return !done
	})
	return done
}

// WITPackage returns the [Package] this [Interface] belongs to.
func (i *Interface) WITPackage() *Package {
	return i.Package
}

// AllFunctions returns a [sequence] that yields each [Function] in an [Interface].
// The sequence stops if yield returns false.
//
// [sequence]: https://github.com/golang/go/issues/61897
func (i *Interface) AllFunctions() iterate.Seq[*Function] {
	return func(yield func(*Function) bool) {
		i.Functions.All()(func(_ string, f *Function) bool {
			return yield(f)
		})
	}
}

// TypeDef represents a WIT type definition. A TypeDef may be named or anonymous,
// and optionally belong to a [World] or [Interface].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
type TypeDef struct {
	_type
	_worldItem
	Name      *string
	Kind      TypeDefKind
	Owner     TypeOwner
	Stability Stability // WIT @since or @unstable (nil if unknown)
	Docs      Docs
}

// TypeName returns the [WIT] type name for t.
// Returns an empty string if t is anonymous.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (t *TypeDef) TypeName() string {
	if t.Name != nil {
		return *t.Name
	}
	return ""
}

// TypeDef returns the parent [TypeDef] of [TypeDef] t.
// If t is not a [type alias], TypeDef returns t.
//
// [type alias]: https://component-model.bytecodealliance.org/design/wit.html#type-aliases
func (t *TypeDef) TypeDef() *TypeDef {
	if t, ok := t.Kind.(*TypeDef); ok {
		return t
	}
	return t
}

// Root returns the root [TypeDef] of [TypeDef] t.
// If t is not a [type alias], Root returns t.
//
// [type alias]: https://component-model.bytecodealliance.org/design/wit.html#type-aliases
func (t *TypeDef) Root() *TypeDef {
	for {
		switch kind := t.Kind.(type) {
		case *TypeDef:
			t = kind
		default:
			return t
		}
	}
}

// Constructor returns the constructor for [TypeDef] t, or nil if none.
// Currently t must be a [Resource] to have a constructor.
func (t *TypeDef) Constructor() *Function {
	var constructor *Function
	t.Owner.AllFunctions()(func(f *Function) bool {
		if c, ok := f.Kind.(*Constructor); ok && c.Type == t {
			constructor = f
			return false
		}
		return true
	})
	return constructor
}

// StaticFunctions returns all static functions for [TypeDef] t.
// Currently t must be a [Resource] to have static functions.
func (t *TypeDef) StaticFunctions() []*Function {
	var statics []*Function
	t.Owner.AllFunctions()(func(f *Function) bool {
		if s, ok := f.Kind.(*Static); ok && s.Type == t {
			statics = append(statics, f)
		}
		return true
	})
	slices.SortFunc(statics, compareFunctions)
	return statics
}

// Methods returns all methods for [TypeDef] t.
// Currently t must be a [Resource] to have methods.
func (t *TypeDef) Methods() []*Function {
	var methods []*Function
	t.Owner.AllFunctions()(func(f *Function) bool {
		if m, ok := f.Kind.(*Method); ok && m.Type == t {
			methods = append(methods, f)
		}
		return true
	})
	slices.SortFunc(methods, compareFunctions)
	return methods
}

func compareFunctions(a, b *Function) int {
	return strings.Compare(a.Name, b.Name)
}

// Size returns the byte size for values of type t.
func (t *TypeDef) Size() uintptr {
	return t.Kind.Size()
}

// Align returns the byte alignment for values of type t.
func (t *TypeDef) Align() uintptr {
	return t.Kind.Align()
}

// Flat returns the [flattened] ABI representation of t.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (t *TypeDef) Flat() []Type {
	return t.Kind.Flat()
}

func (t *TypeDef) dependsOn(pkg *Package) bool {
	if t.Owner != nil && t.Owner.WITPackage() == pkg {
		return true
	}
	return DependsOn(t.Kind, pkg)
}

func (t *TypeDef) hasPointer() bool  { return HasPointer(t.Kind) }
func (t *TypeDef) hasBorrow() bool   { return HasBorrow(t.Kind) }
func (t *TypeDef) hasResource() bool { return HasResource(t.Kind) }

// TypeDefKind represents the underlying type in a [TypeDef], which can be one of
// [Record], [Resource], [Handle], [Flags], [Tuple], [Variant], [Enum],
// [Option], [Result], [List], [Future], [Stream], or [Type].
// It implements the [Node] and [ABI] interfaces.
type TypeDefKind interface {
	Node
	ABI
	isTypeDefKind()
}

// _typeDefKind is an embeddable type that conforms to the [TypeDefKind] interface.
type _typeDefKind struct{}

func (_typeDefKind) isTypeDefKind() {}

// KindOf probes [Type] t to determine if it is a [TypeDef] with [TypeDefKind] K.
// It returns the underlying Kind if present.
func KindOf[K TypeDefKind](t Type) (kind K) {
	if td, ok := t.(*TypeDef); ok {
		if kind, ok = td.Kind.(K); ok {
			return kind
		}
	}
	var zero K
	return zero
}

// PointerTo returns a [Pointer] to [Type] t.
func PointerTo(t Type) *TypeDef {
	return &TypeDef{Kind: &Pointer{Type: t}}
}

// Pointer represents a pointer to a WIT type.
// It is only used for ABI representation, e.g. pointers to function parameters or return values.
type Pointer struct {
	_typeDefKind
	Type Type
}

// Size returns the [ABI byte size] for [Pointer].
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (*Pointer) Size() uintptr { return 4 }

// Align returns the [ABI byte alignment] for [Pointer].
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (*Pointer) Align() uintptr { return 4 }

// Flat returns the [flattened] ABI representation of [Pointer].
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (p *Pointer) Flat() []Type { return []Type{PointerTo(p.Type)} }

func (p *Pointer) dependsOn(pkg *Package) bool { return DependsOn(p.Type, pkg) }
func (*Pointer) hasPointer() bool              { return true }
func (p *Pointer) hasBorrow() bool             { return HasBorrow(p.Type) }
func (p *Pointer) hasResource() bool           { return HasResource(p.Type) }

// Record represents a WIT [record type], akin to a struct.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [record type]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#item-record-bag-of-named-fields
type Record struct {
	_typeDefKind
	Fields []Field
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

func (r *Record) dependsOn(pkg *Package) bool {
	for _, f := range r.Fields {
		if DependsOn(f.Type, pkg) {
			return true
		}
	}
	return false
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

// Field represents a field in a [Record].
type Field struct {
	Name string
	Type Type
	Docs Docs
}

// Resource represents a WIT [resource type].
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [resource type]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#item-resource
type Resource struct{ _typeDefKind }

// Size returns the [ABI byte size] for [Resource].
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (*Resource) Size() uintptr { return 4 }

// Align returns the [ABI byte alignment] for [Resource].
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (*Resource) Align() uintptr { return 4 }

// Flat returns the [flattened] ABI representation of [Resource].
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (*Resource) Flat() []Type { return []Type{U32{}} }

// hasResource always returns true.
func (*Resource) hasResource() bool { return true }

// Handle represents a WIT [handle type].
// It conforms to the [Node], [ABI], and [TypeDefKind] interfaces.
// Handles represent the passing of unique ownership of a resource between
// two components. When the owner of an owned handle drops that handle,
// the resource is destroyed. In contrast, a borrowed handle represents
// a temporary loan of a handle from the caller to the callee for the
// duration of the call.
//
// [handle type]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#handles
type Handle interface {
	TypeDefKind
	isHandle()
}

// _handle is an embeddable type that conforms to the [Handle] interface.
type _handle struct{ _typeDefKind }

func (_handle) isHandle() {}

// Size returns the [ABI byte size] for this [Handle].
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (_handle) Size() uintptr { return 4 }

// Align returns the [ABI byte alignment] for this [Handle].
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (_handle) Align() uintptr { return 4 }

// Flat returns the [flattened] ABI representation of this type.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (_handle) Flat() []Type { return []Type{U32{}} }

// Own represents an WIT [owned handle].
// It implements the [Handle], [Node], [ABI], and [TypeDefKind] interfaces.
//
// [owned handle]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#handles
type Own struct {
	_handle
	Type *TypeDef
}

func (o *Own) dependsOn(pkg *Package) bool { return DependsOn(o.Type, pkg) }
func (o *Own) hasResource() bool           { return HasResource(o.Type) }

// Borrow represents a WIT [borrowed handle].
// It implements the [Handle], [Node], [ABI], and [TypeDefKind] interfaces.
//
// [borrowed handle]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#handles
type Borrow struct {
	_handle
	Type *TypeDef
}

func (b *Borrow) dependsOn(pkg *Package) bool { return DependsOn(b.Type, pkg) }
func (b *Borrow) hasBorrow() bool             { return true }
func (b *Borrow) hasResource() bool           { return HasResource(b.Type) }

// Flags represents a WIT [flags type], stored as a bitfield.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [flags type]: https://component-model.bytecodealliance.org/design/wit.html#flags
type Flags struct {
	_typeDefKind
	Flags []Flag
}

// Size returns the [ABI byte size] of [Flags] f.
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (f *Flags) Size() uintptr {
	n := len(f.Flags)
	switch {
	case n <= 8:
		return 1
	case n <= 16:
		return 2
	}
	return 4 * uintptr((n+31)>>5)
}

// Align returns the [ABI byte alignment] of [Flags] f.
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (f *Flags) Align() uintptr {
	n := len(f.Flags)
	switch {
	case n <= 8:
		return 1
	case n <= 16:
		return 2
	}
	return 4
}

// Flat returns the [flattened] ABI representation of [Flags] f.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (f *Flags) Flat() []Type {
	flat := make([]Type, (len(f.Flags)+31)>>5)
	for i := range flat {
		flat[i] = U32{}
	}
	return flat
}

// Flag represents a single flag value in a [Flags] type.
// It implements the [Node] interface.
type Flag struct {
	Name string
	Docs Docs
}

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

func (v *Variant) dependsOn(pkg *Package) bool {
	for _, t := range v.Types() {
		if DependsOn(t, pkg) {
			return true
		}
	}
	return false
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

// Case represents a single case in a [Variant].
// It implements the [Node] interface.
type Case struct {
	Name string
	Type Type // optional associated [Type] (can be nil)
	Docs Docs
}

func (c *Case) dependsOn(pkg *Package) bool {
	return DependsOn(c.Type, pkg)
}

// Enum represents a WIT [enum type], which is a [Variant] without associated data.
// The equivalent in Go is a set of const identifiers declared with iota.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [enum type]: https://component-model.bytecodealliance.org/design/wit.html#enums
type Enum struct {
	_typeDefKind
	Cases []EnumCase
}

// Despecialize despecializes [Enum] e into a [Variant] with no associated types.
// See the [canonical ABI documentation] for more information.
//
// [canonical ABI documentation]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (e *Enum) Despecialize() TypeDefKind {
	v := &Variant{
		Cases: make([]Case, len(e.Cases)),
	}
	for i := range e.Cases {
		v.Cases[i].Name = e.Cases[i].Name
		v.Cases[i].Docs = e.Cases[i].Docs
	}
	return v
}

// Size returns the [ABI byte size] for [Enum] e, the smallest integer
// type that can represent 0...len(e.Cases).
// It is first [despecialized] into a [Variant] with no associated types, then sized.
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
// [despecialized]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (e *Enum) Size() uintptr {
	return e.Despecialize().Size()
}

// Align returns the [ABI byte alignment] for [Enum] e.
// It is first [despecialized] into a [Variant] with no associated types, then aligned.
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
// [despecialized]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (e *Enum) Align() uintptr {
	return e.Despecialize().Align()
}

// Flat returns the [flattened] ABI representation of [Enum] e.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (v *Enum) Flat() []Type {
	return Discriminant(len(v.Cases)).Flat()
}

// EnumCase represents a single case in an [Enum].
// It implements the [Node] interface.
type EnumCase struct {
	Name string
	Docs Docs
}

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

// List represents a WIT [list type], which is an ordered vector of an arbitrary type.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [list type]: https://component-model.bytecodealliance.org/design/wit.html#lists
type List struct {
	_typeDefKind
	Type Type
}

// Size returns the [ABI byte size] for a [List].
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (*List) Size() uintptr { return 8 } // [2]int32

// Align returns the [ABI byte alignment] a [List].
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (*List) Align() uintptr { return 8 } // [2]int32

// Flat returns the [flattened] ABI representation of [List].
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (l *List) Flat() []Type { return []Type{PointerTo(l.Type), U32{}} }

func (l *List) dependsOn(p *Package) bool { return DependsOn(l.Type, p) }
func (*List) hasPointer() bool            { return true }
func (l *List) hasBorrow() bool           { return HasBorrow(l.Type) }
func (l *List) hasResource() bool         { return HasResource(l.Type) }

// Future represents a WIT [future type], expected to be part of [WASI Preview 3].
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [future type]: https://github.com/bytecodealliance/wit-bindgen/issues/270
// [WASI Preview 3]: https://bytecodealliance.org/articles/webassembly-the-updated-roadmap-for-developers
type Future struct {
	_typeDefKind
	Type Type // optional associated Type (can be nil)
}

// Size returns the [ABI byte size] for a [Future].
// TODO: what is the ABI size of a future?
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (*Future) Size() uintptr { return 0 }

// Align returns the [ABI byte alignment] a [Future].
// TODO: what is the ABI alignment of a future?
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (*Future) Align() uintptr { return 0 }

// Flat returns the [flattened] ABI representation of [Future].
// TODO: what is the ABI representation of a stream?
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (*Future) Flat() []Type { return nil }

func (f *Future) dependsOn(p *Package) bool { return DependsOn(f.Type, p) }
func (f *Future) hasPointer() bool          { return HasPointer(f.Type) }
func (f *Future) hasBorrow() bool           { return HasBorrow(f.Type) }
func (f *Future) hasResource() bool         { return HasResource(f.Type) }

// Stream represents a WIT [stream type], expected to be part of [WASI Preview 3].
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [stream type]: https://github.com/WebAssembly/WASI/blob/main/docs/WitInWasi.md#streams
// [WASI Preview 3]: https://bytecodealliance.org/articles/webassembly-the-updated-roadmap-for-developers
type Stream struct {
	_typeDefKind
	Element Type // optional associated Type (can be nil)
	End     Type // optional associated Type (can be nil)
}

// Size returns the [ABI byte size] for a [Stream].
// TODO: what is the ABI size of a stream?
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#size
func (*Stream) Size() uintptr { return 0 }

// Align returns the [ABI byte alignment] a [Stream].
// TODO: what is the ABI alignment of a stream?
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
func (*Stream) Align() uintptr { return 0 }

// Flat returns the [flattened] ABI representation of [Stream].
// TODO: what is the ABI representation of a stream?
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (*Stream) Flat() []Type { return nil }

func (s *Stream) dependsOn(p *Package) bool { return DependsOn(s.Element, p) || DependsOn(s.End, p) }
func (s *Stream) hasPointer() bool          { return HasPointer(s.Element) || HasPointer(s.End) }
func (s *Stream) hasBorrow() bool           { return HasBorrow(s.Element) || HasBorrow(s.End) }
func (s *Stream) hasResource() bool         { return HasResource(s.Element) || HasResource(s.End) }

// TypeOwner is the interface implemented by any type that can own a TypeDef,
// currently [World] and [Interface].
type TypeOwner interface {
	Node
	AllFunctions() iterate.Seq[*Function]
	WITPackage() *Package
	isTypeOwner()
}

type _typeOwner struct{}

func (_typeOwner) isTypeOwner() {}

// Type is the interface implemented by any type definition. This can be a
// [primitive type] or a user-defined type in a [TypeDef].
// It also conforms to the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
type Type interface {
	TypeDefKind
	TypeName() string
	isType()
}

// _type is an embeddable struct that conforms to the [Type] interface.
// It also implements the [Node], [ABI], and [TypeDefKind] interfaces.
type _type struct{ _typeDefKind }

func (_type) TypeName() string { return "" }
func (_type) isType()          {}

// ParseType parses a WIT [primitive type] string into
// the associated Type implementation from this package.
// It returns an error if the type string is not recognized.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
func ParseType(s string) (Type, error) {
	switch s {
	case "bool":
		return Bool{}, nil
	case "s8":
		return S8{}, nil
	case "u8":
		return U8{}, nil
	case "s16":
		return S16{}, nil
	case "u16":
		return U16{}, nil
	case "s32":
		return S32{}, nil
	case "u32":
		return U32{}, nil
	case "s64":
		return S64{}, nil
	case "u64":
		return U64{}, nil
	case "f32", "float32": // TODO: remove float32 at some point
		return F32{}, nil
	case "f64", "float64": // TODO: remove float64 at some point
		return F64{}, nil
	case "char":
		return Char{}, nil
	case "string":
		return String{}, nil
	}
	return nil, fmt.Errorf("unknown primitive type %q", s)
}

// primitive is a type constraint of the Go equivalents of WIT [primitive types].
//
// [primitive types]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
type primitive interface {
	bool | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64 | float32 | float64 | char | string
}

// char is defined because [rune] is an alias of [int32]
type char rune

// Primitive is the interface implemented by WIT [primitive types].
// It also conforms to the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive types]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
type Primitive interface {
	Type
	isPrimitive()
}

// _primitive is an embeddable struct that conforms to the [PrimitiveType] interface.
// It represents a WebAssembly Component Model [primitive type] mapped to its equivalent Go type.
// It also conforms to the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
type _primitive[T primitive] struct{ _type }

// isPrimitive conforms to the [Primitive] interface.
func (_primitive[T]) isPrimitive() {}

// Size returns the byte size for values of this type.
func (_primitive[T]) Size() uintptr {
	var v T
	switch any(v).(type) {
	case string:
		return 8 // [2]int32
	default:
		return unsafe.Sizeof(v)
	}
}

// Align returns the byte alignment for values of this type.
func (_primitive[T]) Align() uintptr {
	var v T
	switch any(v).(type) {
	case string:
		return 4 // int32
	default:
		return unsafe.Alignof(v)
	}
}

// hasPointer returns whether the ABI representation of this type contains a pointer.
// This will only return true for [String].
func (_primitive[T]) hasPointer() bool {
	var v T
	switch any(v).(type) {
	case string:
		return true
	default:
		return false
	}
}

// Flat returns the [flattened] ABI representation of this type.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (_primitive[T]) Flat() []Type {
	var v T
	switch any(v).(type) {
	case bool, int8, uint8, int16, uint16, int32, uint32, char:
		return []Type{U32{}}
	case int64, uint64:
		return []Type{U64{}}
	case float32:
		return []Type{F32{}}
	case float64:
		return []Type{F64{}}
	case string:
		return []Type{PointerTo(U8{}), U32{}}
	default:
		panic(fmt.Sprintf("BUG: unknown primitive type %T", v)) // should never reach here
	}
}

// Bool represents the WIT [primitive type] bool, a boolean value either true or false.
// It is equivalent to the Go type [bool].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [bool]: https://pkg.go.dev/builtin#bool
type Bool struct{ _primitive[bool] }

// S8 represents the WIT [primitive type] s8, a signed 8-bit integer.
// It is equivalent to the Go type [int8].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [int8]: https://pkg.go.dev/builtin#int8
type S8 struct{ _primitive[int8] }

// U8 represents the WIT [primitive type] u8, an unsigned 8-bit integer.
// It is equivalent to the Go type [uint8].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [uint8]: https://pkg.go.dev/builtin#uint8
type U8 struct{ _primitive[uint8] }

// S16 represents the WIT [primitive type] s16, a signed 16-bit integer.
// It is equivalent to the Go type [int16].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [int16]: https://pkg.go.dev/builtin#int16
type S16 struct{ _primitive[int16] }

// U16 represents the WIT [primitive type] u16, an unsigned 16-bit integer.
// It is equivalent to the Go type [uint16].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [uint16]: https://pkg.go.dev/builtin#uint16
type U16 struct{ _primitive[uint16] }

// S32 represents the WIT [primitive type] s32, a signed 32-bit integer.
// It is equivalent to the Go type [int32].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [int32]: https://pkg.go.dev/builtin#int32
type S32 struct{ _primitive[int32] }

// U32 represents the WIT [primitive type] u32, an unsigned 32-bit integer.
// It is equivalent to the Go type [uint32].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [uint32]: https://pkg.go.dev/builtin#uint32
type U32 struct{ _primitive[uint32] }

// S64 represents the WIT [primitive type] s64, a signed 64-bit integer.
// It is equivalent to the Go type [int64].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [int64]: https://pkg.go.dev/builtin#int64
type S64 struct{ _primitive[int64] }

// U64 represents the WIT [primitive type] u64, an unsigned 64-bit integer.
// It is equivalent to the Go type [uint64].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [uint64]: https://pkg.go.dev/builtin#uint64
type U64 struct{ _primitive[uint64] }

// F32 represents the WIT [primitive type] f32, a 32-bit floating point value.
// It is equivalent to the Go type [float32].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [float32]: https://pkg.go.dev/builtin#float32
type F32 struct{ _primitive[float32] }

// F64 represents the WIT [primitive type] f64, a 64-bit floating point value.
// It is equivalent to the Go type [float64].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [float64]: https://pkg.go.dev/builtin#float64
type F64 struct{ _primitive[float64] }

// Char represents the WIT [primitive type] char, a single Unicode character,
// specifically a [Unicode scalar value]. It is equivalent to the Go type [rune].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [Unicode scalar value]: https://unicode.org/glossary/#unicode_scalar_value
// [rune]: https://pkg.go.dev/builtin#rune
type Char struct{ _primitive[char] }

// String represents the WIT [primitive type] string, a finite string of Unicode characters.
// It is equivalent to the Go type [string].
// It implements the [Node], [ABI], [Type], and [TypeDefKind] interfaces.
//
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
// [string]: https://pkg.go.dev/builtin#string
type String struct{ _primitive[string] }

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

func (f *Function) dependsOn(pkg *Package) bool {
	for _, p := range f.Params {
		if DependsOn(p.Type, pkg) {
			return true
		}
	}
	for _, r := range f.Results {
		if DependsOn(r.Type, pkg) {
			return true
		}
	}
	return false
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

// Static represents a function that is a static method of its associated [Type].
type Static struct {
	_functionKind
	Type Type
}

// Constructor represents a function that is a constructor for its associated [Type].
type Constructor struct {
	_functionKind
	Type Type
}

// Param represents a parameter to or the result of a [Function].
// A Param can be unnamed.
type Param struct {
	Name string
	Type Type
}

// Package represents a [WIT package] within a [Resolve].
// It implements the [Node] interface.
//
// A Package is a collection of [Interface] and [World] values. Additionally,
// a Package contains a unique identifier that affects generated components and uniquely
// identifies this particular package.
//
// [WIT package]: https://component-model.bytecodealliance.org/design/wit.html#packages
type Package struct {
	Name       Ident
	Interfaces ordered.Map[string, *Interface]
	Worlds     ordered.Map[string, *World]
	Docs       Docs
}

func (p *Package) dependsOn(pkg *Package) bool {
	if pkg == p {
		return true
	}
	var done bool
	p.Interfaces.All()(func(_ string, i *Interface) bool {
		done = DependsOn(i, pkg)
		return !done
	})
	if done {
		return true
	}
	p.Worlds.All()(func(_ string, w *World) bool {
		done = DependsOn(w, pkg)
		return !done
	})
	return done
}

// DependsOn returns true if [Node] node depends on [Package] p.
// Because a package implicitly depends on itself, this returns true if node == p.
func DependsOn(node Node, p *Package) bool {
	if node == nil {
		return false
	}
	if node == p {
		return true
	}
	if k, ok := node.(TypeDefKind); ok {
		node = Despecialize(k)
	}
	if d, ok := node.(interface{ dependsOn(*Package) bool }); ok {
		return d.dependsOn(p)
	}
	return false
}

func comparePackages(a, b *Package) int {
	switch {
	case a == b:
		return 0
	case DependsOn(b, a):
		// println(b.Name.String() + " depends on " + a.Name.String())
		return 1
	case DependsOn(a, b):
		// println(a.Name.String() + " depends on " + b.Name.String())
		return -1
	}
	// println(a.Name.String() + " does not depend on " + b.Name.String())
	return strings.Compare(a.Name.String(), b.Name.String())
}

// Stability represents the version or feature-gated stability of a given feature.
type Stability interface {
	Node
	isStability()
}

// _stability is an embeddable type that conforms to the [Stability] interface.
type _stability struct{}

func (_stability) isStability() {}

// Stable represents a stable WIT feature, for example: @since(version = 1.2.3)
//
// Stable features have an explicit since version and an optional feature name.
type Stable struct {
	_stability
	Since      semver.Version
	Deprecated *semver.Version
}

// Unstable represents an unstable WIT feature defined by name.
type Unstable struct {
	_stability
	Feature    string
	Deprecated *semver.Version
}

// Docs represent WIT documentation text extracted from comments.
type Docs struct {
	Contents string // may be empty
}
