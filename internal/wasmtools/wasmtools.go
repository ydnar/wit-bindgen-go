package wasmtools

import (
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"fmt"
	"io"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed wasm-tools.wasm
var wasmTools []byte

// Executor is an interface for running Wasm modules.
type Executor interface {
	Run(ctx context.Context, args []string, stdin io.Reader, fsMap map[string]string) (stdout *bytes.Buffer, stderr *bytes.Buffer, err error)
}

type WasmTools struct {
	runtime wazero.Runtime
	module  wazero.CompiledModule
}

func NewWasmTools(ctx context.Context) (*WasmTools, error) {
	r := wazero.NewRuntime(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return nil, fmt.Errorf("error instantiating WASI: %w", err)
	}

	// Compile and instantiate the module
	module, err := r.CompileModule(ctx, wasmTools)
	if err != nil {
		return nil, fmt.Errorf("error compiling wasm module: %w", err)
	}
	return &WasmTools{runtime: r, module: module}, nil
}

func (w *WasmTools) Close(ctx context.Context) error {
	return w.runtime.Close(ctx)
}

func (w *WasmTools) Run(ctx context.Context, args []string, stdin io.Reader, fsMap map[string]string, name *string) (stdout *bytes.Buffer, stderr *bytes.Buffer, err error) {
	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}

	config := wazero.NewModuleConfig().
		WithStdout(stdout).
		WithStderr(stderr).
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
	for hostPath, mountPath := range fsMap {
		fsConfig = fsConfig.WithDirMount(hostPath, mountPath)
	}
	config = config.WithFSConfig(fsConfig)

	_, err = w.runtime.InstantiateModule(ctx, w.module, config)
	if err != nil {
		return nil, nil, fmt.Errorf("error instantiating wasm module: %w", err)
	}

	return stdout, stderr, nil
}
