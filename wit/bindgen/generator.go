// Package bindgen generates Go source code from a fully-resolved WIT package.
// It generates one or more Go packages, with functions, types, constants, and variables,
// along with the associated code to lift and lower Go types into Canonical ABI representation.
package bindgen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/token"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing/fstest"
	"time"

	"github.com/tetratelabs/wazero/sys"
	"go.bytecodealliance.org/cm"
	"go.bytecodealliance.org/internal/codec"
	"go.bytecodealliance.org/internal/go/gen"
	"go.bytecodealliance.org/internal/stringio"
	"go.bytecodealliance.org/internal/wasm"
	"go.bytecodealliance.org/internal/wasmtools"
	"go.bytecodealliance.org/wit"
	"go.bytecodealliance.org/wit/logging"
)

const (
	cmPackage = "go.bytecodealliance.org/cm"
	emptyAsm  = `// This file exists for testing this package without WebAssembly,
// allowing empty function bodies with a //go:wasmimport directive.
// See https://pkg.go.dev/cmd/compile for more information.
`

	// Predeclare Go types for own<T> and borrow<T>.
	// Currently broken.
	experimentPredeclareHandles = false

	// Define Go GC shape types for variant and result storage.
	experimentCreateShapeTypes = true
)

type typeDecl struct {
	file  *gen.File // The Go file this type belongs to
	scope gen.Scope // Scope for type-local declarations like method names
	name  string    // The unique Go name for this type
}

type funcDecl struct {
	owner      wit.TypeOwner
	dir        wit.Direction
	f          *wit.Function
	goFunc     function // The Go function
	wasmFunc   function // The wasmimport or wasmexport function
	linkerName string   // The wasmimport or wasmexport mangled linker name
}

// function represents a Go function created from a Component Model function
type function struct {
	file     *gen.File // The Go file this function belongs to
	scope    gen.Scope // Scope for function-local declarations
	name     string    // The scoped unique Go name for this function (method names are scoped to receiver type)
	receiver param     // The method receiver, if any
	params   []param   // Function param(s), with unique Go name(s)
	results  []param   // Function result(s), with unique Go name(s)
}

func (f *function) isMethod() bool {
	return f.receiver.typ != nil
}

// param represents a Go function parameter or result.
// name is a unique Go name within the function scope.
type param struct {
	name string
	typ  wit.Type
	dir  wit.Direction
}

type typeUse struct {
	pkg *gen.Package
	dir wit.Direction
	typ *wit.TypeDef
}

type generator struct {
	opts  options
	res   *wit.Resolve
	world *wit.World // optional, can be nil

	// versioned is set to true if there are multiple versions of a WIT package in res,
	// which affects the generated Go package paths.
	versioned bool

	// packages are Go packages indexed on Go package paths.
	packages map[string]*gen.Package

	// witPackages map wit.TypeOwner (World, Interface) to Go packages.
	witPackages map[wit.TypeOwner]*gen.Package

	// exportScopes map wit.TypeOwner to export scopes.
	exportScopes map[wit.TypeOwner]gen.Scope

	// moduleNames map wit.TypeOwner to the wasmimport/wasmexport module names.
	moduleNames map[wit.TypeOwner]string

	// types map wit.TypeDef to their Go equivalent.
	// It is indexed on wit.Direction, either Imported or Exported.
	types [2]map[*wit.TypeDef]*typeDecl

	// functions map wit.Function to their Go equivalent.
	// It is indexed on wit.Direction, either Imported or Exported.
	functions [2]map[*wit.Function]*funcDecl

	// defined represent whether a world, interface, type, or function has been defined.
	// It is indexed on wit.Direction, either Imported or Exported.
	defined [2]map[wit.Node]bool

	// ABI shapes for any type, use for variant and result Shape type parameters.
	shapes map[typeUse]string

	// lowering and lifting functions for defined types.
	lowerFunctions map[typeUse]function
	liftFunctions  map[typeUse]function

	wasmTools *wasmtools.Instance
}

func newGenerator(res *wit.Resolve, opts ...Option) (*generator, error) {
	g := &generator{
		packages:       make(map[string]*gen.Package),
		witPackages:    make(map[wit.TypeOwner]*gen.Package),
		exportScopes:   make(map[wit.TypeOwner]gen.Scope),
		moduleNames:    make(map[wit.TypeOwner]string),
		shapes:         make(map[typeUse]string),
		lowerFunctions: make(map[typeUse]function),
		liftFunctions:  make(map[typeUse]function),
	}
	for i := 0; i < 2; i++ {
		g.types[i] = make(map[*wit.TypeDef]*typeDecl)
		g.functions[i] = make(map[*wit.Function]*funcDecl)
		g.defined[i] = make(map[wit.Node]bool)
	}
	err := g.opts.apply(opts...)
	if err != nil {
		return nil, err
	}
	if g.opts.logger == nil {
		g.opts.logger = logging.NewLogger(io.Discard, logging.LevelNever)
	}
	if g.opts.generatedBy == "" {
		_, file, _, _ := runtime.Caller(0)
		_, g.opts.generatedBy = filepath.Split(filepath.Dir(filepath.Dir(file)))
	}
	if g.opts.cmPackage == "" {
		g.opts.cmPackage = cmPackage
	}
	g.res = res
	for _, g.world = range res.Worlds {
		if g.world.Match(g.opts.world) {
			break
		}
		// otherwise chose the last world
	}
	g.wasmTools, err = wasmtools.New(context.Background())
	if err != nil {
		return nil, err
	}
	return g, nil
}

// TODO: factor this out
func findWorld(r *wit.Resolve, pattern string) *wit.World {
	for _, w := range r.Worlds {
		if w.Match(pattern) {
			return w
		}
	}
	return nil
}

func (g *generator) generate() ([]*gen.Package, error) {
	g.detectVersionedPackages()
	err := g.defineWorlds()
	if err != nil {
		return nil, err
	}
	var packages []*gen.Package
	for _, path := range codec.SortedKeys(g.packages) {
		packages = append(packages, g.packages[path])
	}
	return packages, nil
}

func (g *generator) detectVersionedPackages() {
	if g.opts.versioned {
		g.versioned = true
		g.opts.logger.Infof("Generated versions for all package(s)\n")
		return
	}
	packages := make(map[string]string)
	for _, pkg := range g.res.Packages {
		id := pkg.Name
		id.Version = nil
		path := id.String()
		if packages[path] != "" && packages[path] != pkg.Name.String() {
			g.versioned = true
		} else {
			packages[path] = pkg.Name.String()
		}
	}
	if g.versioned {
		g.opts.logger.Infof("Multiple versions of package(s) detected\n")
	}
}

// define marks a world, interface, type, or function as defined.
// It returns true if was newly defined.
func (g *generator) define(dir wit.Direction, v wit.Node) (defined bool) {
	if g.defined[dir][v] {
		return false
	}
	g.defined[dir][v] = true
	return true
}

// By default, each WIT interface and world maps to a single Go package.
// Options might override the Go package, including combining multiple
// WIT interfaces and/or worlds into a single Go package.
func (g *generator) defineWorlds() error {
	g.opts.logger.Infof("Generating Go for %d world(s)\n", len(g.res.Worlds))
	for _, w := range g.res.Worlds {
		if w == g.world || g.world == nil {
			err := g.defineWorld(w)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generator) defineWorld(w *wit.World) error {
	if !g.define(wit.Exported, w) {
		return nil
	}
	id := w.Package.Name
	id.Extension = w.Name

	g.moduleNames[w] = id.String()

	pkg, err := g.newPackage(w, nil, "")
	if err != nil {
		return err
	}

	// Write WIT file for this world
	// witFile := g.witFileFor(w)
	// witFile.WriteString(g.res.WIT(wit.Filter(w, nil), ""))

	// Write Go package docs
	file := g.fileFor(w)
	var b strings.Builder
	stringio.Write(&b, "Package ", pkg.Name, " represents the ", w.WITKind(), " \"", g.moduleNames[w], "\".\n")
	if w.Docs.Contents != "" {
		b.WriteString("\n")
		b.WriteString(w.Docs.Contents)
	}
	file.PackageDocs = b.String()

	w.Imports.All()(func(name string, v wit.WorldItem) bool {
		switch v := v.(type) {
		case *wit.InterfaceRef:
			// TODO: handle Stability
			err = g.defineInterface(w, wit.Imported, v.Interface, name)
		case *wit.TypeDef:
			err = g.defineTypeDef(wit.Imported, v, name)
		case *wit.Function:
			if v.IsFreestanding() {
				err = g.defineFunction(w, wit.Imported, v)
			}
		}
		return err == nil
	})
	if err != nil {
		return err
	}

	w.Exports.All()(func(name string, v wit.WorldItem) bool {
		switch v := v.(type) {
		case *wit.InterfaceRef:
			// TODO: handle Stability
			err = g.defineInterface(w, wit.Exported, v.Interface, name)
		case *wit.TypeDef:
			// WIT does not currently allow worlds to export types.
			err = errors.New("exported type in world " + w.Name)
		case *wit.Function:
			if v.IsFreestanding() {
				err = g.defineFunction(w, wit.Exported, v)
			}
		}
		return err == nil
	})

	return err
}

func (g *generator) defineInterface(w *wit.World, dir wit.Direction, i *wit.Interface, name string) error {
	if !g.define(dir, i) {
		return nil
	}

	if i.Name == nil {
		g.moduleNames[i] = name
	} else {
		name = *i.Name
		id := i.Package.Name
		id.Extension = name
		g.moduleNames[i] = id.String()
	}

	pkg, err := g.newPackage(w, i, name)
	if err != nil {
		return err
	}

	// Write WIT file for this interface
	// Disabled until we write metadata .wasm file
	// witFile := g.witFileFor(i)
	// witFile.WriteString(g.res.WIT(wit.Filter(g.world, i), ""))

	file := g.fileFor(i)
	{
		var b strings.Builder
		stringio.Write(&b, "Package ", pkg.Name, " represents the ", dir.String(), " ", i.WITKind(), " \"", g.moduleNames[i], "\".\n")
		if i.Docs.Contents != "" {
			b.WriteString("\n")
			b.WriteString(i.Docs.Contents)
		}
		file.PackageDocs = b.String()
	}

	// Declare types
	i.TypeDefs.All()(func(name string, td *wit.TypeDef) bool {
		g.declareTypeDef(nil, dir, td, "")
		return true
	})

	// Define types
	i.TypeDefs.All()(func(name string, td *wit.TypeDef) bool {
		g.defineTypeDef(dir, td, name)
		return true
	})

	// TODO: delete this
	// Declare all functions
	// i.Functions.All()(func(_ string, f *wit.Function) bool {
	// 	g.declareFunction(id, dir, f)
	// 	return true
	// })

	// Define standalone functions
	i.Functions.All()(func(_ string, f *wit.Function) bool {
		if f.IsFreestanding() {
			g.defineFunction(i, dir, f)
		}
		return true
	})

	return nil
}

func (g *generator) defineTypeDef(dir wit.Direction, t *wit.TypeDef, name string) error {
	if !g.define(dir, t) {
		return nil
	}
	if t.Name != nil {
		name = *t.Name
	}

	decl, err := g.declareTypeDef(nil, dir, t, "")
	if err != nil {
		return err
	}

	// Define the type
	var b bytes.Buffer
	stringio.Write(&b, "// ", decl.name, " represents the ")
	if wit.HasResource(t) {
		stringio.Write(&b, dir.String(), " ")
	}
	stringio.Write(&b, t.WITKind(), " \"", g.moduleNames[t.Owner], "#", name, "\".\n")
	b.WriteString("//\n")
	parent := t.TypeDef()
	if parent != t {
		// Type alias
		stringio.Write(&b, "// See [", g.typeRep(decl.file, dir, parent), "] for more information.\n")
		stringio.Write(&b, "type ", decl.name, " = ", g.typeRep(decl.file, dir, parent), "\n\n")
	} else {
		b.WriteString(formatDocComments(t.Docs.Contents, false))
		b.WriteString("//\n")
		b.WriteString(formatDocComments(t.Kind.WIT(nil, t.TypeName()), true))
		stringio.Write(&b, "type ", decl.name, " ", g.typeDefRep(decl.file, dir, t, decl.name), "\n\n")
	}

	_, err = decl.file.Write(b.Bytes())
	if err != nil {
		return err
	}

	// Return now unless the type is a resource.
	if _, ok := t.Kind.(*wit.Resource); !ok {
		return nil
	}

	// Emit type namespace in exports file.
	if dir == wit.Exported {
		exportsFile := g.exportsFileFor(t.Owner)
		scope := g.exportScopes[t.Owner]
		goName := scope.GetName(GoName(*t.Name, true))
		stringio.Write(exportsFile, "\n// ", goName, " represents the caller-defined exports for ", t.WITKind(), " \"", g.moduleNames[t.Owner], "#", name, "\".\n")
		stringio.Write(exportsFile, goName, " struct {")
	}

	// Define any associated functions
	switch dir {
	case wit.Imported:
		if f := t.ResourceDrop(); f != nil {
			err := g.defineFunction(t.Owner, wit.Imported, f)
			if err != nil {
				return nil
			}
		}

	case wit.Exported:
		if f := t.ResourceNew(); f != nil {
			err := g.defineFunction(t.Owner, importedWithExportedTypes, f)
			if err != nil {
				return nil
			}
		}

		if f := t.ResourceRep(); f != nil {
			err := g.defineFunction(t.Owner, importedWithExportedTypes, f)
			if err != nil {
				return nil
			}
		}

		if f := t.ResourceDrop(); f != nil {
			err := g.defineFunction(t.Owner, importedWithExportedTypes, f)
			if err != nil {
				return nil
			}
		}

		if f := t.Destructor(); f != nil {
			err := g.defineFunction(t.Owner, dir, f)
			if err != nil {
				return nil
			}
		}

	default:
		return errors.New("BUG: unknown direction " + dir.String())
	}

	if f := t.Constructor(); f != nil {
		err := g.defineFunction(t.Owner, dir, f)
		if err != nil {
			return nil
		}
	}

	for _, f := range t.StaticFunctions() {
		err := g.defineFunction(t.Owner, dir, f)
		if err != nil {
			return nil
		}
	}

	for _, f := range t.Methods() {
		err := g.defineFunction(t.Owner, dir, f)
		if err != nil {
			return nil
		}
	}

	// End struct definition here.
	if dir == wit.Exported {
		exportsFile := g.exportsFileFor(t.Owner)
		stringio.Write(exportsFile, "\n}\n")
	}

	return nil
}

func (g *generator) declareTypeDef(file *gen.File, dir wit.Direction, t *wit.TypeDef, goName string) (*typeDecl, error) {
	decl, ok := g.types[dir][t]
	if ok {
		return decl, nil
	}
	if goName == "" {
		if t.Name == nil {
			return nil, errors.New("BUG: cannot declare unnamed wit.TypeDef")
		}
		goName = GoName(*t.Name, true)
	}
	if file == nil {
		file = g.fileFor(t.Owner)
	}
	decl = &typeDecl{
		file:  file,
		name:  declareDirectedName(file, dir, goName),
		scope: gen.NewScope(nil),
	}
	g.types[dir][t] = decl

	// Declare the export scope for this type.
	if dir == wit.Exported && g.exportScopes[t.Owner] != nil {
		g.exportScopes[t.Owner].DeclareName(goName)
	}

	// If an imported and exported version of a TypeDef are identical, declare the other.
	otherDir := ^dir & 1
	if _, ok := g.types[otherDir][t]; !ok && !wit.HasResource(t) {
		g.types[otherDir][t] = decl
		g.define(otherDir, t) // Mark this type as defined
	}

	{
		name := "(anonymous)"
		if t.Name != nil {
			name = *t.Name
		}
		g.opts.logger.Debugf("Type:\t%s.%s\n\t%s.%s\n",
			g.moduleNames[t.Owner], name, file.Package.Path, decl.name)
	}

	// Predeclare own<T> and borrow<T> for resource types.
	if experimentPredeclareHandles {
		switch t.Kind.(type) {
		case *wit.Resource:
			var count int
			for _, t2 := range g.res.TypeDefs {
				var err error
				switch kind := t2.Kind.(type) {
				case *wit.Own:
					if kind.Type == t {
						_, err = g.declareTypeDef(file, dir, t2, "Own"+decl.name)
						count++
					}
				case *wit.Borrow:
					if kind.Type == t {
						_, err = g.declareTypeDef(file, dir, t2, "Borrow"+decl.name)
						count++
					}
				}
				if err != nil {
					return nil, err
				}
				if count >= 2 {
					break
				}
			}
		}
	}

	// Predeclare reserved methods.
	switch t.Kind.(type) {
	case *wit.Enum:
		decl.scope.DeclareName("String")        // For fmt.Stringer
		decl.scope.DeclareName("MarshalText")   // For encoding.TextMarshaler
		decl.scope.DeclareName("UnmarshalText") // For encoding.TextUnmarshaler
	case *wit.Variant:
		decl.scope.DeclareName("Tag")    // Method on cm.Variant
		decl.scope.DeclareName("String") // For fmt.Stringer
	}

	return decl, nil
}

func declareDirectedName(scope gen.Scope, dir wit.Direction, name string) string {
	if dir == wit.Exported && scope.HasName(name) {
		if token.IsExported(name) {
			// Go exported, not WIT exported!
			return scope.DeclareName("Export" + name)
		}
		return scope.DeclareName("export" + name)
	}
	return scope.DeclareName(name)
}

// typeDecl returns the typeDecl for [wit.Direction] dir and [wit.TypeDef] t, and whether it was declared.
func (g *generator) typeDecl(dir wit.Direction, t *wit.TypeDef) (decl *typeDecl, ok bool) {
	if decl, ok = g.types[dir][t]; ok {
		return decl, true
	}
	// This may return an exported type used by an imported function, which is disallowed,
	// except for Component Model administrative functions like resource.drop.
	// TODO: figure out a way to enforce that constraint here.
	decl, ok = g.types[^dir&1][t]
	return decl, ok
}

// typeDir returns the declared type direction (imported or exported) for t.
// If t is not a *[wit.TypeDef], then it returns dir, false.
// If t is not declared, it returns dir, false.
// If t is declared, but in the other direction, it returns the other direction, true.
func (g *generator) typeDir(dir wit.Direction, t wit.Type) (tdir wit.Direction, ok bool) {
	td, ok := t.(*wit.TypeDef)
	if !ok {
		return dir, false
	}
	if _, ok = g.types[dir][td]; ok {
		return dir, true
	}
	// This may return an exported type used by an imported function, which is disallowed,
	// except for Component Model administrative functions like resource.drop.
	// TODO: figure out a way to enforce that constraint here.
	dir2 := ^dir & 1
	if _, ok = g.types[dir2][td]; ok {
		return dir2, true
	}
	return dir, false
}

func (g *generator) typeDefRep(file *gen.File, dir wit.Direction, t *wit.TypeDef, goName string) string {
	switch kind := t.Kind.(type) {
	case *wit.Pointer:
		return g.pointerRep(file, dir, kind)
	case wit.Type:
		return g.typeRep(file, dir, kind)
	case *wit.Record:
		return g.recordRep(file, dir, kind, goName)
	case *wit.Tuple:
		return g.tupleRep(file, dir, kind, goName)
	case *wit.Flags:
		return g.flagsRep(file, dir, kind, goName)
	case *wit.Enum:
		return g.enumRep(file, dir, kind, goName)
	case *wit.Variant:
		return g.variantRep(file, dir, t, goName)
	case *wit.Result:
		return g.resultRep(file, dir, kind)
	case *wit.Option:
		return g.optionRep(file, dir, kind)
	case *wit.List:
		return g.listRep(file, dir, kind)
	case *wit.Resource:
		return g.resourceRep(file, dir, kind)
	case *wit.Own:
		return g.ownRep(file, dir, kind)
	case *wit.Borrow:
		return g.borrowRep(file, dir, kind)
	case *wit.Stream:
		return g.streamRep(file, dir, kind)
	case *wit.Future:
		return g.futureRep(file, dir, kind)
	default:
		panic(fmt.Sprintf("BUG: unknown wit.TypeDefKind %T", kind)) // should never reach here
	}
}

func (g *generator) pointerRep(file *gen.File, dir wit.Direction, p *wit.Pointer) string {
	return "*" + g.typeRep(file, dir, p.Type)
}

func (g *generator) typeRep(file *gen.File, dir wit.Direction, t wit.Type) string {
	switch t := t.(type) {
	// Special-case nil for the _ in result<T, _>
	case nil:
		return "struct{}"

	case *wit.TypeDef:
		if decl, ok := g.typeDecl(dir, t); ok {
			return file.RelativeName(decl.file.Package, decl.name)
		}
		// FIXME: this is only valid for built-in WIT types.
		// User-defined types must be named, so the Ident check above must have succeeded.
		// See https://component-model.bytecodealliance.org/design/wit.html#built-in-types
		// and https://component-model.bytecodealliance.org/design/wit.html#user-defined-types.
		// TODO: add wit.Type.BuiltIn() method?
		return g.typeDefRep(file, dir, t, "")
	case wit.Primitive:
		return g.primitiveRep(file, t)
	default:
		panic(fmt.Sprintf("BUG: unknown wit.Type %T", t)) // should never reach here
	}
}

func (g *generator) primitiveRep(file *gen.File, p wit.Primitive) string {
	switch p := p.(type) {
	case wit.Bool:
		return "bool"
	case wit.S8:
		return "int8"
	case wit.U8:
		return "uint8"
	case wit.S16:
		return "int16"
	case wit.U16:
		return "uint16"
	case wit.S32:
		return "int32"
	case wit.U32:
		return "uint32"
	case wit.S64:
		return "int64"
	case wit.U64:
		return "uint64"
	case wit.F32:
		return "float32"
	case wit.F64:
		return "float64"
	case wit.Char:
		return "rune"
	case wit.String:
		return "string"
	case wit.ErrorContext:
		return file.Import(g.opts.cmPackage) + ".ErrorContext"
	default:
		panic(fmt.Sprintf("BUG: unknown wit.Primitive %T", p)) // should never reach here
	}
}

func (g *generator) recordRep(file *gen.File, dir wit.Direction, r *wit.Record, goName string) string {
	exported := len(goName) == 0 || token.IsExported(goName)
	var b strings.Builder
	b.WriteString("struct {\n")
	stringio.Write(&b, "_ ", file.Import(g.opts.cmPackage), ".HostLayout `json:\"-\"`")
	for i, f := range r.Fields {
		if i == 0 || i > 0 && f.Docs.Contents != "" {
			b.WriteRune('\n')
		}
		b.WriteString(formatDocComments(f.Docs.Contents, false))
		stringio.Write(&b, fieldName(f.Name, exported), " ", g.typeRep(file, dir, f.Type), " `json:\"", f.Name, "\"`\n")
	}
	b.WriteRune('}')
	return b.String()
}

// Field names are implicitly scoped to their parent struct,
// so we don't need to track the mapping between WIT names and Go names.
func fieldName(name string, export bool) string {
	if name == "" {
		return ""
	}
	if name[0] >= '0' && name[0] <= '9' {
		name = "f" + name
	}
	return gen.UniqueName(GoName(name, export), gen.IsReserved)
}

func (g *generator) tupleRep(file *gen.File, dir wit.Direction, t *wit.Tuple, goName string) string {
	var b strings.Builder
	if typ := t.Type(); typ != nil {
		stringio.Write(&b, "[", strconv.Itoa(len(t.Types)), "]", g.typeRep(file, dir, typ))
	} else if len(t.Types) == 0 || len(t.Types) > cm.MaxTuple {
		// Force struct representation
		return g.recordRep(file, dir, t.Despecialize().(*wit.Record), goName)
	} else {
		stringio.Write(&b, file.Import(g.opts.cmPackage), ".Tuple")
		if len(t.Types) > 2 {
			b.WriteString(strconv.Itoa(len(t.Types)))
		}
		b.WriteRune('[')
		for i, typ := range t.Types {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(g.typeRep(file, dir, typ))
		}
		b.WriteRune(']')
	}
	return b.String()
}

func (g *generator) flagsRep(file *gen.File, dir wit.Direction, flags *wit.Flags, goName string) string {
	var b strings.Builder

	// FIXME: this isn't ideal
	var typ wit.Type
	size := flags.Size()
	switch size {
	case 1:
		typ = wit.U8{}
	case 2:
		typ = wit.U16{}
	case 4:
		typ = wit.U32{}
	case 8:
		typ = wit.U64{}
	default:
		panic(fmt.Sprintf("FIXME: cannot emit a flags type with %d cases", len(flags.Flags)))
	}

	b.WriteString(g.typeRep(file, dir, typ))
	b.WriteString("\n\n")
	b.WriteString("const (\n")
	for i, flag := range flags.Flags {
		if i > 0 && flag.Docs.Contents != "" {
			b.WriteRune('\n')
		}
		b.WriteString(formatDocComments(flag.Docs.Contents, false))
		flagName := file.DeclareName(goName + GoName(flag.Name, true))
		b.WriteString(flagName)
		if i == 0 {
			stringio.Write(&b, " ", goName, " = 1 << iota")
		}
		b.WriteRune('\n')
	}
	b.WriteString(")\n")
	return b.String()
}

func (g *generator) enumRep(file *gen.File, dir wit.Direction, e *wit.Enum, goName string) string {
	var b strings.Builder
	disc := wit.Discriminant(len(e.Cases))
	b.WriteString(g.typeRep(file, dir, disc))
	b.WriteString("\n\n")
	b.WriteString("const (\n")
	for i, c := range e.Cases {
		if i > 0 && c.Docs.Contents != "" {
			b.WriteRune('\n')
		}
		b.WriteString(formatDocComments(c.Docs.Contents, false))
		b.WriteString(file.DeclareName(goName + GoName(c.Name, true)))
		if i == 0 {
			b.WriteRune(' ')
			b.WriteString(goName)
			b.WriteString(" = iota")
		}
		b.WriteRune('\n')
	}
	b.WriteString(")\n\n")

	stringsName := file.DeclareName("_" + GoName(goName, true) + "Strings")
	stringio.Write(&b, "var ", stringsName, " = [", fmt.Sprintf("%d", len(e.Cases)), "]string {\n")
	for _, c := range e.Cases {
		stringio.Write(&b, `"`, c.Name, `"`, ",\n")
	}
	b.WriteString("}\n\n")

	b.WriteString(formatDocComments("String implements [fmt.Stringer], returning the enum case name of e.", true))
	stringio.Write(&b, "func (e ", goName, ") String() string {\n")
	stringio.Write(&b, "return ", stringsName, "[e]\n")
	b.WriteString("}\n\n")

	b.WriteString(formatDocComments("MarshalText implements [encoding.TextMarshaler].", true))
	stringio.Write(&b, "func (e ", goName, ") MarshalText() ([]byte, error) {\n")
	stringio.Write(&b, "return []byte(e.String()), nil\n")
	b.WriteString("}\n\n")

	unmarshalName := file.DeclareName("_" + GoName(goName, true) + "UnmarshalCase")

	b.WriteString(formatDocComments("UnmarshalText implements [encoding.TextUnmarshaler], unmarshaling into an enum case. Returns an error if the supplied text is not one of the enum cases.", true))
	stringio.Write(&b, "func (e *", goName, ") UnmarshalText(text []byte) error {\n")
	stringio.Write(&b, "return ", unmarshalName, "(e, text)\n")
	b.WriteString("}\n\n")

	stringio.Write(&b, "var ", unmarshalName, " = ", file.Import(g.opts.cmPackage), ".CaseUnmarshaler[", goName, "](", stringsName, "[:])\n")

	return b.String()
}

// variantRep is special because it accepts a [wit.TypeDef] instead of a [wit.Variant].
// The TypeDef is used to determine the scope for generating accessor methods on the type.
func (g *generator) variantRep(file *gen.File, dir wit.Direction, t *wit.TypeDef, goName string) string {
	v := t.Kind.(*wit.Variant)
	// If the variant has no associated types, represent the variant as an enum.
	if e := v.Enum(); e != nil {
		return g.enumRep(file, dir, e, goName)
	}

	disc := wit.Discriminant(len(v.Cases))
	types := v.Types()
	shape := variantShape(types)
	align := variantAlign(types)

	var typeShape string
	if len(types) == 1 {
		typeShape = g.typeRep(file, dir, shape)
	} else {
		typeShape = g.typeShape(file, dir, shape)
	}

	decl, _ := g.typeDecl(dir, t)
	scope := decl.scope

	// Emit type
	var b strings.Builder
	cm := file.Import(g.opts.cmPackage)
	stringio.Write(&b, cm, ".Variant[", g.typeRep(file, dir, disc), ", ", typeShape, ", ", g.typeRep(file, dir, align), "]\n\n")

	// Emit cases
	for i, c := range v.Cases {
		caseNum := strconv.Itoa(i)
		caseName := scope.DeclareName(GoName(c.Name, true))
		constructorName := file.DeclareName(goName + caseName)
		typeRep := g.typeRep(file, dir, c.Type)

		// Emit constructor
		stringio.Write(&b, "// ", constructorName, " returns a [", goName, "] of case \"", c.Name, "\".\n")
		b.WriteString("//\n")
		b.WriteString(formatDocComments(c.Docs.Contents, false))
		stringio.Write(&b, "func ", constructorName, "(")
		dataName := "data"
		if c.Type != nil {
			stringio.Write(&b, dataName, " ", typeRep)
		}
		stringio.Write(&b, ") ", goName, " {")
		if c.Type == nil {
			stringio.Write(&b, "var ", dataName, " ", typeRep, "\n")
		}
		stringio.Write(&b, "return ", g.cmCall(file, "New["+goName+"]", caseNum+", "+dataName), "\n")
		b.WriteString("}\n\n")

		// Emit getter
		if c.Type == nil {
			// Case without an associated type returns bool
			stringio.Write(&b, "// ", caseName, " returns true if [", goName, "] represents the variant case \"", c.Name, "\".\n")
			stringio.Write(&b, "func (self *", goName, ") ", caseName, "() bool {\n")
			stringio.Write(&b, "return self.Tag() == ", caseNum)
			b.WriteString("}\n\n")
		} else {
			// Case with associated type T returns *T
			stringio.Write(&b, "// ", caseName, " returns a non-nil *[", typeRep, "] if [", goName, "] represents the variant case \"", c.Name, "\".\n")
			stringio.Write(&b, "func (self *", goName, ") ", caseName, "() *", typeRep, " {\n")
			stringio.Write(&b, "return ", g.cmCall(file, "Case["+typeRep+"]", "self, "+caseNum))
			b.WriteString("}\n\n")
		}
	}

	stringsName := file.DeclareName("_" + GoName(goName, true) + "Strings")
	stringio.Write(&b, "var ", stringsName, " = [", fmt.Sprintf("%d", len(v.Cases)), "]string {\n")
	for _, c := range v.Cases {
		stringio.Write(&b, `"`, c.Name, `"`, ",\n")
	}
	b.WriteString("}\n\n")

	b.WriteString(formatDocComments("String implements [fmt.Stringer], returning the variant case name of v.", true))
	stringio.Write(&b, "func (v ", goName, ") String() string {\n")
	stringio.Write(&b, "return ", stringsName, "[v.Tag()]\n")
	b.WriteString("}\n\n")

	return b.String()
}

func (g *generator) resultRep(file *gen.File, dir wit.Direction, r *wit.Result) string {
	var typeShape string
	shape := variantShape(r.Types())
	if len(r.Types()) == 1 {
		typeShape = g.typeRep(file, dir, shape)
	} else {
		typeShape = g.typeShape(file, dir, shape)
	}

	// Emit type
	var b strings.Builder
	b.WriteString(file.Import(g.opts.cmPackage))
	if r.OK == nil && r.Err == nil {
		b.WriteString(".BoolResult")
	} else {
		stringio.Write(&b, ".Result[", typeShape, ", ", g.typeRep(file, dir, r.OK), ", ", g.typeRep(file, dir, r.Err), "]")
	}
	return b.String()
}

func (g *generator) optionRep(file *gen.File, dir wit.Direction, o *wit.Option) string {
	var b strings.Builder
	stringio.Write(&b, file.Import(g.opts.cmPackage), ".Option[", g.typeRep(file, dir, o.Type), "]")
	return b.String()
}

func (g *generator) listRep(file *gen.File, dir wit.Direction, l *wit.List) string {
	var b strings.Builder
	stringio.Write(&b, file.Import(g.opts.cmPackage), ".List[", g.typeRep(file, dir, l.Type), "]")
	return b.String()
}

func (g *generator) resourceRep(file *gen.File, dir wit.Direction, r *wit.Resource) string {
	return file.Import(g.opts.cmPackage) + ".Resource"
}

func (g *generator) ownRep(file *gen.File, dir wit.Direction, o *wit.Own) string {
	return g.typeRep(file, dir, o.Type)
}

func (g *generator) borrowRep(file *gen.File, dir wit.Direction, b *wit.Borrow) string {
	switch dir {
	case wit.Imported:
		return g.typeRep(file, dir, b.Type)
	case wit.Exported:
		// Exported borrow<T> are represented by a concrete i32 rep.
		return file.Import(g.opts.cmPackage) + ".Rep"
	default:
		panic("BUG: unknown direction " + dir.String())
	}
}

func (g *generator) streamRep(file *gen.File, dir wit.Direction, s *wit.Stream) string {
	var b strings.Builder
	stringio.Write(&b, file.Import(g.opts.cmPackage), ".Stream[", g.typeRep(file, dir, s.Type), "]")
	return b.String()
}

func (g *generator) futureRep(file *gen.File, dir wit.Direction, f *wit.Future) string {
	var b strings.Builder
	stringio.Write(&b, file.Import(g.opts.cmPackage), ".Future[", g.typeRep(file, dir, f.Type), "]")
	return b.String()
}

func (g *generator) typeShape(file *gen.File, dir wit.Direction, t wit.Type) string {
	if !experimentCreateShapeTypes {
		return g.typeRep(file, dir, t)
	}

	switch t := t.(type) {
	case *wit.TypeDef:
		// t = t.Root()
		return g.typeDefShape(file, dir, t)
	default:
		return g.typeRep(file, dir, t)
	}
}

func (g *generator) typeDefShape(file *gen.File, dir wit.Direction, t *wit.TypeDef) string {
	switch kind := t.Root().Kind.(type) {
	case wit.Type:
		return g.typeShape(file, dir, kind)
	case *wit.Variant:
		if kind.Enum() != nil {
			// Variants that can be represented as an enum do not need a custom shape.
			return g.typeRep(file, dir, t)
		}
	case *wit.Tuple:
		if kind.Type() != nil {
			// Monotypic tuples have a packed memory layout.
			return g.typeRep(file, dir, t)
		}
	case *wit.Enum, *wit.Flags, *wit.List,
		*wit.Resource, *wit.Own, *wit.Borrow, *wit.Stream, *wit.Future:
		// Certain types do not need custom type shapes:
		// own, borrow, error-context, stream, future, enum, flags, and list
		return g.typeRep(file, dir, t)
	}

	use := typeUse{file.Package, dir, t}
	name, ok := g.shapes[use]
	if !ok {
		abiFile := g.abiFile(file.Package)
		name = abiFile.DeclareName(g.typeDefGoName(dir, t) + "Shape")
		g.shapes[use] = name
		var b bytes.Buffer
		stringio.Write(&b, "// ", name, " is used for storage in variant or result types.\n")
		stringio.Write(&b, "type ", name, " struct {\n")
		stringio.Write(&b, "_ ", abiFile.Import(g.opts.cmPackage), ".HostLayout\n")
		stringio.Write(&b, "shape [", abiFile.Import("unsafe"), ".Sizeof(", g.typeRep(abiFile, dir, t), "{})]byte\n")
		b.WriteString("}\n\n")
		abiFile.Write(b.Bytes())
	}
	return name
}

// typeDefGoName returns a mangled Go name for t.
func (g *generator) typeDefGoName(dir wit.Direction, t *wit.TypeDef) string {
	if decl, ok := g.types[dir][t]; ok && decl.name != "" {
		return decl.name
	}
	return GoName(t.WIT(nil, t.TypeName()), true)
}

func (g *generator) lowerType(file *gen.File, dir wit.Direction, t wit.Type, input string) string {
	switch t := t.(type) {
	case nil:
		// TODO: should this exist?
		return ""
	case *wit.TypeDef:
		t = t.Root()
		return g.lowerTypeDef(file, dir, t, input)
	case wit.Primitive:
		return g.lowerPrimitive(file, dir, t, input)
	default:
		panic(fmt.Sprintf("BUG: unknown type %T", t)) // should never reach here
	}
}

func (g *generator) lowerTypeDef(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	flat := t.Flat()
	switch kind := t.Kind.(type) {
	case *wit.Pointer:
		// TODO: convert pointer to unsafe.Pointer or uintptr?
		return input
	case wit.Type:
		return g.lowerType(file, dir, kind, input)
	case *wit.Record:
		return g.lowerRecord(file, dir, t, input)
	case *wit.Tuple:
		return g.lowerTuple(file, dir, t, input)
	case *wit.Flags:
		return g.lowerFlags(file, dir, t, input)
	case *wit.Enum:
		return g.cast(file, dir, t, flat[0], input)
	case *wit.Variant:
		return g.lowerVariant(file, dir, t, input)
	case *wit.Result:
		return g.lowerResult(file, dir, t, input)
	case *wit.Option:
		return g.lowerOption(file, dir, t, input)
	case *wit.List:
		return g.cmCall(file, "LowerList", input)
	case *wit.Resource, *wit.Own, *wit.Borrow, *wit.Stream, *wit.Future:
		return g.cmCall(file, "Reinterpret["+g.typeRep(file, dir, flat[0])+"]", input)
	default:
		panic(fmt.Sprintf("BUG: unknown wit.TypeDef %T", kind)) // should never reach here
	}
}

func (g *generator) typeDefLowerFunction(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string, body string) string {
	use := typeUse{file.Package, dir, t}
	f, ok := g.lowerFunctions[use]
	if !ok {
		abiFile := g.abiFile(file.Package)
		name := abiFile.DeclareName("lower_" + g.typeDefGoName(dir, t))
		f = g.goFunction(abiFile, dir, wit.Imported, wit.LowerFunction(t), name)
		g.lowerFunctions[use] = f
		stringio.Write(abiFile, "func ", name, g.functionSignature(abiFile, f), " {\n", body, "}\n\n")
	}
	return f.name + "(" + input + ")"
}

func (g *generator) lowerRecord(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	abiFile := g.abiFile(file.Package)
	r := t.Kind.(*wit.Record)
	var b strings.Builder
	i := 0
	for _, f := range r.Fields {
		for j := range f.Type.Flat() {
			if j > 0 {
				b.WriteString(", ")
			}
			stringio.Write(&b, "f"+strconv.Itoa(i))
			i++
		}
		stringio.Write(&b, " = ", g.lowerType(abiFile, dir, f.Type, "v."+fieldName(f.Name, true)), "\n")
	}
	b.WriteString("return\n")
	return g.typeDefLowerFunction(file, dir, t, input, b.String())
}

func (g *generator) lowerTuple(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	tup := t.Kind.(*wit.Tuple)
	mono := tup.Type()
	abiFile := g.abiFile(file.Package)
	var b strings.Builder
	var f int
	for i, tt := range tup.Types {
		for j := range tt.Flat() {
			if j > 0 {
				b.WriteString(", ")
			}
			stringio.Write(&b, "f"+strconv.Itoa(f))
			f++
		}
		field := "v.F" + strconv.Itoa(i)
		if mono != nil {
			field = "v[" + strconv.Itoa(i) + "]" // Monotypic tuples are represented as a fixed-length Go array
		}
		stringio.Write(&b, " = ", g.lowerType(abiFile, dir, tt, field), "\n")
	}
	b.WriteString("return\n")
	return g.typeDefLowerFunction(file, dir, t, input, b.String())
}

func (g *generator) lowerFlags(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	flags := t.Kind.(*wit.Flags)
	flat := t.Flat()
	if len(flat) == 1 {
		return g.cast(file, dir, wit.Discriminant(len(flags.Flags)), flat[0], input)
	}
	body := "// TODO: lower flags with > 32 values\n"
	return g.typeDefLowerFunction(file, dir, t, input, body)
}

func (g *generator) lowerVariant(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	// decl, _ := g.typeDecl(dir, t)
	v := t.Kind.(*wit.Variant)
	flat := t.Flat()
	if v.Enum() != nil {
		return g.cast(file, dir, t, flat[0], input)
	}
	abiFile := g.abiFile(file.Package)
	var b strings.Builder
	stringio.Write(&b, "f0 = ", g.cast(abiFile, dir, wit.Discriminant(len(v.Cases)), flat[0], "v.Tag()"), "\n")
	stringio.Write(&b, "switch f0 {\n")
	for i, c := range v.Cases {
		if c.Type == nil {
			continue
		}
		caseNum := strconv.Itoa(i)
		// caseName := decl.scope.GetName(GoName(c.Name, true))
		input := "*" + g.cmCall(abiFile, "Case["+g.typeRep(file, dir, c.Type)+"]", "&v, "+caseNum)
		stringio.Write(&b, "case ", caseNum, ": // ", c.Name, "\n")
		b.WriteString(g.lowerVariantCaseInto(abiFile, dir, c.Type, flat[1:], input))
	}
	b.WriteString("}\n")
	b.WriteString("return\n")
	return g.typeDefLowerFunction(file, dir, t, input, b.String())
}

func (g *generator) lowerResult(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	r := t.Kind.(*wit.Result)
	if r.OK == nil && r.Err == nil {
		return g.cast(file, dir, wit.Bool{}, wit.U32{}, input)
	}
	flat := t.Flat()
	abiFile := g.abiFile(file.Package)
	var b strings.Builder
	stringio.Write(&b, "if v.IsOK() {\n")
	b.WriteString(g.lowerVariantCaseInto(abiFile, dir, r.OK, flat[1:], "*v.OK()"))
	b.WriteString("} else {\n")
	b.WriteString("f0 = 1\n")
	b.WriteString(g.lowerVariantCaseInto(abiFile, dir, r.Err, flat[1:], "*v.Err()"))
	b.WriteString("}\n")
	b.WriteString("return\n")
	return g.typeDefLowerFunction(file, dir, t, input, b.String())
}

func (g *generator) lowerOption(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	o := t.Kind.(*wit.Option)
	flat := t.Flat()
	abiFile := g.abiFile(file.Package)
	var b strings.Builder
	stringio.Write(&b, "some := v.Some()\n")
	b.WriteString("if some != nil {\n")
	b.WriteString("f0 = 1\n")
	b.WriteString(g.lowerVariantCaseInto(abiFile, dir, o.Type, flat[1:], "*some"))
	b.WriteString("}\n")
	b.WriteString("return\n")
	return g.typeDefLowerFunction(file, dir, t, input, b.String())
}

func (g *generator) lowerVariantCaseInto(file *gen.File, dir wit.Direction, t wit.Type, into []wit.Type, input string) string {
	if t == nil {
		return ""
	}
	var b strings.Builder
	for i := range t.Flat() {
		if i > 0 {
			b.WriteString(", ")
		}
		stringio.Write(&b, "v"+strconv.Itoa(i+1))
	}
	stringio.Write(&b, " := ", g.lowerType(file, dir, t, input), "\n")
	for i, from := range t.Flat() {
		stringio.Write(&b, "f"+strconv.Itoa(i+1), " = ", g.cast(file, dir, from, into[i], "v"+strconv.Itoa(i+1)), "\n")
	}
	return b.String()
}

func (g *generator) lowerPrimitive(file *gen.File, dir wit.Direction, p wit.Primitive, input string) string {
	flat := p.Flat()
	switch p := p.(type) {
	case wit.String:
		return g.cmCall(file, "LowerString", input)
	case wit.ErrorContext:
		return g.cmCall(file, "Reinterpret["+g.typeRep(file, dir, flat[0])+"]", input)
	default:
		return g.cast(file, dir, p, flat[0], input)
	}
}

// liftTypeInput returns a string of typecast parameters for lifting into type t.
func (g *generator) liftTypeInput(file *gen.File, dir wit.Direction, t wit.Type, params []param) string {
	var b strings.Builder
	flat := t.Flat()
	for i, p := range params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(g.cast(file, dir, p.typ, flat[i], p.name))
	}
	return b.String()
}

func (g *generator) liftType(file *gen.File, dir wit.Direction, t wit.Type, input string) string {
	switch t := t.(type) {
	case nil:
		// TODO: should this exist?
		return ""
	case *wit.TypeDef:
		t = t.Root()
		return g.liftTypeDef(file, dir, t, input)
	case wit.Primitive:
		return g.liftPrimitive(file, dir, t, input)
	default:
		panic(fmt.Sprintf("BUG: unknown type %T", t)) // should never reach here
	}
}

func (g *generator) liftTypeDef(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	flat := t.Flat()
	switch kind := t.Kind.(type) {
	case wit.Primitive:
		return g.liftPrimitive(file, dir, t, input)
	case *wit.Pointer:
		// TODO: convert pointer to unsafe.Pointer or uintptr?
		return input
	case wit.Type:
		return g.liftType(file, dir, kind, input)
	case *wit.Record:
		return g.liftRecord(file, dir, t, input)
	case *wit.Tuple:
		return g.liftTuple(file, dir, t, input)
	case *wit.Flags:
		return g.liftFlags(file, dir, t, input)
	case *wit.Enum:
		return g.cast(file, dir, flat[0], t, input)
	case *wit.Variant:
		return g.liftVariant(file, dir, t, input)
	case *wit.Result:
		return g.liftResult(file, dir, t, input)
	case *wit.Option:
		return g.liftOption(file, dir, t, input)
	case *wit.List:
		return g.cmCall(file, "LiftList["+g.typeRep(file, dir, t)+"]", input)
	case *wit.Resource, *wit.Own, *wit.Borrow, *wit.Stream, *wit.Future:
		return g.cmCall(file, "Reinterpret["+g.typeRep(file, dir, t)+"]", input)
	default:
		panic(fmt.Sprintf("BUG: unknown wit.TypeDef %T", kind)) // should never reach here
	}
}

func (g *generator) typeDefLiftFunction(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string, body string) string {
	use := typeUse{file.Package, dir, t}
	f, ok := g.liftFunctions[use]
	if !ok {
		abiFile := g.abiFile(file.Package)
		name := abiFile.DeclareName("lift_" + g.typeDefGoName(dir, t))
		f = g.goFunction(abiFile, dir, wit.Imported, wit.LiftFunction(t), name)
		g.liftFunctions[use] = f
		stringio.Write(abiFile, "func ", name, g.functionSignature(abiFile, f), " {\n", body, "}\n\n")
	}
	return f.name + "(" + input + ")"
}

func (g *generator) liftRecord(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	r := t.Kind.(*wit.Record)
	abiFile := g.abiFile(file.Package)
	var b strings.Builder
	i := 0
	for _, f := range r.Fields {
		var b2 strings.Builder
		for j := range f.Type.Flat() {
			if j > 0 {
				b2.WriteString(", ")
			}
			stringio.Write(&b2, "f"+strconv.Itoa(i))
			i++
		}
		stringio.Write(&b, "v."+fieldName(f.Name, true), " = ", g.liftType(abiFile, dir, f.Type, b2.String()), "\n")
	}
	b.WriteString("return\n")
	return g.typeDefLiftFunction(abiFile, dir, t, input, b.String())
}

func (g *generator) liftTuple(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	tup := t.Kind.(*wit.Tuple)
	mono := tup.Type()
	abiFile := g.abiFile(file.Package)
	var b strings.Builder
	k := 0
	for i, tt := range tup.Types {
		var b2 strings.Builder
		for j := range tt.Flat() {
			if j > 0 {
				b2.WriteString(", ")
			}
			stringio.Write(&b2, "f"+strconv.Itoa(k))
			k++
		}
		field := "v.F" + strconv.Itoa(i)
		if mono != nil {
			field = "v[" + strconv.Itoa(i) + "]" // Monotypic tuples are represented as a fixed-length Go array
		}
		stringio.Write(&b, field, " = ", g.liftType(abiFile, dir, tt, b2.String()), "\n")
	}
	b.WriteString("return\n")
	return g.typeDefLiftFunction(abiFile, dir, t, input, b.String())
}

func (g *generator) liftFlags(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	// flags := t.Kind.(*wit.Flags)
	flat := t.Flat()
	if len(flat) == 1 {
		return g.cast(file, dir, flat[0], t, input)
	}
	body := "// TODO: lift flags with > 32 values\n"
	return g.typeDefLiftFunction(file, dir, t, input, body)
}

func (g *generator) liftVariant(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	v := t.Kind.(*wit.Variant)
	flat := t.Flat()
	if v.Enum() != nil {
		return g.cast(file, dir, flat[0], t, input)
	}
	abiFile := g.abiFile(file.Package)
	var b strings.Builder
	stringio.Write(&b, "switch f0 {\n")
	for i, c := range v.Cases {
		tag := strconv.Itoa(i)
		stringio.Write(&b, "case ", tag, ":\n")
		stringio.Write(&b, "return ", g.cmCall(abiFile, "New["+g.typeRep(abiFile, dir, t)+"]", tag+", "+g.liftVariantCase(abiFile, dir, c.Type, flat[1:])), "\n")
	}
	b.WriteString("}\n")
	stringio.Write(&b, "panic(\"lift variant: unknown case: \" + ", abiFile.Import("strconv"), ".Itoa(int(f0)))\n")
	return g.typeDefLiftFunction(file, dir, t, input, b.String())
}

func (g *generator) liftResult(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	r := t.Kind.(*wit.Result)
	flat := t.Flat()
	if r.OK == nil && r.Err == nil {
		return g.cast(file, dir, wit.Bool{}, t, g.cast(file, dir, flat[0], wit.Bool{}, input))
	}
	abiFile := g.abiFile(file.Package)
	var b strings.Builder
	stringio.Write(&b, "switch f0 {\n")
	b.WriteString("case 0:\n")
	stringio.Write(&b, "return ", g.cmCall(abiFile, "OK["+g.typeRep(abiFile, dir, t)+"]", g.liftVariantCase(abiFile, dir, r.OK, flat[1:])), "\n")
	b.WriteString("case 1:\n")
	stringio.Write(&b, "return ", g.cmCall(abiFile, "Err["+g.typeRep(abiFile, dir, t)+"]", g.liftVariantCase(abiFile, dir, r.Err, flat[1:])), "\n")
	b.WriteString("}\n")
	stringio.Write(&b, "panic(\"lift result: unknown case: \" + ", abiFile.Import("strconv"), ".Itoa(int(f0)))\n")
	return g.typeDefLiftFunction(file, dir, t, input, b.String())
}

func (g *generator) liftOption(file *gen.File, dir wit.Direction, t *wit.TypeDef, input string) string {
	o := t.Kind.(*wit.Option)
	flat := t.Flat()
	abiFile := g.abiFile(file.Package)
	var b strings.Builder
	b.WriteString("if f0 == 0 {\n")
	b.WriteString("return")
	b.WriteString("}\n")
	stringio.Write(&b, "return ", g.cast(abiFile, dir, t, t, g.cmCall(abiFile, "Some["+g.typeRep(abiFile, dir, o.Type)+"]", g.liftVariantCase(abiFile, dir, o.Type, flat[1:]))), "\n")
	return g.typeDefLiftFunction(file, dir, t, input, b.String())
}

func (g *generator) liftVariantCase(file *gen.File, dir wit.Direction, t wit.Type, from []wit.Type) string {
	if t == nil {
		return "struct{}{}"
	}
	var b strings.Builder
	for i, f := range t.Flat() {
		if i > 0 {
			b.WriteString(", ")
		}
		stringio.Write(&b, g.cast(file, dir, from[i], f, "f"+strconv.Itoa(i+1)))
	}
	return g.liftType(file, dir, t, b.String())
}

func (g *generator) liftPrimitive(file *gen.File, dir wit.Direction, t wit.Type, input string) string {
	p, ok := t.(wit.Primitive)
	if !ok {
		p = wit.KindOf[wit.Primitive](t)
		if p == nil {
			panic("BUG: cannot lift non-primitive type")
		}
	}
	flat := p.Flat()
	switch p.(type) {
	case wit.String:
		return g.cmCall(file, "LiftString["+g.typeRep(file, dir, t)+"]", input)
	case wit.ErrorContext:
		return g.cmCall(file, "Reinterpret["+g.typeRep(file, dir, t)+"]", input)
	default:
		return g.cast(file, dir, flat[0], t, input)
	}
}

func (g *generator) cast(file *gen.File, dir wit.Direction, from, to wit.Type, input string) string {
	if castable(from, to) {
		return "(" + g.typeRep(file, dir, to) + ")(" + input + ")"
	}
	t := derefPointer(to)
	if t != nil {
		return g.cmCall(file, goKind(from)+"To"+goKind(to)+"["+g.typeRep(file, dir, t)+"]", input)
	}
	return "(" + g.typeRep(file, dir, to) + ")(" + g.cmCall(file, goKind(from)+"To"+goKind(to), input) + ")"
}

func goKind(t wit.Node) string {
	return GoName(t.WITKind(), true)
}

func castable(from, to wit.Type) bool {
	// Downcast to primitive types if possible
	if p := wit.KindOf[wit.Primitive](from); p != nil {
		from = p
	}
	if p := wit.KindOf[wit.Primitive](to); p != nil {
		to = p
	}

	if to == from {
		return true
	}

	var fromBool, fromInt, fromFloat, fromString, fromPointer bool

	switch from.(type) {
	case wit.Bool:
		fromBool = true
	case wit.S8, wit.U8, wit.S16, wit.U16, wit.S32, wit.U32, wit.S64, wit.U64, wit.Char:
		fromInt = true
	case wit.F32, wit.F64:
		fromFloat = true
	case wit.String:
		fromString = true
	default:
		if r := wit.KindOf[*wit.Result](from); r != nil && len(r.Types()) == 0 {
			fromBool = true
		} else if v := wit.KindOf[*wit.Variant](from); v != nil && v.Enum() != nil {
			fromInt = true
		} else if wit.KindOf[*wit.Enum](from) != nil {
			fromInt = true
		} else if f := wit.KindOf[*wit.Flags](from); f != nil && len(f.Flags) <= 64 {
			fromInt = true
		} else if isPointer(from) {
			fromPointer = true
		} else {
			return false
		}
	}

	switch to.(type) {
	case wit.Bool:
		return fromBool
	case wit.S8, wit.U8, wit.S16, wit.U16, wit.S32, wit.U32, wit.S64, wit.U64, wit.Char:
		return fromInt
	case wit.F32, wit.F64:
		return fromFloat
	case wit.String:
		return fromString
	default:
		if r := wit.KindOf[*wit.Result](to); r != nil && len(r.Types()) == 0 {
			return fromBool
		} else if v := wit.KindOf[*wit.Variant](to); v != nil && v.Enum() != nil {
			return fromInt
		} else if wit.KindOf[*wit.Enum](to) != nil {
			return fromInt
		} else if f := wit.KindOf[*wit.Flags](to); f != nil && len(f.Flags) <= 64 {
			return fromInt
		} else if isPointer(to) {
			return fromPointer
		}
		return false
	}
}

func (g *generator) cmCall(file *gen.File, f string, input string) string {
	return file.Import(g.opts.cmPackage) + "." + f + "(" + input + ")"
}

func (g *generator) ensureParamImports(file *gen.File, dir wit.Direction, params []wit.Param) {
	for i := range params {
		// Ensure type is used in this file to get import path,
		// otherwise short package name may collide with param name.
		_ = g.typeRep(file, dir, params[i].Type)
	}
}

func (g *generator) goFunction(file *gen.File, tdir, dir wit.Direction, f *wit.Function, goName string) function {
	scope := gen.NewScope(file)
	out := function{
		file:    file,
		scope:   scope,
		name:    goName,
		params:  g.goParams(scope, tdir, f.Params),
		results: g.goParams(scope, tdir, f.Results),
	}
	if len(out.results) == 1 && out.results[0].name == "" {
		out.results[0].name = scope.DeclareName("result")
	}
	if dir == wit.Imported && f.IsMethod() {
		out.receiver = out.params[0]
		// out.params = out.params[1:]
	}
	return out
}

func (g *generator) goParams(scope gen.Scope, dir wit.Direction, params []wit.Param) []param {
	out := make([]param, len(params))
	for i, p := range params {
		tdir, _ := g.typeDir(dir, p.Type)
		out[i].name = scope.DeclareName(GoName(p.Name, false))
		out[i].typ = p.Type
		out[i].dir = tdir
	}
	return out
}

func (g *generator) declareFunction(owner wit.TypeOwner, dir wit.Direction, f *wit.Function) (*funcDecl, error) {
	file := g.fileFor(owner)
	wasmFile := g.wasmFileFor(owner)
	var scope gen.Scope = file
	wasm := f.CoreFunction(dir)
	tdir := dir
	module := g.moduleNames[owner]
	if _, ok := owner.(*wit.World); ok {
		module = "$root"
	}
	var goPrefix, linkerName string

	switch dir {
	case wit.Imported:
		goPrefix = "wasmimport_"
		linkerName = module + " " + f.Name

	case wit.Exported:
		scope = g.exportScopes[owner]
		goPrefix = "wasmexport_"
		if module == "$root" {
			linkerName = f.Name
		} else {
			linkerName = module + "#" + f.Name
		}

	case importedWithExportedTypes:
		dir = wit.Imported  // Imported function...
		tdir = wit.Exported // ...with exported types
		goPrefix = "wasmimport_"
		linkerName = "[export]" + module + " " + f.Name

	default:
		return nil, errors.New("BUG: unknown direction " + dir.String())
	}

	if fdecl, ok := g.functions[dir][f]; ok {
		return fdecl, nil
	}

	if dir == wit.Imported {
		g.ensureParamImports(file, tdir, f.Params)
		g.ensureParamImports(file, tdir, f.Results)
	}

	var funcName, wasmName string
	switch f.Kind.(type) {
	case *wit.Freestanding:
		baseName := GoName(f.BaseName(), true)
		funcName = declareDirectedName(scope, dir, baseName)
		wasmName = wasmFile.DeclareName(goPrefix + baseName)

	case *wit.Constructor:
		t := f.Type().(*wit.TypeDef)
		td, _ := g.typeDecl(tdir, t)
		baseName := "New" + td.name
		if dir == wit.Exported {
			baseName = GoName(f.BaseName(), true)
		}
		funcName = declareDirectedName(scope, dir, baseName)
		wasmName = wasmFile.DeclareName(goPrefix + baseName)

	case *wit.Static:
		t := f.Type().(*wit.TypeDef)
		td, _ := g.typeDecl(tdir, t)
		baseName := td.name + GoName(f.BaseName(), true)
		if dir == wit.Exported {
			baseName = GoName(f.BaseName(), true)
		}
		funcName = declareDirectedName(scope, dir, baseName)
		wasmName = wasmFile.DeclareName(goPrefix + baseName)

	case *wit.Method:
		t := f.Type().(*wit.TypeDef)
		if t.Owner != owner {
			return nil, fmt.Errorf("cannot emit methods in package %s on type %s", owner.WITPackage().Name.String(), t.TypeName())
		}
		td, _ := g.typeDecl(tdir, t)
		switch dir {
		case wit.Imported:
			funcName = td.scope.DeclareName(GoName(f.BaseName(), true))
			if wasm.IsMethod() {
				wasmName = td.scope.DeclareName(goPrefix + funcName)
			} else {
				wasmName = wasmFile.DeclareName(goPrefix + td.name + funcName)
			}
		case wit.Exported:
			funcName = td.scope.DeclareName(GoName(f.BaseName(), true))
			wasmName = wasmFile.DeclareName(goPrefix + GoName(*t.Name, true) + GoName(f.BaseName(), true))
		}
	}

	fdecl := &funcDecl{
		owner:      owner,
		dir:        dir,
		f:          f,
		goFunc:     g.goFunction(file, tdir, dir, f, funcName),
		wasmFunc:   g.goFunction(wasmFile, tdir, dir, wasm, wasmName),
		linkerName: linkerName,
	}
	g.functions[dir][f] = fdecl
	return fdecl, nil
}

// FIXME: this is a fun hack
const importedWithExportedTypes = 2

func (g *generator) defineFunction(owner wit.TypeOwner, dir wit.Direction, f *wit.Function) error {
	decl, err := g.declareFunction(owner, dir, f)
	if err != nil {
		return err
	}

	switch dir {
	case wit.Imported, importedWithExportedTypes:
		return g.defineImportedFunction(decl)
	case wit.Exported:
		err := g.defineExportedFunction(decl)
		if err != nil {
			return err
		}
		// A post-return function is mechanically different than
		// other functions, in that it might not have a user-defined
		// implementation. Since Go has a GC, this work can be deferred.
		// Additional design work is needed here.
		// if pf := f.PostReturn(dir); pf != nil {
		// 	return g.defineFunction(owner, dir, pf)
		// }
	default:
		return errors.New("BUG: unknown direction " + dir.String())
	}

	return nil
}

func (g *generator) defineImportedFunction(decl *funcDecl) error {
	dir := wit.Imported
	if !g.define(dir, decl.f) {
		return nil
	}

	file := decl.goFunc.file

	// Bridging between Go and wasm function
	callParams := slices.Clone(decl.wasmFunc.params)
	for i := range callParams {
		callParams[i].name = decl.goFunc.scope.DeclareName(callParams[i].name)
	}
	callResults := slices.Clone(decl.wasmFunc.results)
	for i := range callResults {
		callResults[i].name = decl.goFunc.scope.DeclareName(callResults[i].name)
	}

	var compoundParams param
	var compoundResults param
	var pointerParam param
	var pointerResult param
	if len(callParams) > 0 {
		p := callParams[0]
		t := derefAnonRecord(p.typ)
		if len(decl.goFunc.params) > 0 && t != nil {
			compoundParams = p
			g.declareTypeDef(file, dir, t, decl.wasmFunc.name+"_params")
			compoundParams.typ = t
		} else if len(decl.goFunc.params) > 0 && derefPointer(p.typ) == decl.goFunc.params[0].typ {
			pointerParam = p
		}

		p = *last(callParams)
		t = derefAnonRecord(p.typ)
		if len(decl.goFunc.results) > 0 && t != nil && t != compoundParams.typ {
			compoundResults = p
			g.declareTypeDef(file, dir, t, decl.wasmFunc.name+"_results")
			compoundResults.typ = t
		} else if len(decl.goFunc.results) > 0 && derefPointer(p.typ) == decl.goFunc.results[0].typ {
			last(callParams).name = decl.goFunc.results[0].name // Ensure results local, not results_
			pointerResult = p
		}
	}

	var b bytes.Buffer

	// Emit docs
	b.WriteString(g.functionDocs(dir, decl.f, decl.goFunc.name))

	// Emit Go function
	b.WriteString("//go:nosplit\n")
	b.WriteString("func ")
	if decl.goFunc.isMethod() {
		stringio.Write(&b, "(", decl.goFunc.receiver.name, " ", g.typeRep(file, decl.goFunc.receiver.dir, decl.goFunc.receiver.typ), ") ", decl.goFunc.name)
	} else {
		b.WriteString(decl.goFunc.name)
	}
	b.WriteString(g.functionSignature(file, decl.goFunc))

	// Emit function body
	b.WriteString(" {\n")

	// Lower into wasmimport variables
	if pointerParam.typ != nil {
		stringio.Write(&b, callParams[0].name, " := &", decl.goFunc.params[0].name, "\n")
	} else if compoundParams.typ != nil {
		stringio.Write(&b, compoundParams.name, " := ", g.typeRep(file, compoundParams.dir, compoundParams.typ), "{ ")
		for i, p := range decl.goFunc.params {
			if i > 0 {
				b.WriteString(", ")
			}
			// compound parameter struct field names are identical to parameter names
			stringio.Write(&b, p.name, ": ", p.name)
		}
		b.WriteString(" }\n")
	} else if len(callParams) > 0 {
		i := 0
		for _, p := range decl.goFunc.params {
			flat := p.typ.Flat()
			for j := range flat {
				if j > 0 {
					b.WriteString(", ")
				}
				stringio.Write(&b, callParams[i].name)
				i++
			}
			stringio.Write(&b, " := ", g.lowerType(file, p.dir, p.typ, p.name), "\n")
		}
	}

	// Declare result variables
	if compoundResults.typ != nil {
		stringio.Write(&b, "var ", compoundResults.name, " ", g.typeRep(file, compoundResults.dir, compoundResults.typ), "\n")
	}

	// Emit call to wasmimport function
	if len(callResults) > 0 {
		for i, r := range callResults {
			if i > 0 {
				b.WriteString(", ")
			}
			stringio.Write(&b, r.name)
		}
		b.WriteString(" := ")
	}
	stringio.Write(&b, decl.wasmFunc.name, "(")
	for i, p := range callParams {
		if i > 0 {
			b.WriteString(", ")
		}
		t := derefPointer(p.typ)
		// TODO: this logic is ugly
		if t != nil && (t == compoundParams.typ || t == compoundResults.typ || p.typ == pointerResult.typ) {
			b.WriteRune('&')
			b.WriteString(p.name)
		} else {
			b.WriteString(g.cast(file, p.dir, p.typ, p.typ, p.name))
		}
	}
	b.WriteString(")\n")
	if compoundResults.typ != nil {
		rec := wit.KindOf[*wit.Record](compoundResults.typ)
		b.WriteString("return ")
		for i, f := range rec.Fields {
			if i > 0 {
				b.WriteString(", ")
			}
			stringio.Write(&b, compoundResults.name, ".", fieldName(f.Name, false))
		}
		b.WriteString("\n")
	} else if len(callResults) > 0 {
		i := 0
		for _, r := range decl.goFunc.results {
			flat := r.typ.Flat()
			stringio.Write(&b, r.name, " = ", g.liftType(file, r.dir, r.typ, g.liftTypeInput(file, r.dir, r.typ, callResults[i:i+len(flat)])), "\n")
			i += len(flat)
		}
		b.WriteString("return\n")
	} else {
		b.WriteString("return\n")
	}
	b.WriteString("}\n\n")

	// Emit wasmimport function in wasm file
	wasmFile := decl.wasmFunc.file

	stringio.Write(wasmFile, "//go:wasmimport ", decl.linkerName, "\n")
	wasmFile.WriteString("//go:noescape\n")
	wasmFile.WriteString("func ")
	if decl.wasmFunc.isMethod() {
		stringio.Write(wasmFile, "(", decl.wasmFunc.receiver.name, " ", g.typeRep(wasmFile, decl.wasmFunc.receiver.dir, decl.wasmFunc.receiver.typ), ") ", decl.wasmFunc.name)
	} else {
		wasmFile.WriteString(decl.wasmFunc.name)
	}
	wasmFile.WriteString(g.functionSignature(wasmFile, decl.wasmFunc))

	wasmFile.WriteString("\n\n")

	// Emit shared types
	if t, ok := compoundParams.typ.(*wit.TypeDef); ok {
		td, _ := g.typeDecl(dir, t)
		stringio.Write(&b, "// ", td.name, " represents the flattened function params for [", decl.wasmFunc.name, "].\n")
		stringio.Write(&b, "// See the Canonical ABI flattening rules for more information.\n")
		stringio.Write(&b, "type ", td.name, " ", g.typeDefRep(file, dir, t, td.name), "\n\n")
	}

	if t, ok := compoundResults.typ.(*wit.TypeDef); ok {
		td, _ := g.typeDecl(dir, t)
		stringio.Write(&b, "// ", td.name, " represents the flattened function results for [", decl.wasmFunc.name, "].\n")
		stringio.Write(&b, "// See the Canonical ABI flattening rules for more information.\n")
		stringio.Write(&b, "type ", td.name, " ", g.typeDefRep(file, dir, t, td.name), "\n\n")
	}

	// Write to file
	file.Write(b.Bytes())

	return g.ensureEmptyAsm(file.Package)
}

func (g *generator) defineExportedFunction(decl *funcDecl) error {
	dir := wit.Exported
	if !g.define(dir, decl.f) {
		return nil
	}
	file := decl.goFunc.file
	scope := g.exportScopes[decl.owner]

	// Bridging between wasm and Go function
	callParams := slices.Clone(decl.goFunc.params)
	for i := range callParams {
		callParams[i].name = decl.wasmFunc.scope.DeclareName(callParams[i].name)
	}
	callResults := slices.Clone(decl.goFunc.results)
	for i := range callResults {
		callResults[i].name = decl.wasmFunc.scope.DeclareName(callResults[i].name)
	}

	var compoundParams param
	var compoundResults param
	if len(decl.wasmFunc.params) > 0 {
		p := decl.wasmFunc.params[0]
		t := derefAnonRecord(p.typ)
		if len(callParams) > 0 && t != nil {
			compoundParams = p
			g.declareTypeDef(file, dir, t, decl.wasmFunc.name+"_params")
			compoundParams.typ = t
		}
	}

	if len(decl.wasmFunc.results) > 0 {
		r := decl.wasmFunc.results[0]
		t := derefAnonRecord(r.typ)
		if len(callResults) > 0 && t != nil {
			compoundResults = r
			g.declareTypeDef(file, dir, t, decl.wasmFunc.name+"_results")
			compoundResults.typ = t
		}
	}

	// Emit exports declaration in exports file
	{
		exportsFile := g.exportsFileFor(decl.owner)
		stringio.Write(exportsFile, "\n", g.functionDocs(dir, decl.f, decl.goFunc.name))
		stringio.Write(exportsFile, decl.goFunc.name, " func", g.functionSignature(exportsFile, decl.goFunc), "\n")
	}

	// Emit wasmexport function in wasm file
	wasmFile := decl.wasmFunc.file

	stringio.Write(wasmFile, "//go:wasmexport ", decl.linkerName, "\n")
	stringio.Write(wasmFile, "//export ", decl.linkerName, "\n") // TODO: remove this once TinyGo supports go:wasmexport.
	stringio.Write(wasmFile, "func ", decl.wasmFunc.name, g.functionSignature(wasmFile, decl.wasmFunc))

	// Emit function body
	wasmFile.WriteString(" {\n")

	// Lift arguments
	if compoundParams.typ == nil {
		i := 0
		for _, p := range callParams {
			if i < len(decl.wasmFunc.params) && p.typ == derefPointer(decl.wasmFunc.params[i].typ) {
				stringio.Write(wasmFile, p.name, " := *", decl.wasmFunc.params[i].name, "\n")
				i++
				continue
			}
			flat := p.typ.Flat()
			var input string
			if len(flat) > 0 && len(decl.wasmFunc.params) > 0 {
				input = g.liftTypeInput(wasmFile, p.dir, p.typ, decl.wasmFunc.params[i:i+len(flat)])
			}
			stringio.Write(wasmFile, p.name, " := ", g.liftType(wasmFile, p.dir, p.typ, input), "\n")
			i += len(flat)
		}
	}

	// Emit call to caller-defined Go function
	if compoundResults.typ != nil {
		rec := wit.KindOf[*wit.Record](compoundResults.typ)
		stringio.Write(wasmFile, compoundResults.name, " = new(", g.typeRep(wasmFile, compoundResults.dir, compoundResults.typ), ")\n")
		for i, f := range rec.Fields {
			if i > 0 {
				wasmFile.WriteString(", ")
			}
			stringio.Write(wasmFile, compoundResults.name, ".", fieldName(f.Name, false))
		}
		wasmFile.WriteString(" = ")
	} else if len(callResults) > 0 {
		for i, r := range callResults {
			if i > 0 {
				wasmFile.WriteString(", ")
			}
			wasmFile.WriteString(r.name)
		}
		wasmFile.WriteString(" := ")
	}

	// Emit caller-defined function name
	fqName := file.GetName("Exports") + "." + decl.goFunc.name
	if t := decl.f.Type(); t != nil {
		fqName = file.GetName("Exports") + "." + scope.GetName(GoName(t.TypeName(), true)) + "." + decl.goFunc.name
	}
	stringio.Write(wasmFile, fqName, "(")

	// Emit call params
	if compoundParams.typ != nil {
		rec := wit.KindOf[*wit.Record](compoundParams.typ)
		for i, f := range rec.Fields {
			if i > 0 {
				wasmFile.WriteString(", ")
			}
			stringio.Write(wasmFile, compoundParams.name, ".", fieldName(f.Name, false))
		}
	} else {
		for i, p := range callParams {
			if i > 0 {
				wasmFile.WriteString(", ")
			}
			if isPointer(p.typ) {
				wasmFile.WriteString("*")
			}
			wasmFile.WriteString(p.name)
		}
	}
	wasmFile.WriteString(")\n")

	// Lower results
	if len(callResults) > 0 && compoundResults.typ == nil {
		i := 0
		for _, r := range callResults {
			if i < len(decl.wasmFunc.results) {
				wr := decl.wasmFunc.results[i]
				if r.typ == derefPointer(wr.typ) {
					stringio.Write(wasmFile, wr.name, " = &", r.name, "\n")
					i++
					continue
				}
			}
			flat := r.typ.Flat()
			if len(flat) == 0 {
				stringio.Write(wasmFile, "_ = ", r.name, "\n")
			} else {
				for j := range flat {
					if j > 0 {
						wasmFile.WriteString(", ")
					}
					wasmFile.WriteString(decl.wasmFunc.results[i].name)
					i++
				}
				stringio.Write(wasmFile, " = ", g.lowerType(wasmFile, dir, r.typ, r.name), "\n")
			}
		}
	}

	wasmFile.WriteString("return\n")
	wasmFile.WriteString("}\n\n")

	var b bytes.Buffer

	// Emit default function body
	if strings.HasPrefix(decl.f.Name, "[dtor]") || strings.HasPrefix(decl.f.Name, "cabi_post_") {
		stringio.Write(&b, "func init() {")
		stringio.Write(&b, fqName, " = func", g.functionSignature(file, decl.goFunc), " {}\n")
		b.WriteString("}\n\n")
	}

	// Emit shared types
	if t, ok := compoundParams.typ.(*wit.TypeDef); ok {
		td, _ := g.typeDecl(dir, t)
		stringio.Write(&b, "// ", td.name, " represents the flattened function params for [", decl.wasmFunc.name, "].\n")
		stringio.Write(&b, "// See the Canonical ABI flattening rules for more information.\n")
		stringio.Write(&b, "type ", td.name, " ", g.typeDefRep(file, dir, t, td.name), "\n\n")
	}

	if t, ok := compoundResults.typ.(*wit.TypeDef); ok {
		td, _ := g.typeDecl(dir, t)
		stringio.Write(&b, "// ", td.name, " represents the flattened function results for [", decl.wasmFunc.name, "].\n")
		stringio.Write(&b, "// See the Canonical ABI flattening rules for more information.\n")
		stringio.Write(&b, "type ", td.name, " ", g.typeDefRep(file, dir, t, td.name), "\n\n")
	}

	// Write to file
	file.Write(b.Bytes())

	return g.ensureEmptyAsm(file.Package)
}

func (g *generator) functionSignature(file *gen.File, f function) string {
	var b strings.Builder

	b.WriteRune('(')

	// Emit params
	params := f.params
	if f.isMethod() {
		params = params[1:]
	}
	for i, p := range params {
		if i > 0 {
			b.WriteString(", ")
		}
		stringio.Write(&b, p.name, " ", g.typeRep(file, p.dir, p.typ))
	}
	b.WriteString(") ")

	// Emit results
	if len(f.results) == 1 && f.results[0].name == "" {
		b.WriteString(g.typeRep(file, f.results[0].dir, f.results[0].typ))
	} else if len(f.results) > 0 {
		b.WriteRune('(')
		for i, r := range f.results {
			if i > 0 {
				b.WriteString(", ")
			}
			stringio.Write(&b, r.name, " ", g.typeRep(file, r.dir, r.typ))
		}
		b.WriteRune(')')
	}

	return b.String()
}

func last[S ~[]E, E any](s S) *E {
	if len(s) == 0 {
		return nil
	}
	return &s[len(s)-1]
}

func isPointer(t wit.Type) bool {
	if td, ok := t.(*wit.TypeDef); ok {
		if _, ok := td.Kind.(*wit.Pointer); ok {
			return true
		}
	}
	return false
}

func derefPointer(t wit.Type) wit.Type {
	if td, ok := t.(*wit.TypeDef); ok {
		if p, ok := td.Kind.(*wit.Pointer); ok {
			return p.Type
		}
	}
	return nil
}

func derefTypeDef(t wit.Type) *wit.TypeDef {
	if td, ok := derefPointer(t).(*wit.TypeDef); ok {
		return td
	}
	return nil
}

func anonRecord(params []param) *wit.TypeDef {
	r := &wit.Record{}
	for _, p := range params {
		r.Fields = append(r.Fields,
			wit.Field{
				Name: p.name,
				Type: p.typ,
			})
	}
	return &wit.TypeDef{Kind: r}
}

func derefAnonRecord(t wit.Type) *wit.TypeDef {
	if td := derefTypeDef(t); td != nil && td.Name == nil && td.Owner == nil {
		if _, ok := td.Kind.(*wit.Record); ok {
			return td
		}
	}
	return nil
}

func (g *generator) functionDocs(dir wit.Direction, f *wit.Function, goName string) string {
	var b strings.Builder
	kind := f.WITKind()
	dirString := "the " + dir.String()
	if dir == wit.Exported {
		dirString = "the caller-defined, exported"
	}
	if f.IsAdmin() && f.Type() != nil {
		stringio.Write(&b, "// ", goName, " represents ", dirString, " ", f.BaseName(), " for ", f.Type().WITKind(), " \"", f.Type().TypeName(), "\".\n")
	} else if f.IsConstructor() {
		stringio.Write(&b, "// ", goName, " represents ", dirString, " constructor for ", f.Type().WITKind(), " \"", f.Type().TypeName(), "\".\n")
	} else if f.IsFreestanding() {
		stringio.Write(&b, "// ", goName, " represents ", dirString, " ", kind, " \"", f.Name, "\".\n")
	} else {
		stringio.Write(&b, "// ", goName, " represents ", dirString, " ", kind, " \"", f.BaseName(), "\".\n")
	}
	if f.Docs.Contents != "" {
		b.WriteString("//\n")
		b.WriteString(formatDocComments(f.Docs.Contents, false))
	}
	b.WriteString("//\n")
	if !f.IsAdmin() {
		w := strings.TrimSuffix(f.WIT(nil, f.BaseName()), ";")
		b.WriteString(formatDocComments(w, true))
	}
	return b.String()
}

func (g *generator) ensureEmptyAsm(pkg *gen.Package) error {
	f := pkg.File("empty.s")
	if len(f.Content) > 0 {
		return nil
	}
	_, err := f.Write([]byte(emptyAsm))
	return err
}

func (g *generator) abiFile(pkg *gen.Package) *gen.File {
	file := pkg.File("abi.go")
	file.GeneratedBy = g.opts.generatedBy
	return file
}

func (g *generator) fileFor(owner wit.TypeOwner) *gen.File {
	pkg := g.packageFor(owner)
	file := pkg.File(path.Base(pkg.Path) + ".wit.go")
	file.GeneratedBy = g.opts.generatedBy
	return file
}

func (g *generator) witFileFor(owner wit.TypeOwner) *gen.File {
	pkg := g.packageFor(owner)
	file := pkg.File(path.Base(pkg.Path) + ".wit")
	return file
}

func (g *generator) exportsFileFor(owner wit.TypeOwner) *gen.File {
	pkg := g.packageFor(owner)
	file := pkg.File(path.Base(pkg.Path) + ".exports.go")
	file.GeneratedBy = g.opts.generatedBy
	if len(file.Header) == 0 {
		exports := file.GetName("Exports")
		var b strings.Builder
		stringio.Write(&b, "// ", exports, " represents the caller-defined exports from \"", g.moduleNames[owner], "\".\n")
		stringio.Write(&b, "var ", exports, " struct {")
		file.Header = b.String()
	}
	file.Trailer = "}\n"
	return file
}

func (g *generator) wasmFileFor(owner wit.TypeOwner) *gen.File {
	pkg := g.packageFor(owner)
	file := pkg.File(path.Base(pkg.Path) + ".wasm.go")
	file.GeneratedBy = g.opts.generatedBy
	if len(file.Header) == 0 {
		file.Header = fmt.Sprintf("// This file contains wasmimport and wasmexport declarations for \"%s\".\n\n", owner.WITPackage().Name.String())
	}
	return file
}

func (g *generator) cgoFileFor(owner wit.TypeOwner) *gen.File {
	pkg := g.packageFor(owner)
	file := pkg.File(path.Base(pkg.Path) + ".cgo.go")
	file.GeneratedBy = g.opts.generatedBy
	if file.GoBuild == "" {
		file.GoBuild = "tinygo.wasm"
	}
	return file
}

func (g *generator) packageFor(owner wit.TypeOwner) *gen.Package {
	return g.witPackages[owner]
}

func (g *generator) newPackage(w *wit.World, i *wit.Interface, name string) (*gen.Package, error) {
	var owner wit.TypeOwner
	var id wit.Ident

	if i == nil {
		// Derive Go package from the WIT world
		owner = w
		id = w.Package.Name
		id.Extension = w.Name
		name = id.Extension
	} else {
		owner = i
		if i.Name == nil {
			// Derive Go package from the interface declared in the WIT world
			if i.Package != w.Package {
				return nil, fmt.Errorf("BUG: nested interface package %q != world package %q", i.Package.Name.String(), w.Package.Name.String())
			}
			id = w.Package.Name
			id.Extension = w.Name
		} else {
			// Derive Go package from package-scoped interface
			name = *i.Name
			id = i.Package.Name
			id.Extension = name
		}
	}

	// Don’t create the same package twice
	pkg := g.witPackages[owner]
	if pkg != nil {
		return pkg, nil
	}

	// Create the package path and name
	var segments []string
	if g.opts.packageRoot != "" && g.opts.packageRoot != "std" {
		segments = append(segments, g.opts.packageRoot)
	}
	segments = append(segments, id.Namespace, id.Package)
	if g.versioned && id.Version != nil {
		segments = append(segments, "v"+id.Version.String())
	}
	segments = append(segments, id.Extension)
	if name != id.Extension {
		segments = append(segments, name) // for anonymous interfaces nested under worlds
	}
	pkgPath := strings.Join(segments, "/")

	// TODO: write tests for this
	goName := GoPackageName(name)
	// Ensure local name doesn’t conflict with Go keywords or predeclared identifiers
	if gen.UniqueName(goName, gen.IsReserved) != goName {
		// Try with package prefix, like error -> ioerror
		goName = FlatName(id.Package + goName)
		if gen.UniqueName(goName, gen.IsReserved) != goName {
			// Try with namespace prefix, like ioerror -> wasiioerror
			goName = gen.UniqueName(FlatName(id.Namespace+goName), gen.IsReserved)
		}
	}

	pkg = gen.NewPackage(pkgPath + "#" + goName)
	g.packages[pkg.Path] = pkg
	g.witPackages[owner] = pkg
	g.exportScopes[owner] = gen.NewScope(nil)
	pkg.DeclareName("Exports")

	// Write a WebAssembly file that includes a custom section
	// with a name prefixed with "component-type". The contents are the
	// Component Model definition for a world that encapsulates the
	// Component Model types and functions imported into and/or exported
	// from this Go package.
	{
		// Synthesize a unique-ish name
		worldID := w.Package.Name
		worldID.Extension = "WORLD-" + w.Name
		if i != nil {
			worldID.Extension += "-INTERFACE-" + name
		}
		worldName := worldID.String()
		worldName = replacer.Replace(worldName)
		// libFile := pkg.File("lib" + worldName + ".a")

		// Generate wasm file
		res, world := synthesizeWorld(g.res, w, worldName)
		witText := res.WIT(wit.Filter(world, i), "")
		world.Package.Worlds.Delete(worldName) // Undo mutation
		if g.opts.generateWIT {
			witFile := g.witFileFor(owner)
			witFile.WriteString(witText)
		}
		content, err := g.componentEmbed(witText)
		if err != nil {
			g.opts.logger.Errorf("WIT:\n%s\n\n", witText)
			return nil, err
		}
		componentType := &wasm.CustomSection{
			Name:     "component-type:" + worldName,
			Contents: content,
		}
		if false {
			sysoFile := pkg.File(path.Base(pkg.Path) + ".wasm.syso")
			wasm.Write(sysoFile, []wasm.Section{&wasm.LinkingSection{}, componentType})
		}

		// Write Cgo file
		// cgoFile := g.cgoFileFor(owner)
		// stringio.Write(cgoFile, "// #cgo LDFLAGS: -L. -l", lib, "\n")
		// stringio.Write(cgoFile, "import \"C\"\n")
	}

	return pkg, nil
}

var replacer = strings.NewReplacer("/", "-", ":", "-", "@", "-v", ".", "", "%", "")

// componentEmbed runs generated WIT through wasm-tools to generate a wasm file with a component-type custom section.
func (g *generator) componentEmbed(witData string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// TODO: --all-features?
	filename := "component.wit"
	args := []string{"component", "embed", "--only-custom", filename}
	fsMap := map[string]fs.FS{
		"": fstest.MapFS{
			filename: &fstest.MapFile{Data: []byte(witData)},
		},
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := g.wasmTools.Run(ctx, nil, stdout, stderr, fsMap, args...)
	if err != nil {
		if _, ok := err.(*sys.ExitError); ok {
			return nil, fmt.Errorf("wasm-tools: %s", stderr.String())
		}
		return nil, fmt.Errorf("wasm-tools: %w", err)
	}
	return stdout.Bytes(), err
}

func synthesizeWorld(r *wit.Resolve, w *wit.World, name string) (*wit.Resolve, *wit.World) {
	w = w.Clone()
	w.Name = name
	w.Docs = wit.Docs{}
	w.Package.Worlds.Set(name, w)

	r = r.Clone()
	r.Worlds = append(r.Worlds, w)

	return r, w
}
