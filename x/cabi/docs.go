// Package cabi contains a single WebAssembly function exported as cabi_realloc.
//
// To use, import this package with _:
//
//	import _ "go.bytecodealliance.org/x/cabi"
//
// Function realloc is a WebAssembly [core function] that is validated to have the following core function type:
//
//	(func (param $originalPtr i32)
//	      (param $originalSize i32)
//	      (param $alignment i32)
//	      (param $newSize i32)
//	      (result i32))
//
// The [Canonical ABI] will use realloc both to allocate (passing 0 for the first two parameters) and reallocate. If the Canonical ABI needs realloc, validation requires this option to be present (there is no default).
//
// [core function]: https://www.w3.org/TR/wasm-core-2/syntax/modules.html#functions
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
package cabi
