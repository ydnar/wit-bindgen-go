VERSION ?= 0.1.5
REVISION ?= $(shell git rev-parse --short HEAD)$(shell if ! git diff --no-ext-diff --quiet --exit-code; then echo .m; fi)
PKG=github.com/ydnar/wasm-tools-go

GO_LDFLAGS=-ldflags '-X $(PKG)/version.Version=$(VERSION) -X $(PKG)/version.Revision=$(REVISION)'

wit_files = $(sort $(shell find testdata -name '*.wit' ! -name '*.golden.*'))

.PHONY: json
json: $(wit_files)

.PHONY: $(wit_files)
$(wit_files):
	wasm-tools component wit -j --all-features $@ > $@.json

.PHONY: build
build:
	go build $(GO_LDFLAGS) ./cmd/wit-bindgen-go
