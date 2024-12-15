package wit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"go.bytecodealliance.org/internal/wasmtools"
)

// LoadJSON loads a [WIT] JSON file from path.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func LoadJSON(path string) (*Resolve, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return DecodeJSON(f)
}

// LoadWIT loads [WIT] data from path by processing it through [wasm-tools].
// This will fail if wasm-tools is not in $PATH.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
// [wasm-tools]: https://crates.io/crates/wasm-tools
func LoadWIT(path string) (*Resolve, error) {
	return loadWIT(path, nil)
}

// DecodeWIT decodes [WIT] data from Reader r by processing it through [wasm-tools].
// This will fail if wasm-tools is not in $PATH.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
// [wasm-tools]: https://crates.io/crates/wasm-tools
func DecodeWIT(r io.Reader) (*Resolve, error) {
	return loadWIT("", r)
}

// loadWIT loads WIT data from path or reader by processing it through wasm-tools.
// It accepts either a path or an io.Reader as input, but not both.
// If the path is not "" and "-", it will be used as the input file.
// Otherwise, the reader will be used as the input.
func loadWIT(path string, reader io.Reader) (*Resolve, error) {
	if path != "" && reader != nil {
		return nil, errors.New("cannot set both path and reader; provide only one")
	}

	ctx := context.Background()
	args := []string{"component", "wit", "-j", "--all-features"}
	fsMap := make(map[string]fs.FS)
	var stdin io.Reader

	if path != "" {
		args = append(args, path)
		dir := filepath.Dir(path)
		fsMap[dir] = os.DirFS(dir)
	} else {
		stdin = reader
	}
	wasmTools, err := wasmtools.New(ctx)
	if err != nil {
		return nil, err
	}
	stdout := &bytes.Buffer{}
	err = wasmTools.Run(ctx, stdin, stdout, nil, fsMap, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing wasm-tools: %w", err)
	}
	return DecodeJSON(stdout)
}
