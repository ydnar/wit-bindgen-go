package exports

import (
	"unsafe"

	"github.com/ydnar/wasm-tools-go/cm"
)

type NumberHandle cm.Resource

func (n NumberHandle) ResourceRep() cm.Rep {
	return wasmimport_NumberResourceRep(n)
}

//go:wasmimport [export]example:resources/simple [resource-rep]number
//go:noescape
func wasmimport_NumberResourceRep(handle NumberHandle) cm.Rep

//go:wasmimport [export]example:resources/simple [resource-new]number
//go:noescape
func wasmimport_NumberResourceNew(rep cm.Rep) NumberHandle

// implemented by user code
var impl_Number func(rep cm.Rep) NumberMethods

//go:wasmexport example:resources/simple#[constructor]number
func wasmexport_NumberConstructor(value int32) NumberHandle {
	return impl_NumberConstructor(value)
}

// implemented by user code
var impl_NumberConstructor func(value int32) NumberHandle

//go:wasmexport example:resources/simple#[static]number.merge
func wasmexport_NumberMerge(a NumberHandle, b NumberHandle) NumberHandle {
	return impl_NumberMerge(a, b)
}

// implemented by user code
var impl_NumberMerge func(a NumberHandle, b NumberHandle) NumberHandle

//go:wasmexport example:resources/simple#[method]number.value
func wasmexport_NumberValue(rep cm.Rep) int32 {
	self := impl_Number(rep)
	return self.Value()
}

//go:wasmexport example:resources/simple#[method]number.string
func wasmexport_NumberString(rep cm.Rep, result *string) {
	self := impl_Number(rep)
	*result = self.String()
}

// ExportNumber allows the caller to provide a concrete,
// exported implementation of resource "number".
func ExportNumber[T any, Rep NumberRep[T], Exports NumberExports[Rep, T]](exports Exports) {
	impl_Number = func(rep cm.Rep) NumberMethods {
		return unsafeCast[Rep](rep)
	}
	impl_NumberConstructor = func(value int32) NumberHandle {
		return exports.Constructor(value)
	}
	impl_NumberMerge = func(a NumberHandle, b NumberHandle) NumberHandle {
		return exports.Merge(a, b)
	}
}

func unsafeCast[T, V any](v V) T {
	return *(*T)(unsafe.Pointer(&v))
}

type NumberExports[Rep NumberRep[T], T any] interface {
	Constructor(value int32) NumberHandle
	Merge(a NumberHandle, b NumberHandle) NumberHandle
}

type NumberMethods interface {
	Value() int32
	String() string
	ResourceDestructor()
	ResourceRep() cm.Rep
}

type NumberRep[T any] interface {
	cm.RepTypes[T]
	NumberMethods
}

type NumberInterface[Number NumberRep[T], T any] interface {
	Constructor(value int32) NumberHandle
	Merge(a NumberHandle, b NumberHandle) Number
	Choose(a Number, b Number) Number
}

type Interface[Number NumberRep[T0], T0 any] interface {
	Number() NumberInterface[Number, T0]
}
