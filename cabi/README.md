# Canonical ABI

## Types

Package `cabi` declares a number of types, including [Component Model](https://component-model.bytecodealliance.org/) [primitive types](https://component-model.bytecodealliance.org/design/wit.html#primitive-types), along with resource and handle types.

## cabi_realloc

The `cabi` package contains an exported WebAssembly function `cabi_realloc` ([Canonical ABI] realloc). To use, import this package with `_`:

```
import _ "github.com/ydnar/wasm-tools-go/cabi"
```

`cabi_realloc` is a WebAssembly [core function](https://www.w3.org/TR/wasm-core-2/syntax/modules.html#functions) that is validated to have the following core function type:

```
(func (param $originalPtr i32)
      (param $originalSize i32)
      (param $alignment i32)
      (param $newSize i32)
      (result i32))
```

The [Canonical ABI] will use realloc both to allocate (passing 0 for the first two parameters) and reallocate. If the Canonical ABI needs realloc, validation requires this option to be present (there is no default).

[Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
