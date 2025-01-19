// Package tests exists as a standalone module in the wasm-tools-go repository for
// testing Go code generated from WIT using wit-bindgen-go.
// There are no user-importable packages in this module.

package tests

//go:generate rm -rf ./generated/*
//go:generate mkdir -p ./generated
//go:generate go run go.bytecodealliance.org/cmd/wit-bindgen-go generate --versioned -o ./generated ../testdata/wasi/cli.wit.json
//go:generate go run go.bytecodealliance.org/cmd/wit-bindgen-go generate --versioned -o ./generated ../testdata/issues/issue284.wit
