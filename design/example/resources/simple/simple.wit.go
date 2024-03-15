// Code generated by wit-bindgen-go. DO NOT EDIT.

//go:build !wasip1

// Package simple represents the interface "example:resources/simple".
package simple

import (
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

// NumberChoose represents static function "choose".
//
//	choose: static func(a: borrow<number>, b: borrow<number>) -> borrow<number>
//
//go:nosplit
func NumberChoose(a Number, b Number) Number {
	return wasmimport_NumberChoose(a, b)
}

//go:wasmimport example:resources/simple [static]number.choose
//go:noescape
func wasmimport_NumberChoose(a Number, b Number) Number

// NumberMerge represents static function "merge".
//
//	merge: static func(a: own<number>, b: own<number>) -> own<number>
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
