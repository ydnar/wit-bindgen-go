//go:build !tinygo
// +build !tinygo

package wasmtools

import (
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed wasm-tools.wasm
var wasmTools []byte

// Instance is a compiled wazero instance.
type Instance struct {
	runtime wazero.Runtime
	module  wazero.CompiledModule
}

// New creates a new wazero instance.
func New(ctx context.Context) (*Instance, error) {
	c := wazero.NewRuntimeConfig().WithCloseOnContextDone(true)
	r := wazero.NewRuntimeWithConfig(ctx, c)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return nil, fmt.Errorf("error instantiating WASI: %w", err)
	}

	module, err := r.CompileModule(ctx, wasmTools)
	if err != nil {
		return nil, fmt.Errorf("error compiling wasm module: %w", err)
	}
	return &Instance{runtime: r, module: module}, nil
}

// Close closes the wazero runtime resource.
func (w *Instance) Close(ctx context.Context) error {
	return w.runtime.Close(ctx)
}

// Run runs the wasm module with the context, arguments, stdin, filesystem map, and name.
// It returns the stdout, stderr, and error.
// The execution times out after 10 seconds.
func (w *Instance) Run(ctx context.Context, args []string, stdin io.Reader, fsMap map[fs.FS]string, name *string) (stdout io.Reader, stderr io.Reader, err error) {
	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}

	config := wazero.NewModuleConfig().
		WithStdout(stdoutBuffer).
		WithStderr(stderrBuffer).
		WithRandSource(rand.Reader).
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithArgs(append([]string{"wasm-tools.wasm"}, args...)...)
	if name != nil {
		config = config.WithName(*name)
	}
	if stdin != nil {
		config = config.WithStdin(stdin)
	}

	fsConfig := wazero.NewFSConfig()
	for f, guestPath := range fsMap {
		fsConfig = fsConfig.WithFSMount(f, guestPath)
	}
	config = config.WithFSConfig(fsConfig)

	// timeout to 10 seconds
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err = w.runtime.InstantiateModule(ctx, w.module, config)
	if err != nil {
		return nil, nil, fmt.Errorf("error instantiating wasm module: %w", err)
	}

	return stdoutBuffer, stderrBuffer, nil
}
