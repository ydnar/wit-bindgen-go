package wit

import (
	"fmt"
	"slices"
	"strings"
)

// Node is the common interface implemented by the WIT ([WebAssembly Interface Type])
// types in this package.
//
// [WebAssembly Interface Type]: https://component-model.bytecodealliance.org/design/wit.html
type Node interface {
	// WITKind returns the human-readable WIT kind this Node represents, e.g. "type" or "function".
	WITKind() string

	// WIT returns the WIT text format for a Node in a given context, which may be nil.
	WIT(ctx Node, name string) string
}

// Returns true if node == dep.
// Returns false if node or dep are nil.
func DependsOn(node, dep Node) bool {
	if node == nil || dep == nil {
		return false
	}
	if node == dep {
		return true
	}

	// Dereference InterfaceRefs
	if ref, ok := dep.(*InterfaceRef); ok {
		dep = ref.Interface
	}

	// TODO: is it harmful to despecialize before doing dependency check?
	// e.g. it breaks the node == dep check?
	if k, ok := node.(TypeDefKind); ok {
		node = Despecialize(k)
	}
	if d, ok := node.(dependent); ok {
		return d.dependsOn(dep)
	}

	return false
}

type dependent interface {
	dependsOn(dep Node) bool
}

// Filter returns a [Node] suitable to passing to (Resolve).WIT to filter
// the emitted WIT to a specific [World] and/or [Interface].
func Filter(w *World, i *Interface) Node {
	if w == nil && i == nil {
		return nil
	}
	return &witFilter{
		w: w,
		i: i,
	}
}

type witFilter struct {
	w *World
	i *Interface
}

func (*witFilter) WITKind() string             { panic("BUG: WITKind called on filter") }
func (*witFilter) WIT(_ Node, _ string) string { panic("BUG: WIT called on filter") }

// filter returns true if the supplied [Node] should be filtered out.
func (f *witFilter) filter(node Node) bool {
	if f == nil {
		return false // allow for nil receiver
	}
	return (f.w != nil && filterNode(f.w, node)) || (f.i != nil && filterNode(f.i, node))
}

func filterNode(target, node Node) bool {
	switch node := node.(type) {
	case *Package:
		return !DependsOn(node, target) && !DependsOn(target, node)
	case *World:
		return !DependsOn(node, target)
	}
	return !DependsOn(target, node)
}

func indent(s string) string {
	const ws = "\t"
	return strings.ReplaceAll(strings.TrimSuffix(ws+strings.ReplaceAll(s, "\n", "\n"+ws), ws), ws+"\n", "\n")
}

// unwrap unwraps a multiline string into a single line, if:
// 1. its length is <= 50 chars
// 2. its line count is <= 5
// This is used for single-line [Record], [Flags], [Variant], and [Enum] declarations.
func unwrap(s string) string {
	const chars = 50
	const lines = 5
	if len(s) > chars || strings.Count(s, "\n") > lines || strings.Contains(s, "//") {
		return s
	}
	var b strings.Builder
	for i, line := range strings.Split(s, "\n") {
		if i > 0 {
			b.WriteRune(' ')
		}
		b.WriteString(strings.Trim(line, " \t\r\n"))
	}
	line, found := strings.CutSuffix(b.String(), ", }")
	if found {
		line += " }"
	}
	return line
}

// WITKind returns the WIT kind.
func (*Resolve) WITKind() string { return "resolve" }

// WIT returns the [WIT] text format for [Resolve] r.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (r *Resolve) WIT(ctx Node, _ string) string {
	// Sort packages topologically by dependency
	packages := slices.Clone(r.Packages)
	slices.SortFunc(packages, comparePackages)
	slices.Reverse(packages)

	var b strings.Builder
	var hasContent bool
	i := 0
	for _, p := range packages {
		var name string
		if i != 0 {
			// Write subsequent packages with explicit name, which renders the package WIT with nested braces.
			name = p.Name.WIT(ctx, "")
		}
		wit := p.WIT(ctx, name)
		if ctx == nil || len(packages) == 1 || strings.Count(wit, "\n") > 1 {
			if hasContent {
				b.WriteString("\n")
			}
			hasContent = true
			b.WriteString(wit)
			i++
		}
	}
	return b.String()
}

// WITKind returns the WIT kind.
func (*Stable) WITKind() string { return "@since" }

// WIT returns the [WIT] text format for [Stable] s.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (s *Stable) WIT(_ Node, _ string) string {
	var b strings.Builder
	b.WriteString("@since(version = ")
	b.WriteString(s.Since.String())
	b.WriteRune(')')
	if s.Deprecated != nil {
		b.WriteRune('\n')
		b.WriteString("@deprecated(version = ")
		b.WriteString(s.Deprecated.String())
		b.WriteRune(')')
	}
	return b.String()
}

// WITKind returns the WIT kind.
func (*Unstable) WITKind() string { return "@unstable" }

// WIT returns the [WIT] text format for [Unstable] u.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (u *Unstable) WIT(_ Node, _ string) string {
	var b strings.Builder
	b.WriteString("@unstable(feature = ")
	b.WriteString(u.Feature)
	b.WriteRune(')')
	if u.Deprecated != nil {
		b.WriteRune('\n')
		b.WriteString("@deprecated(version = ")
		b.WriteString(u.Deprecated.String())
		b.WriteRune(')')
	}
	return b.String()
}

// WITKind returns the WIT kind.
func (*Docs) WITKind() string { return "docs" }

// WIT returns the [WIT] text format for [Docs] d.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (d *Docs) WIT(_ Node, _ string) string {
	if d.Contents == "" {
		return ""
	}
	var b strings.Builder
	lineLength := 0
	for _, c := range d.Contents {
		if lineLength == 0 {
			b.WriteString(DocPrefix)
			lineLength = len(DocPrefix)
		}
		switch c {
		case '\n':
			b.WriteRune('\n')
			lineLength = 0
			continue
		case ' ':
			switch {
			case lineLength == len(DocPrefix):
				// Ignore leading spaces
				continue
			case lineLength > LineLength:
				b.WriteRune('\n')
				lineLength = 0
				continue
			}
		default:
			if lineLength == len(DocPrefix) {
				b.WriteRune(' ')
				lineLength++
			}
		}
		b.WriteRune(c)
		lineLength++
	}
	if lineLength != 0 {
		b.WriteRune('\n')
	}
	return b.String()
}

const (
	DocPrefix  = "///"
	LineLength = 80
)

// WITKind returns the WIT kind.
func (*World) WITKind() string { return "world" }

// WIT returns the [WIT] text format for [World] w.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (w *World) WIT(ctx Node, name string) string {
	filter, _ := ctx.(*witFilter)
	if name == "" {
		name = w.Name
	}
	var b strings.Builder
	b.WriteString(w.Docs.WIT(nil, ""))
	if w.Stability != nil {
		b.WriteString(w.Stability.WIT(nil, ""))
		b.WriteRune('\n')
	}
	b.WriteString("world ")
	b.WriteString(escape(name)) // TODO: compare to w.Name?
	b.WriteString(" {")
	n := 0
	w.Imports.All()(func(name string, i WorldItem) bool {
		if filter.filter(i) {
			return true
		}
		if f, ok := i.(*Function); ok {
			if !f.IsFreestanding() {
				return true
			}
		}
		if n == 0 {
			b.WriteRune('\n')
		}
		b.WriteString(indent(i.WIT(worldImport{w}, name)))
		b.WriteRune('\n')
		n++
		return true
	})
	w.Exports.All()(func(name string, i WorldItem) bool {
		if filter.filter(i) {
			return true
		}
		if n == 0 {
			b.WriteRune('\n')
		}
		b.WriteString(indent(i.WIT(worldExport{w}, name)))
		b.WriteRune('\n')
		n++
		return true
	})
	b.WriteRune('}')
	return b.String()
}

type (
	worldImport struct{ *World }
	worldExport struct{ *World }
)

// WITKind returns the WIT kind.
func (*InterfaceRef) WITKind() string { return "interface ref" }

// WIT returns the [WIT] text format for [InterfaceRef] i.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (ref *InterfaceRef) WIT(ctx Node, name string) string {
	if ref.Stability == nil {
		return ref.Interface.WIT(ctx, name)
	}
	return ref.Stability.WIT(ctx, "") + "\n" + ref.Interface.WIT(ctx, name)
}

// WITKind returns the WIT kind.
func (*Interface) WITKind() string { return "interface" }

// WIT returns the [WIT] text format for [Interface] i.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (i *Interface) WIT(ctx Node, name string) string {
	// filter, _ := ctx.(*witFilter)

	if i.Name != nil && name == "" {
		name = *i.Name
	}

	var b strings.Builder

	switch ctx := ctx.(type) {
	case worldImport:
		rname := relativeName(i, ctx.Package)
		if rname != "" {
			return "import " + escape(rname) + ";"
		}

		// Otherwise, this is an inline interface decl.
		b.WriteString(i.Docs.WIT(ctx, ""))
		b.WriteString("import ")
		b.WriteString(escape(name))
		b.WriteString(": interface ")

	case worldExport:
		rname := relativeName(i, ctx.Package)
		if rname != "" {
			return "export " + escape(rname) + ";"
		}

		// Otherwise, this is an inline interface decl.
		b.WriteString(i.Docs.WIT(ctx, ""))
		b.WriteString("export ")
		b.WriteString(escape(name))
		b.WriteString(": interface ")

	default: // e.g. *Package
		b.WriteString(i.Docs.WIT(ctx, ""))
		if i.Stability != nil {
			b.WriteString(i.Stability.WIT(ctx, ""))
			b.WriteRune('\n')
		}
		b.WriteString("interface ")
		b.WriteString(escape(name))
		b.WriteRune(' ')
	}

	b.WriteRune('{')
	n := 0

	// Emit use statements first
	i.TypeDefs.All()(func(name string, td *TypeDef) bool {
		// if filter.filter(td) {
		// 	return true
		// }
		if td.Root().Owner == td.Owner {
			return true // Skip declarations
		}
		if n == 0 || td.Docs.Contents != "" {
			b.WriteRune('\n')
		}
		b.WriteString(indent(td.WIT(i, name)))
		b.WriteRune('\n')
		n++
		return true
	})

	// Declarations
	i.TypeDefs.All()(func(name string, td *TypeDef) bool {
		// if filter.filter(td) {
		// 	return true
		// }
		if td.Root().Owner != td.Owner {
			return true // Skip use statements
		}
		if n == 0 || td.Docs.Contents != "" {
			b.WriteRune('\n')
		}
		b.WriteString(indent(td.WIT(i, name)))
		b.WriteRune('\n')
		n++
		return true
	})

	// Functions
	i.Functions.All()(func(name string, f *Function) bool {
		// if filter.filter(f) {
		// 	return true
		// }
		if !f.IsFreestanding() {
			return true
		}
		if n == 0 || f.Docs.Contents != "" {
			b.WriteRune('\n')
		}
		b.WriteString(indent(f.WIT(i, name)))
		b.WriteRune('\n')
		n++
		return true
	})

	b.WriteRune('}')
	return b.String()
}

// WITKind returns the [WIT] kind.
func (t *TypeDef) WITKind() string {
	if alias := t.TypeDef(); alias != t {
		return "type alias"
	}
	return t.Kind.WITKind()
}

// WIT returns the [WIT] text format for [TypeDef] t.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (t *TypeDef) WIT(ctx Node, name string) string {
	if t.Name != nil && name == "" {
		name = *t.Name
	}
	switch ctx := ctx.(type) {
	// If context is another TypeDef, then this is an imported type.
	case *TypeDef:
		// Emit an type alias if same Owner.
		if t.Owner == ctx.Owner && t.Name != nil {
			return "type " + escape(name) + " = " + escape(*t.Name)
		}
		ownerName := relativeName(t.Owner, ctx.Owner.WITPackage())
		if t.Name != nil && *t.Name != name {
			return fmt.Sprintf("use %s.{%s as %s};", ownerName, escape(*t.Name), escape(name))
		}
		return fmt.Sprintf("use %s.{%s};", ownerName, escape(name))

	case worldImport, worldExport, *Interface:
		var b strings.Builder
		b.WriteString(t.Docs.WIT(ctx, ""))
		if t.Stability != nil {
			b.WriteString(t.Stability.WIT(ctx, ""))
			b.WriteRune('\n')
		}
		b.WriteString(t.Kind.WIT(t, name))
		constructor := t.Constructor()
		methods := t.Methods()
		statics := t.StaticFunctions()
		if constructor != nil || len(methods) > 0 || len(statics) > 0 {
			b.WriteString(" {\n")
			n := 0
			if constructor != nil {
				b.WriteString(indent(constructor.WIT(t, "constructor")))
				b.WriteRune('\n')
				n++
			}
			slices.SortFunc(methods, compareFunctions)
			for _, f := range methods {
				if f.Docs.Contents != "" {
					b.WriteRune('\n')
				}
				b.WriteString(indent(f.WIT(t, "")))
				b.WriteRune('\n')
				n++
			}
			slices.SortFunc(statics, compareFunctions)
			for _, f := range statics {
				if f.Docs.Contents != "" {
					b.WriteRune('\n')
				}
				b.WriteString(indent(f.WIT(t, "")))
				b.WriteRune('\n')
				n++
			}
			b.WriteRune('}')
		}
		s := b.String()
		if s[len(s)-1] != '}' && s[len(s)-1] != ';' {
			b.WriteRune(';')
		}
		return b.String()
	}
	if name != "" {
		return escape(name)
	}
	return t.Kind.WIT(ctx, name)
}

func escape(name string) string {
	if witKeywords[name] {
		return "%" + name
	}
	return name
}

// A map of all [WIT keywords].
//
// [WIT keywords]: https://github.com/bytecodealliance/wasm-tools/blob/main/crates/wit-parser/src/ast/lex.rs#L528-L578
var witKeywords = map[string]bool{
	"as":            true,
	"bool":          true,
	"borrow":        true,
	"char":          true,
	"constructor":   true,
	"enum":          true,
	"error-context": true,
	"export":        true,
	"f32":           true,
	"f64":           true,
	"flags":         true,
	"from":          true,
	"func":          true,
	"future":        true,
	"import":        true,
	"include":       true,
	"interface":     true,
	"list":          true,
	"option":        true,
	"own":           true,
	"package":       true,
	"record":        true,
	"resource":      true,
	"result":        true,
	"s16":           true,
	"s32":           true,
	"s64":           true,
	"s8":            true,
	"static":        true,
	"stream":        true,
	"string":        true,
	"tuple":         true,
	"type":          true,
	"u16":           true,
	"u32":           true,
	"u64":           true,
	"u8":            true,
	"use":           true,
	"variant":       true,
	"wit":           true,
	"world":         true,
}

func relativeName(o TypeOwner, p *Package) string {
	var op *Package
	var name string
	switch o := o.(type) {
	case *Interface:
		if o.Name == nil {
			return ""
		}
		op = o.Package
		name = *o.Name

	case *World:
		op = o.Package
		name = o.Name
	}
	if op == p {
		return name
	}
	if op == nil {
		return ""
	}
	qualifiedName := op.Name
	qualifiedName.Package += "/" + escape(name)
	return qualifiedName.String()
}

// WITKind returns the WIT kind.
func (*Pointer) WITKind() string { return "pointer" }

// WIT returns the [WIT] text format for [Pointer] p.
// Note: there is no canonical WIT representation of a pointer.
// This method exists solely to implement the [Node] interface.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (p *Pointer) WIT(_ Node, name string) string {
	var b strings.Builder
	if name != "" {
		b.WriteString("type ")
		b.WriteString(escape(name))
		b.WriteString(" = ")
	}
	b.WriteString("pointer<")
	b.WriteString(p.Type.WIT(p, ""))
	b.WriteRune('>')
	return b.String()
}

// WITKind returns the WIT kind.
func (*Record) WITKind() string { return "record" }

// WIT returns the [WIT] text format for [Record] r.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (r *Record) WIT(ctx Node, name string) string {
	var b strings.Builder
	b.WriteString("record ")
	b.WriteString(escape(name))
	b.WriteString(" {")
	if len(r.Fields) > 0 {
		b.WriteRune('\n')
		for i := range r.Fields {
			b.WriteString(indent(r.Fields[i].WIT(ctx, "")))
			b.WriteString(",\n")
		}
	}
	b.WriteRune('}')
	if ctx == nil {
		return b.String()
	}
	return unwrap(b.String())
}

// WITKind returns the WIT kind.
func (*Field) WITKind() string { return "field" }

// WIT returns the [WIT] text format for [Field] f.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (f *Field) WIT(ctx Node, name string) string {
	if ctx == nil {
		// Omit docs
		return escape(f.Name) + ": " + f.Type.WIT(f, "")
	}
	return f.Docs.WIT(ctx, "") + escape(f.Name) + ": " + f.Type.WIT(f, "")
}

// WITKind returns the WIT kind.
func (*Resource) WITKind() string { return "resource" }

// WIT returns the [WIT] text format for [Resource] r.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (r *Resource) WIT(ctx Node, name string) string {
	return "resource " + escape(name)
}

// WITKind returns the WIT kind.
func (*Own) WITKind() string { return "own" }

// WIT returns the [WIT] text format for [Own] h.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (o *Own) WIT(ctx Node, name string) string {
	var b strings.Builder
	if name != "" {
		b.WriteString("type ")
		b.WriteString(escape(name))
		b.WriteString(" = ")
		b.WriteString("own<")
	}
	b.WriteString(o.Type.WIT(o, ""))
	if name != "" {
		b.WriteRune('>')
	}
	return b.String()
}

// WITKind returns the WIT kind.
func (*Borrow) WITKind() string { return "borrow" }

// WIT returns the [WIT] text format for [Borrow] h.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (b *Borrow) WIT(ctx Node, name string) string {
	var s strings.Builder
	if name != "" {
		s.WriteString("type ")
		s.WriteString(escape(name))
		s.WriteString(" = ")
	}
	s.WriteString("borrow<")
	s.WriteString(b.Type.WIT(b, ""))
	s.WriteRune('>')
	return s.String()
}

// WITKind returns the WIT kind.
func (ErrorContext) WITKind() string { return "error-context" }

// WIT returns the [WIT] text format for [ErrorContext] r.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (ErrorContext) WIT(_ Node, name string) string {
	var b strings.Builder
	if name != "" {
		b.WriteString("type ")
		b.WriteString(escape(name))
		b.WriteString(" = ")
	}
	b.WriteString("error-context")
	return b.String()
}

// WITKind returns the WIT kind.
func (*Flags) WITKind() string { return "flags" }

// WIT returns the [WIT] text format for [Flags] f.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (f *Flags) WIT(ctx Node, name string) string {
	var b strings.Builder
	b.WriteString("flags ")
	b.WriteString(escape(name))
	b.WriteString(" {\n")
	if len(f.Flags) > 0 {
		for i := range f.Flags {
			b.WriteString(indent(f.Flags[i].WIT(ctx, "")))
			b.WriteString(",\n")
		}
	}
	b.WriteRune('}')
	if ctx == nil {
		return b.String()
	}
	return unwrap(b.String())
}

// WITKind returns the WIT kind.
func (*Flag) WITKind() string { return "flag" }

// WIT returns the [WIT] text format for [Flag] f.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (f *Flag) WIT(ctx Node, _ string) string {
	if ctx == nil {
		// Omit docs
		return escape(f.Name)
	}
	return f.Docs.WIT(ctx, "") + escape(f.Name)
}

// WITKind returns the WIT kind.
func (*Tuple) WITKind() string { return "tuple" }

// WIT returns the [WIT] text format for [Tuple] t.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (t *Tuple) WIT(ctx Node, name string) string {
	var b strings.Builder
	if name != "" {
		b.WriteString("type ")
		b.WriteString(escape(name))
		b.WriteString(" = ")
	}
	b.WriteString("tuple<")
	for i := range t.Types {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(t.Types[i].WIT(t, ""))
	}
	b.WriteString(">")
	return b.String()
}

// WITKind returns the WIT kind.
func (*Variant) WITKind() string { return "variant" }

// WIT returns the [WIT] text format for [Variant] v.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (v *Variant) WIT(ctx Node, name string) string {
	var b strings.Builder
	b.WriteString("variant ")
	b.WriteString(escape(name))
	b.WriteString(" {")
	if len(v.Cases) > 0 {
		b.WriteRune('\n')
		for i := range v.Cases {
			b.WriteString(indent(v.Cases[i].WIT(ctx, "")))
			b.WriteString(",\n")
		}
	}
	b.WriteRune('}')
	if ctx == nil {
		return b.String()
	}
	return unwrap(b.String())
}

// WITKind returns the WIT kind.
func (*Case) WITKind() string { return "case" }

// WIT returns the [WIT] text format for [Variant] [Case] c.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (c *Case) WIT(ctx Node, _ string) string {
	var b strings.Builder
	if ctx != nil {
		b.WriteString(c.Docs.WIT(ctx, ""))
	}
	b.WriteString(escape(c.Name))
	if c.Type != nil {
		b.WriteRune('(')
		b.WriteString(c.Type.WIT(c, ""))
		b.WriteRune(')')
	}
	return b.String()
}

// WITKind returns the WIT kind.
func (*Enum) WITKind() string { return "enum" }

// WIT returns the [WIT] text format for [Enum] e.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (e *Enum) WIT(ctx Node, name string) string {
	var b strings.Builder
	b.WriteString("enum ")
	b.WriteString(escape(name))
	b.WriteString(" {")
	if len(e.Cases) > 0 {
		b.WriteRune('\n')
		for i := range e.Cases {
			if i > 0 {
				b.WriteString(",\n")
			}
			b.WriteString(indent(e.Cases[i].WIT(ctx, "")))
		}
		b.WriteRune('\n')
	}
	b.WriteRune('}')
	if ctx == nil {
		return b.String()
	}
	return unwrap(b.String())
}

// WITKind returns the WIT kind.
func (*EnumCase) WITKind() string { return "enum-case" }

// WIT returns the [WIT] text format for [EnumCase] c.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (c *EnumCase) WIT(ctx Node, _ string) string {
	if ctx == nil {
		// Omit docs
		return escape(c.Name)
	}
	return c.Docs.WIT(ctx, "") + escape(c.Name)
}

// WITKind returns the WIT kind.
func (*Option) WITKind() string { return "option" }

// WIT returns the [WIT] text format for [Option] o.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (o *Option) WIT(_ Node, name string) string {
	var b strings.Builder
	if name != "" {
		b.WriteString("type ")
		b.WriteString(escape(name))
		b.WriteString(" = ")
	}
	b.WriteString("option<")
	b.WriteString(o.Type.WIT(o, ""))
	b.WriteRune('>')
	return b.String()
}

// WITKind returns the WIT kind.
func (*Result) WITKind() string { return "result" }

// WIT returns the [WIT] text format for [Result] r.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (r *Result) WIT(_ Node, name string) string {
	var b strings.Builder
	if name != "" {
		b.WriteString("type ")
		b.WriteString(escape(name))
		b.WriteString(" = ")
	}
	b.WriteString("result")
	if r.OK == nil && r.Err == nil {
		return b.String()
	}
	b.WriteRune('<')
	if r.OK != nil {
		b.WriteString(r.OK.WIT(r, ""))
	} else {
		b.WriteRune('_')
	}
	if r.Err != nil {
		b.WriteString(", ")
		b.WriteString(r.Err.WIT(r, ""))
	}
	b.WriteRune('>')
	return b.String()
}

// WITKind returns the WIT kind.
func (*List) WITKind() string { return "list" }

// WIT returns the [WIT] text format for [List] l.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (l *List) WIT(_ Node, name string) string {
	var b strings.Builder
	if name != "" {
		b.WriteString("type ")
		b.WriteString(escape(name))
		b.WriteString(" = ")
	}
	b.WriteString("list<")
	b.WriteString(l.Type.WIT(l, ""))
	b.WriteRune('>')
	return b.String()
}

// WITKind returns the WIT kind.
func (*Future) WITKind() string { return "future" }

// WIT returns the [WIT] text format for [Future] f.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (f *Future) WIT(_ Node, name string) string {
	var b strings.Builder
	if name != "" {
		b.WriteString("type ")
		b.WriteString(escape(name))
		b.WriteString(" = ")
	}
	b.WriteString("future")
	if f.Type != nil {
		b.WriteRune('<')
		b.WriteString(f.Type.WIT(f, ""))
		b.WriteRune('>')
	}
	return b.String()
}

// WITKind returns the WIT kind.
func (*Stream) WITKind() string { return "stream" }

// WIT returns the [WIT] text format for [Stream] s.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (s *Stream) WIT(_ Node, name string) string {
	var b strings.Builder
	if name != "" {
		b.WriteString("type ")
		b.WriteString(escape(name))
		b.WriteString(" = ")
	}
	b.WriteString("stream")
	if s.Type != nil {
		b.WriteRune('<')
		b.WriteString(s.Type.WIT(s, ""))
		b.WriteRune('>')
	}
	return b.String()
}

// WITKind returns the canonical [primitive type] kind in [WIT] text format.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
// [primitive type]: https://component-model.bytecodealliance.org/design/wit.html#primitive-types
func (_primitive[T]) WITKind() string {
	var v T
	switch any(v).(type) {
	case bool:
		return "bool"
	case int8:
		return "s8"
	case uint8:
		return "u8"
	case int16:
		return "s16"
	case uint16:
		return "u16"
	case int32:
		return "s32"
	case uint32:
		return "u32"
	case int64:
		return "s64"
	case uint64:
		return "u64"
	case float32:
		return "f32"
	case float64:
		return "f64"
	case char:
		return "char"
	case string:
		return "string"
	case errorContext:
		return "error-context"
	default:
		panic(fmt.Sprintf("BUG: unknown primitive type %T", v)) // should never reach here
	}
}

// WIT returns the [WIT] text format for this [_primitive].
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (p _primitive[T]) WIT(_ Node, name string) string {
	if name != "" {
		return "type " + name + " = " + p.WITKind()
	}
	return p.WITKind()
}

// WITKind returns the WIT kind for [Function] f.
func (f *Function) WITKind() string {
	switch f.Kind.(type) {
	case *Freestanding:
		return "function"
	case *Constructor:
		return "constructor"
	case *Static:
		return "static function"
	case *Method:
		return "method"
	default:
		panic("unknown function type for func " + f.Name)
	}
}

// WIT returns the [WIT] text format for [Function] f.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (f *Function) WIT(ctx Node, name string) string {
	if name == "" {
		name = f.Name
		if _, after, found := strings.Cut(name, "."); found {
			name = after
		}
	}
	var b strings.Builder
	if ctx != nil {
		b.WriteString(f.Docs.WIT(ctx, ""))
		if f.Stability != nil {
			b.WriteString(f.Stability.WIT(ctx, ""))
			b.WriteRune('\n')
		}
	}
	switch ctx.(type) {
	case worldImport:
		b.WriteString("import ")
	case worldExport:
		b.WriteString("export ")
	}
	var isConstructor, isMethod bool
	switch f.Kind.(type) {
	case *Constructor:
		// constructor is a keyword in WIT, but should not be escaped as a function name
		b.WriteString(name)
		b.WriteRune('(')
		isConstructor = true
	case *Freestanding, *Method:
		b.WriteString(escape(name))
		b.WriteString(": func(")
		isMethod = true
	case *Static:
		b.WriteString(escape(name))
		b.WriteString(": static func(")
	}
	b.WriteString(paramsWIT(f.Params, isMethod))
	b.WriteRune(')')
	if !isConstructor && len(f.Results) > 0 {
		parens := len(f.Results) > 1 || f.Results[0].Name != ""
		b.WriteString(" -> ")
		if parens {
			b.WriteRune('(')
		}
		b.WriteString(paramsWIT(f.Results, false))
		if parens {
			b.WriteRune(')')
		}
	}
	b.WriteRune(';')
	return b.String()
}

func paramsWIT(params []Param, isMethod bool) string {
	var b strings.Builder
	var i int
	for _, param := range params {
		if param.Name == "self" && isMethod {
			continue
		}
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(param.WIT(nil, ""))
		i++
	}
	return b.String()
}

// WITKind returns the WIT kind.
func (*Param) WITKind() string { return "param" }

// WIT returns the [WIT] text format of [Param] p.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (p *Param) WIT(_ Node, _ string) string {
	if p.Name == "" {
		return p.Type.WIT(p, "")
	}
	return escape(p.Name) + ": " + p.Type.WIT(p, "")
}

// WITKind returns the WIT kind.
func (*Package) WITKind() string { return "package" }

// WIT returns the [WIT] text format of [Package] p.
// Specify name to render braced, multi-package form.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (p *Package) WIT(ctx Node, name string) string {
	filter, _ := ctx.(*witFilter)
	multi := name != ""
	var b strings.Builder
	b.WriteString(p.Docs.WIT(ctx, ""))
	b.WriteString("package ")
	b.WriteString(p.Name.WIT(p, ""))
	if multi {
		b.WriteString(" {")
	} else {
		b.WriteString(";\n")
	}
	i := 0
	p.Interfaces.All()(func(name string, face *Interface) bool {
		if filter.filter(face) {
			return true
		}
		b.WriteRune('\n')
		if multi {
			b.WriteString(indent(face.WIT(ctx, name)))
		} else {
			b.WriteString(face.WIT(ctx, name))
		}
		b.WriteRune('\n')
		i++
		return true
	})
	p.Worlds.All()(func(name string, w *World) bool {
		if filter.filter(w) {
			return true
		}
		b.WriteRune('\n')
		if multi {
			b.WriteString(indent(w.WIT(ctx, name)))
		} else {
			b.WriteString(w.WIT(ctx, name))
		}
		b.WriteRune('\n')
		i++
		return true
	})
	if multi {
		b.WriteString("}\n")
	}
	// Return empty string if package has no contents
	// if i == 0 {
	// 	return ""
	// }
	return b.String()
}

// WITKind returns the WIT kind.
func (id *Ident) WITKind() string { return "ident" }

// WIT returns the [WIT] text format of [Ident] id.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func (id *Ident) WIT(_ Node, _ string) string {
	ide := Ident{
		Namespace: escape(id.Namespace),
		Package:   escape(id.Package),
		Extension: escape(id.Extension),
		Version:   id.Version,
	}
	return ide.String()
}
