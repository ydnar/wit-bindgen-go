wit_files = $(sort $(shell find testdata -name '*.wit' ! -name '*.golden.*'))

# json recompiles the JSON intermediate representation test files.
.PHONY: json
json: $(wit_files)

.PHONY: $(wit_files)
$(wit_files): internal/wasmtools/wasm-tools.wasm
	wasm-tools component wit -j --all-features $@ > $@.json

# golden recompiles the .golden.wit test files.
.PHONY: golden
golden: json
	go test ./wit -update

# generated generated writes test Go code to the filesystem
.PHONY: generated
generated: clean json
	go test ./wit/bindgen -write

.PHONY: clean
clean:
	rm -rf ./generated/*
	rm -f internal/wasmtools/wasm-tools.wasm
	rm -f internal/wasmtools/wasm-tools.wasm.gz

# tests/generated writes generated Go code to the tests directory
.PHONY: tests/generated
tests/generated: json
	go generate ./tests

# wasm-tools builds the internal/wasmtools/wasm-tools.wasm.gz artifact
.PHONY: wasm-tools
wasm-tools: internal/wasmtools/target/wasm32-wasip1/release/wasm-tools.wasm
	gzip -c $< > internal/wasmtools/wasm-tools.wasm.gz

internal/wasmtools/target/wasm32-wasip1/release/wasm-tools.wasm: internal/wasmtools/Cargo.*
	cd internal/wasmtools && \
	cargo build --target wasm32-wasip1 --release -p wasm-tools

# internal/wasmtools/wasm-tools.wasm decompresses wasm-tools.wasm.gz for other make targets
internal/wasmtools/wasm-tools.wasm: internal/wasmtools/wasm-tools.wasm.gz
	gzip -dc internal/wasmtools/wasm-tools.wasm.gz > $@

# test runs Go and TinyGo tests
GOTESTARGS :=
GOTESTMODULES := ./... ./cm/...
.PHONY: test
test:
	go test $(GOTESTARGS) $(GOTESTMODULES)
	GOARCH=wasm GOOS=wasip1 go test $(GOTESTARGS) $(GOTESTMODULES)
	tinygo test $(GOTESTARGS) $(GOTESTMODULES)
	tinygo test -target=wasip1 $(GOTESTARGS) $(GOTESTMODULES)
	tinygo test -target=wasip2 $(GOTESTARGS) $(GOTESTMODULES)
	tinygo test -target=wasip2 $(GOTESTARGS) ./tests/...
