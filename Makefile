wit_files = $(sort $(shell find testdata -name '*.wit' ! -name '*.golden.*'))

# json recompiles the JSON intermediate representation test files.
.PHONY: json
json: $(wit_files)

.PHONY: $(wit_files)
$(wit_files):
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

# tests/generated writes generated Go code to the tests directory
.PHONY: tests/generated
tests/generated: json
	go generate ./tests

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
