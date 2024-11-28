//go:build tinygo
// +build tinygo

package wasmtools

import (
	"context"
	"errors"
	"io"
	"io/fs"
)

type Instance struct{}

func New(ctx context.Context) (*Instance, error) {
	return &Instance{}, errors.New("wasm-tools functionality is disabled under TinyGo")
}

func (w *Instance) Close(ctx context.Context) error {
	return nil
}

func (w *Instance) Run(ctx context.Context, args []string, stdin io.Reader, fsMap map[fs.FS]string, name *string) (stdout io.Reader, stderr io.Reader, err error) {
	return nil, nil, errors.New("wasm-tools functionality is disabled under TinyGo")
}
