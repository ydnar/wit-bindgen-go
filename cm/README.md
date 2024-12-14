# go.bytecodealliance.org/cm

[![pkg.go.dev](https://img.shields.io/badge/docs-pkg.go.dev-blue.svg)](https://pkg.go.dev/go.bytecodealliance.org/cm) [![build status](https://img.shields.io/github/actions/workflow/status/bytecodealliance/go-modules/test.yaml?branch=main)](https://github.com/bytecodealliance/go-modules/actions)

## About

Package `cm` contains helper types and functions used by generated packages, such as `option<t>`, `result<ok, err>`, `variant`, `list`, and `resource`. These are intended for use by generated [Component Model](https://github.com/WebAssembly/component-model/blob/main/design/mvp/Explainer.md#type-definitions) bindings, where the caller converts to a Go equivalent. It attempts to map WIT semantics to their equivalent in Go where possible.

### Note on Memory Safety

Package `cm` and generated bindings from `wit-bindgen-go` may have compatibility issues with the Go garbage collector, as they directly represent `variant` and `result` types as tagged unions where a pointer shape may be occupied by a non-pointer value. The GC may detect and throw an error if it detects a non-pointer value in an area it expects to see a pointer. This is an area of active development.

## License

This project is licensed under the Apache 2.0 license with the LLVM exception. See [LICENSE](../LICENSE) for more details.
