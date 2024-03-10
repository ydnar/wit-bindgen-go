//go:build !wasip1

// Package simple represents the interface "example:resources/simple".
package simple

import (
	"unsafe"

	"github.com/ydnar/wasm-tools-go/cm"
)

// Number represents the resource "example:resources/simple#number".
//
//	resource number
type Number cm.Resource

// ResourceDrop represents the Canonical ABI function "resource-drop".
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

// NewNumber represents constructor "[constructor]number".
//
//	[constructor]number(value: s32)
//
//go:nosplit
func NewNumber(value int32) Number {
	return wasmimport_NewNumber(value)
}

//go:wasmimport example:resources/simple [constructor]number
//go:noescape
func wasmimport_NewNumber(value int32) Number

// NumberMerge represents static function "merge".
//
//	merge: static func(a: borrow<number>, b: borrow<number>) -> own<number>
//
//go:nosplit
func NumberMerge(a Number, b Number) Number {
	return wasmimport_NumberMerge(a, b)
}

//go:wasmimport example:resources/simple [static]number.merge
//go:noescape
func wasmimport_NumberMerge(a Number, b Number) Number

// String represents method "string".
//
//	string: func() -> string
//
//go:nosplit
func (self Number) String() string {
	var result string
	self.wasmimport_String(&result)
	return result
}

//go:wasmimport example:resources/simple [method]number.string
//go:noescape
func (self Number) wasmimport_String(result *string)

// Value represents method "value".
//
//	value: func() -> s32
//
//go:nosplit
func (self Number) Value() int32 {
	return self.wasmimport_Value()
}

//go:wasmimport example:resources/simple [method]number.value
//go:noescape
func (self Number) wasmimport_Value() int32

// -----------------------------------------------
// Below is a design sketch for exported resources
// -----------------------------------------------

//go:wasmimport [export]example:resources/simple [resource-new]number
//go:noescape
func wasmimport_NumberResourceNew(unsafe.Pointer) Number

// implemented by user code
var impl_Number func(rep unsafe.Pointer) NumberMethods

//go:wasmexport example:resources/simple#[constructor]number
func wasmexport_NumberConstructor(value int32) Number {
	ptr := impl_NumberConstructor(value)
	return wasmimport_NumberResourceNew(ptr)
}

// implemented by user code
var impl_NumberConstructor func(value int32) unsafe.Pointer

//go:wasmexport example:resources/simple#[static]number.merge
func wasmexport_NumberMerge(a unsafe.Pointer, b unsafe.Pointer) Number {
	ptr := impl_NumberMerge(a, b)
	return wasmimport_NumberResourceNew(ptr)
}

// implemented by user code
var impl_NumberMerge func(a unsafe.Pointer, b unsafe.Pointer) unsafe.Pointer

//go:wasmexport example:resources/simple#[method]number.value
func wasmexport_NumberValue(rep unsafe.Pointer) int32 {
	self := impl_Number(rep)
	return self.Value()
}

//go:wasmexport example:resources/simple#[method]number.string
func wasmexport_NumberString(rep unsafe.Pointer, result *string) {
	self := impl_Number(rep)
	*result = self.String()
}

// ExportNumber allows the caller to provide a concrete,
// exported implementation of resource "number".
func ExportNumber[
	Exports NumberExports[Rep, T],
	Rep interface {
		*T
		NumberMethods
	},
	T any](exports Exports) {
	impl_Number = func(rep unsafe.Pointer) NumberMethods {
		return Rep(rep)
	}
	impl_NumberConstructor = func(value int32) unsafe.Pointer {
		return unsafe.Pointer(exports.Constructor(value))
	}
	impl_NumberMerge = func(a unsafe.Pointer, b unsafe.Pointer) unsafe.Pointer {
		return unsafe.Pointer(exports.Merge(Rep(a), Rep(b)))
	}
}

type NumberExports[Rep interface {
	*T
	NumberMethods
}, T any] interface {
	Constructor(value int32) Rep
	Merge(a Rep, b Rep) Rep
}

type NumberMethods interface {
	Value() int32
	String() string
}
