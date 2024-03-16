package exports

import (
	"github.com/ydnar/wasm-tools-go/cm"
)

// TODO: make this a cm.Handle[T]
type NumberHandle cm.Resource

// NumberResourceNew represents the imported function "[export]example:resources/simple#[resource-new]number".
//
// Create a new Component Model resource handle for [Number].
//
//go:nosplit
func NumberResourceNew(i Number) NumberHandle {
	return wasmimport_NumberResourceNew(i.ResourceRep())
}

//go:wasmimport [export]example:resources/simple [resource-new]number
//go:noescape
func wasmimport_NumberResourceNew(rep cm.Rep) NumberHandle

// ResourceRep represents the the Canonical ABI function "resource-rep".
//
// Return a rep for [Number] from handle.
// Caller is responsible for converting to a concrete implementation of [Number]
//
//go:nosplit
func (self NumberHandle) ResourceRep() cm.Rep {
	return self.wasmimport_ResourceRep()
}

//go:wasmimport [export]example:resources/simple [resource-rep]number
//go:noescape
func (self NumberHandle) wasmimport_ResourceRep() cm.Rep

// ResourceDrop represents represents the Canonical ABI function "resource-drop".
//
// Drops a resource handle.
//
//go:nosplit
func (self NumberHandle) ResourceDrop() {
	self.wasmimport_ResourceDrop()
}

//go:wasmimport example:resources/simple [resource-drop]number
//go:noescape
func (self NumberHandle) wasmimport_ResourceDrop()

//go:wasmexport example:resources/simple#[constructor]number
func wasmexport_NumberConstructor(value int32) NumberHandle {
	return impl.Number().Constructor(value)
}

//go:wasmexport example:resources/simple#[static]number.merge
func wasmexport_NumberMerge(a NumberHandle, b NumberHandle) NumberHandle {
	return impl.Number().Merge(a, b)
}

//go:wasmexport example:resources/simple#[static]number.choose
func wasmexport_NumberChoose(a cm.Rep, b cm.Rep) NumberHandle {
	return impl.Number().Choose(impl_Number(a), impl_Number(b))
}

//go:wasmexport example:resources/simple#[method]number.value
func wasmexport_NumberValue(rep cm.Rep) int32 {
	return impl_Number(rep).Value()
}

//go:wasmexport example:resources/simple#[method]number.string
func wasmexport_NumberString(rep cm.Rep, result *string) {
	*result = impl_Number(rep).String()
}

type Number interface {
	ResourceRep() cm.Rep
	ResourceDestructor()
	Value() int32
	String() string
}

type Interface interface {
	Number() interface {
		Constructor(value int32) NumberHandle
		Merge(a NumberHandle, b NumberHandle) NumberHandle
		Choose(a Number, b Number) NumberHandle
	}
}

func Export[T0 Number](i Interface) {
	impl = i
	impl_Number = func(rep cm.Rep) Number {
		return cm.AsRep[T0](rep)
	}
}

var (
	impl        Interface
	impl_Number func(cm.Rep) Number
)
