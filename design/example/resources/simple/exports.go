package simple

import "unsafe"

//go:wasmimport [export]example:resources/simple [resource-new]number
//go:noescape
func wasmimport_NumberResourceNew(uint) Number

// implemented by user code
var impl_Number func(rep uint) NumberMethods

//go:wasmexport example:resources/simple#[constructor]number
func wasmexport_NumberConstructor(value int32) Number {
	ptr := impl_NumberConstructor(value)
	return wasmimport_NumberResourceNew(ptr)
}

// implemented by user code
var impl_NumberConstructor func(value int32) uint

//go:wasmexport example:resources/simple#[static]number.merge
func wasmexport_NumberMerge(a uint, b uint) Number {
	ptr := impl_NumberMerge(a, b)
	return wasmimport_NumberResourceNew(ptr)
}

// implemented by user code
var impl_NumberMerge func(a uint, b uint) uint

//go:wasmexport example:resources/simple#[method]number.value
func wasmexport_NumberValue(rep uint) int32 {
	self := impl_Number(rep)
	return self.Value()
}

//go:wasmexport example:resources/simple#[method]number.string
func wasmexport_NumberString(rep uint, result *string) {
	self := impl_Number(rep)
	*result = self.String()
}

// ExportNumber allows the caller to provide a concrete,
// exported implementation of resource "number".
func ExportNumber[T any, Rep NumberRep[T], Exports NumberExports[Rep, T]](exports Exports) {
	impl_Number = func(rep uint) NumberMethods {
		return unsafeCast[Rep](rep)
	}
	impl_NumberConstructor = func(value int32) uint {
		return unsafeCast[uint](exports.Constructor(value))
	}
	impl_NumberMerge = func(a uint, b uint) uint {
		return unsafeCast[uint](exports.Merge(unsafeCast[Rep](a), unsafeCast[Rep](b)))
	}
}

func unsafeCast[T, V any](v V) T {
	return *(*T)(unsafe.Pointer(&v))
}

type NumberExports[Rep NumberRep[T], T any] interface {
	Constructor(value int32) Rep
	Merge(a Rep, b Rep) Rep
}

type NumberMethods interface {
	Value() int32
	String() string
}

type NumberRep[T any] interface {
	Rep | *T
	NumberMethods
}

// This should be cm.Rep
type Rep interface {
	~int32 | ~uint32 | ~uintptr
}
