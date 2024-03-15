package exports

import (
	"unsafe"

	"github.com/ydnar/wasm-tools-go/cm"
)

// NumberResourceNew represents the imported function "example:resources/simple#[resource-new]number".
//
// Create a new Component Model resource handle for [Number].
func NumberResourceNew(i Number) cm.Own[Number] {
	return wasmimport_NumberResourceNew(i.ResourceRep())
}

//go:wasmimport [export]example:resources/simple [resource-new]number
//go:noescape
func wasmimport_NumberResourceNew(rep cm.Rep) cm.Own[Number]

// NumberResourceRep represents the imported function "example:resources/simple#[resource-rep]number".
//
// Return a rep for [Number] from handle.
// Caller is responsible for converting to a concrete implementation of [Number]
func NumberResourceRep(handle cm.Own[Number]) cm.Rep {
	return wasmimport_NumberResourceRep(handle)
}

//go:wasmimport [export]example:resources/simple [resource-rep]number
//go:noescape
func wasmimport_NumberResourceRep(handle cm.Own[Number]) cm.Rep

//go:wasmexport example:resources/simple#[constructor]number
func wasmexport_NumberConstructor(value int32) cm.Own[Number] {
	return impl.Number().Constructor(value)
}

//go:wasmexport example:resources/simple#[static]number.merge
func wasmexport_NumberMerge(a cm.Own[Number], b cm.Own[Number]) cm.Own[Number] {
	return impl.Number().Merge(a, b)
}

//go:wasmexport example:resources/simple#[static]number.choose
func wasmexport_NumberChoose(a cm.Rep, b cm.Rep) cm.Own[Number] {
	return impl.Number().Choose(impl.Number().FromRep(a), impl.Number().FromRep(b))
}

//go:wasmexport example:resources/simple#[method]number.value
func wasmexport_NumberValue(rep cm.Rep) int32 {
	return impl.Number().FromRep(rep).Value()
}

//go:wasmexport example:resources/simple#[method]number.string
func wasmexport_NumberString(rep cm.Rep, result *string) {
	*result = impl.Number().FromRep(rep).String()
}

func unsafeCast[T, V any](v V) T {
	return *(*T)(unsafe.Pointer(&v))
}

type Number interface {
	Value() int32
	String() string
	ResourceDestructor()
	ResourceRep() cm.Rep
}

type Interface interface {
	Number() interface {
		FromRep(cm.Rep) Number
		Constructor(value int32) cm.Own[Number]
		Merge(a cm.Own[Number], b cm.Own[Number]) cm.Own[Number]
		Choose(a Number, b Number) cm.Own[Number]
	}
}

var impl Interface
