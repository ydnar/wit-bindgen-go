package wit

import (
	"io"

	"github.com/coreos/go-semver/semver"
	"go.bytecodealliance.org/internal/codec"
	"go.bytecodealliance.org/internal/codec/json"
)

// DecodeJSON decodes JSON from r into a [Resolve] struct.
// It returns any error that may occur during decoding.
func DecodeJSON(r io.Reader) (*Resolve, error) {
	res := &Resolve{}
	dec := json.NewDecoder(r, res)
	err := dec.Decode(res)
	return res, err
}

// ResolveCodec implements the [codec.Resolver] interface
// translating types to decoding/encoding-aware versions.
func (res *Resolve) ResolveCodec(v any) codec.Codec {
	switch v := v.(type) {
	// References
	case **World:
		return &worldCodec{v, res}
	case **Interface:
		return &interfaceCodec{v, res}
	case **TypeDef:
		return &typeDefCodec{v, res}
	case **Package:
		return &packageCodec{v, res}

	// Allocation required
	case **Function:
		return codec.Must(v)

	// Enums
	case *FunctionKind:
		return &functionKindCodec{v}
	case *Handle:
		return &handleCodec{v}
	case *Stability:
		return &stabilityCodec{v}
	case *Type:
		return &typeCodec{v, res}
	case *TypeDefKind:
		return &typeDefKindCodec{v}
	case *TypeOwner:
		return &typeOwnerCodec{v}
	case *WorldItem:
		return &worldItemCodec{v}

	// Other
	case *InterfaceRef:
		return &interfaceRefCodec{v, res}

	// Imported
	case *semver.Version:
		return &semverCodec{&v}
	case **semver.Version:
		return &semverCodec{v}
	}

	return nil
}

func (c *Resolve) getWorld(i int) *World {
	return mustElement(&c.Worlds, i)
}

func (c *Resolve) getInterface(i int) *Interface {
	return mustElement(&c.Interfaces, i)
}

func (c *Resolve) getTypeDef(i int) *TypeDef {
	return mustElement(&c.TypeDefs, i)
}

func (c *Resolve) getPackage(i int) *Package {
	return mustElement(&c.Packages, i)
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (c *Resolve) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "worlds":
		return codec.DecodeSlice(dec, &c.Worlds)
	case "interfaces":
		return codec.DecodeSlice(dec, &c.Interfaces)
	case "types":
		return codec.DecodeSlice(dec, &c.TypeDefs)
	case "packages":
		return codec.DecodeSlice(dec, &c.Packages)
	}
	return nil
}

// worldCodec translates WIT World references or structures into a *World.
type worldCodec struct {
	w **World
	*Resolve
}

func (c *worldCodec) DecodeInt(i int) error {
	*c.w = c.getWorld(i)
	return nil
}

func (c *worldCodec) DecodeField(dec codec.Decoder, name string) error {
	w := codec.Must(c.w)
	switch name {
	case "name":
		return dec.Decode(&w.Name)
	case "imports":
		return dec.Decode(&w.Imports)
	case "exports":
		return dec.Decode(&w.Exports)
	case "package":
		return dec.Decode(&w.Package)
	case "stability":
		return dec.Decode(&w.Stability)
	case "docs":
		return dec.Decode(&w.Docs)
	}
	return nil
}

// interfaceCodec translates WIT Interface references or structures into an *Interface.
type interfaceCodec struct {
	i **Interface
	*Resolve
}

func (c *interfaceCodec) DecodeInt(i int) error {
	*c.i = c.getInterface(i)
	return nil
}

func (c *interfaceCodec) DecodeField(dec codec.Decoder, name string) error {
	i := codec.Must(c.i)
	switch name {
	case "name":
		return dec.Decode(&i.Name)
	case "types":
		return dec.Decode(&i.TypeDefs)
	case "functions":
		return dec.Decode(&i.Functions)
	case "package":
		return dec.Decode(&i.Package)
	case "stability":
		return dec.Decode(&i.Stability)
	case "docs":
		return dec.Decode(&i.Docs)
	}
	return nil
}

// typeDefCodec translates WIT TypeDef references or structures into a *TypeDef.
type typeDefCodec struct {
	t **TypeDef
	*Resolve
}

func (c *typeDefCodec) DecodeInt(i int) error {
	*c.t = c.getTypeDef(i)
	return nil
}

func (c *typeDefCodec) DecodeField(dec codec.Decoder, name string) error {
	t := codec.Must(c.t)
	switch name {
	case "kind":
		return dec.Decode(&t.Kind)
	case "name":
		return dec.Decode(&t.Name)
	case "owner":
		return dec.Decode(&t.Owner)
	case "stability":
		return dec.Decode(&t.Stability)
	case "docs":
		return dec.Decode(&t.Docs)
	}
	return nil
}

// packageCodec translates WIT Package references or structures into a *Package.
type packageCodec struct {
	p **Package
	*Resolve
}

func (c *packageCodec) DecodeInt(i int) error {
	*c.p = c.getPackage(i)
	return nil
}

func (c *packageCodec) DecodeField(dec codec.Decoder, name string) error {
	p := codec.Must(c.p)
	switch name {
	case "name":
		return dec.Decode(&p.Name)
	case "interfaces":
		return dec.Decode(&p.Interfaces)
	case "worlds":
		return dec.Decode(&p.Worlds)
	case "docs":
		return dec.Decode(&p.Docs)
	}
	return nil
}

// DecodeString implements the [codec.StringDecoder] interface
// to decode a string value into an [Ident].
func (pn *Ident) DecodeString(s string) error {
	var err error
	*pn, err = ParseIdent(s)
	return err
}

// interfaceRefCodec translates WIT interface references into an *InterfaceRef.
type interfaceRefCodec struct {
	ref *InterfaceRef
	*Resolve
}

// This exists to support legacy JSON from wasm-tools pre v1.209.0.
// See https://github.com/bytecodealliance/go-modules/issues/151.
func (c *interfaceRefCodec) DecodeInt(i int) error {
	c.ref.Interface = c.getInterface(i)
	return nil
}

func (c *interfaceRefCodec) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "id":
		return dec.Decode(&c.ref.Interface)
	case "stability":
		return dec.Decode(&c.ref.Stability)
	}
	return nil
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (d *Docs) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "contents":
		return dec.Decode(&d.Contents)
	}
	return nil
}

// worldItemCodec translates typed WorldItem references into a WorldItem,
// currently either an Interface or a TypeDef.
type worldItemCodec struct {
	v *WorldItem
}

func (c *worldItemCodec) DecodeField(dec codec.Decoder, name string) error {
	var err error
	switch name {
	case "interface":
		v := &InterfaceRef{}
		err = dec.Decode(v)
		*c.v = v
	case "function":
		var v *Function
		err = dec.Decode(&v)
		*c.v = v
	case "type":
		var v *TypeDef
		err = dec.Decode(&v)
		*c.v = v
	}
	return err
}

// typeCodec translates WIT type strings or reference Idents into a Type.
type typeCodec struct {
	t *Type
	*Resolve
}

// DecodeString translates s into to a primitive WIT type.
func (c *typeCodec) DecodeString(s string) error {
	var err error
	*c.t, err = ParseType(s)
	return err
}

func (c *typeCodec) DecodeInt(i int) error {
	*c.t = c.getTypeDef(i)
	return nil
}

// typeOwnerCodec translates WIT type owner enums into a [TypeOwner].
type typeOwnerCodec struct {
	v *TypeOwner
}

func (c *typeOwnerCodec) DecodeField(dec codec.Decoder, name string) error {
	var err error
	switch name {
	case "interface":
		var v *Interface
		err = dec.Decode(&v)
		*c.v = v
	case "world":
		var v *World
		err = dec.Decode(&v)
		*c.v = v
	}
	return err
}

// typeDefKindCodec translates WIT type owner enums into a [TypeDefKind].
type typeDefKindCodec struct {
	v *TypeDefKind
}

func (c *typeDefKindCodec) DecodeString(s string) error {
	switch s {
	case "resource":
		*c.v = &Resource{}
	case "errorcontext", // https://github.com/bytecodealliance/wasm-tools/pull/1964
		"error-context":
		*c.v = ErrorContext{}
	}
	return nil
}

func (c *typeDefKindCodec) DecodeField(dec codec.Decoder, name string) error {
	var err error
	switch name {
	case "type":
		var v Type
		err = dec.Decode(&v)
		*c.v = v
	case "record":
		v := &Record{}
		err = dec.Decode(v)
		*c.v = v
	case "flags":
		v := &Flags{}
		err = dec.Decode(v)
		*c.v = v
	case "tuple":
		v := &Tuple{}
		err = dec.Decode(v)
		*c.v = v
	case "list":
		v := &List{}
		err = dec.Decode(&v.Type)
		*c.v = v
	case "variant":
		v := &Variant{}
		err = dec.Decode(v)
		*c.v = v
	case "enum":
		v := &Enum{}
		err = dec.Decode(v)
		*c.v = v
	case "option":
		v := &Option{}
		err = dec.Decode(&v.Type)
		*c.v = v
	case "result":
		v := &Result{}
		err = dec.Decode(v)
		*c.v = v
	case "handle":
		var v Handle
		err = dec.Decode(&v)
		*c.v = v
	case "future":
		v := &Future{}
		err = dec.Decode(&v.Type)
		*c.v = v
	case "stream":
		v := &Stream{}
		err = dec.Decode(&v.Type)
		*c.v = v
	}
	return err
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (r *Record) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "fields":
		return codec.DecodeSlice(dec, &r.Fields)
	}
	return nil
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (f *Field) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "name":
		return dec.Decode(&f.Name)
	case "type":
		return dec.Decode(&f.Type)
	case "docs":
		return dec.Decode(&f.Docs)
	}
	return nil
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (f *Flags) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "flags":
		return codec.DecodeSlice(dec, &f.Flags)
	}
	return nil
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (f *Flag) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "name":
		return dec.Decode(&f.Name)
	case "docs":
		return dec.Decode(&f.Docs)
	}
	return nil
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (t *Tuple) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "types":
		return codec.DecodeSlice(dec, &t.Types)
	}
	return nil
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (v *Variant) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "cases":
		return codec.DecodeSlice(dec, &v.Cases)
	}
	return nil
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (c *Case) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "name":
		return dec.Decode(&c.Name)
	case "type":
		return dec.Decode(&c.Type)
	case "docs":
		return dec.Decode(&c.Docs)
	}
	return nil
}

type handleCodec struct {
	v *Handle
}

func (c *handleCodec) DecodeField(dec codec.Decoder, name string) error {
	var err error
	switch name {
	case "own":
		v := &Own{}
		err = dec.Decode(&v.Type)
		*c.v = v
	case "borrow":
		v := &Borrow{}
		err = dec.Decode(&v.Type)
		*c.v = v
	}
	return err
}

type stabilityCodec struct {
	v *Stability
}

func (c *stabilityCodec) DecodeField(dec codec.Decoder, name string) error {
	var err error
	switch name {
	case "stable":
		v := &Stable{}
		err = dec.Decode(v)
		*c.v = v
	case "unstable":
		v := &Unstable{}
		err = dec.Decode(v)
		*c.v = v
	}
	return err
}

func (s *Stable) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "since":
		return dec.Decode(&s.Since)
	case "deprecated":
		return dec.Decode(&s.Deprecated)
	}
	return nil
}

func (u *Unstable) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "feature":
		return dec.Decode(&u.Feature)
	case "deprecated":
		return dec.Decode(&u.Deprecated)
	}
	return nil
}

type semverCodec struct {
	v **semver.Version
}

func (c *semverCodec) DecodeString(s string) error {
	codec.Must(c.v)
	return (*c.v).Set(s)
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (e *Enum) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "cases":
		return codec.DecodeSlice(dec, &e.Cases)
	}
	return nil
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (r *Result) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "ok":
		return dec.Decode(&r.OK)
	case "err":
		return dec.Decode(&r.Err)
	}
	return nil
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (c *EnumCase) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "name":
		return dec.Decode(&c.Name)
	case "docs":
		return dec.Decode(&c.Docs)
	}
	return nil
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (f *Function) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "name":
		return dec.Decode(&f.Name)
	case "kind":
		return dec.Decode(&f.Kind)
	case "params":
		return codec.DecodeSlice(dec, &f.Params)
	case "result":
		// Multiple function results were removed in this PR:
		// https://github.com/bytecodealliance/wasm-tools/pull/2050/files
		// For now, this package preserves the underlying ability to support > 1 result.
		f.Results = []Param{{}}
		return dec.Decode(&f.Results[0].Type)
	case "results":
		return codec.DecodeSlice(dec, &f.Results)
	case "stability":
		return dec.Decode(&f.Stability)
	case "docs":
		return dec.Decode(&f.Docs)
	}
	return nil
}

type functionKindCodec struct {
	v *FunctionKind
}

func (c *functionKindCodec) DecodeString(s string) error {
	switch s {
	case "freestanding":
		*c.v = &Freestanding{}
	}
	return nil
}

func (c *functionKindCodec) DecodeField(dec codec.Decoder, name string) error {
	var err error
	switch name {
	case "method":
		v := &Method{}
		err = dec.Decode(&v.Type)
		*c.v = v
	case "static":
		v := &Static{}
		err = dec.Decode(&v.Type)
		*c.v = v
	case "constructor":
		v := &Constructor{}
		err = dec.Decode(&v.Type)
		*c.v = v
	}
	return err
}

// DecodeField implements the [codec.FieldDecoder] interface
// to decode a struct or JSON object.
func (p *Param) DecodeField(dec codec.Decoder, name string) error {
	switch name {
	case "name":
		return dec.Decode(&p.Name)
	case "type":
		return dec.Decode(&p.Type)
	}
	return nil
}

// mustElement resizes s and allocates a new instance of T if necessary.
func mustElement[S ~[]*E, E any](s *S, i int) *E {
	if i < 0 {
		return nil
	}
	if codec.Resize(s, i) == nil {
		(*s)[i] = new(E)
	}
	return (*s)[i]
}
