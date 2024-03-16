package exports

import (
	"github.com/ydnar/wasm-tools-go/cm"
)

// TODO: make this a cm.Handle[T]
type Number cm.Resource

// NumberResourceNew represents the imported function "[export]example:resources/simple#[resource-new]number".
//
// Create a new Component Model resource handle for [NumberInterface].
//
//go:nosplit
func NumberResourceNew(i NumberInterface) Number {
	return wasmimport_NumberResourceNew(i.ResourceRep())
}

//go:wasmimport [export]example:resources/simple [resource-new]number
//go:noescape
func wasmimport_NumberResourceNew(rep cm.Rep) Number

// ResourceRep represents the the Canonical ABI function "resource-rep".
//
// Return a [NumberInterface] from a resource handle.
//
//go:nosplit
func (self Number) ResourceRep() NumberInterface {
	return impl.Number(self.wasmimport_ResourceRep())
}

//go:wasmimport [export]example:resources/simple [resource-rep]number
//go:noescape
func (self Number) wasmimport_ResourceRep() cm.Rep

// ResourceDrop represents represents the Canonical ABI function "resource-drop".
//
// Drops a resource handle.
//
//go:nosplit
func (self Number) ResourceDrop() {
	self.wasmimport_ResourceDrop()
}

//go:wasmimport example:resources/simple [resource-drop]number
//go:noescape
func (self Number) wasmimport_ResourceDrop()

//go:wasmexport example:resources/simple#[constructor]number
func wasmexport_NewNumber(value int32) Number {
	return impl.NewNumber(value)
}

//go:wasmexport example:resources/simple#[dtor]number
func wasmexport_NumberDestructor(rep cm.Rep) {
	impl.Number(rep).ResourceDestructor()
}

//go:wasmexport example:resources/simple#[static]number.merge
func wasmexport_NumberMerge(a Number, b Number) Number {
	return impl.NumberMerge(a, b)
}

//go:wasmexport example:resources/simple#[static]number.choose
func wasmexport_NumberChoose(a cm.Rep, b cm.Rep) Number {
	return impl.NumberChoose(impl.Number(a), impl.Number(b))
}

//go:wasmexport example:resources/simple#[method]number.value
func wasmexport_NumberValue(rep cm.Rep) int32 {
	return impl.Number(rep).Value()
}

//go:wasmexport example:resources/simple#[method]number.string
func wasmexport_NumberString(rep cm.Rep, result *string) {
	*result = impl.Number(rep).String()
}

type NumberInterface interface {
	ResourceRep() cm.Rep
	ResourceDestructor()
	Value() int32
	String() string
}

type Interface interface {
	Number(rep cm.Rep) NumberInterface
	NewNumber(value int32) Number
	NumberMerge(a Number, b Number) Number
	NumberChoose(a NumberInterface, b NumberInterface) Number
}

func Export(i Interface) {
	impl = i
}

var impl Interface
