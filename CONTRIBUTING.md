# Contributing

This project welcomes contributions for bug fixes, documentation updates, and new features. Development is done through GitHub pull requests. Feel free to reach out to us on the [Bytecode Alliance Zulip](https://bytecodealliance.zulipchat.com/) or the [Gophers Slack](https://gophers.slack.com/).

## Development

Developing and testing this package requires an up-to-date installation of [Go](https://go.dev), [TinyGo](https://tinygo.org), [Wasmtime](https://wasmtime.dev), and [wasm-tools](https://crates.io/crates/wasm-tools).

## Testing

Tests are supported under both Go and TinyGo, on Linux, macOS, and WebAssembly.

```console
go test ./...
tinygo test ./...
```

Testing with WebAssembly (`wasip1`) requires an installation of [`go_wasip1_wasm32_exec`](https://go.dev/blog/wasi) and [Wasmtime](https://wasmtime.dev). WASI 0.2 `wasip2` is supported under TinyGo version 0.33.0 or later.

```console
GOARCH=wasm GOOS=wasip1 go test ./...
GOARCH=wasm GOOS=wasip1 tinygo test ./...
tinygo test -target=wasip2 ./... # requires TinyGo 0.33.0 or later
```

The [testdata](./testdata) directory contains tests that validate the round-trip parsing and re-serialization of WIT files back into WIT. To rebuild the test files, run `make golden`. This will reregenerate the JSON intermediate representation and the equivalent `*.golden.wit` files.

## Continuous Integration

All changes to this project are required to pass the CI suite powered by GitHub Actions. Pull requests will automatically have checks performed and can only be merged once all tests are passing. CI checks currently include:

* Code is formatted correctly (use `go fmt` locally to pass this).
* Tests pass on Go and TinyGo.
* Tests pass on Linux, macOS, and WebAssembly targets.

## License

This project is licensed under the Apache 2.0 license with the LLVM exception. See [LICENSE](LICENSE) for more details.

## Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in this project by you, as defined in the Apache-2.0 license, shall be licensed as above, without any additional terms or conditions.
