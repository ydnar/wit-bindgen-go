//go:build tinygo

package wasmtools

import (
	"context"
	"errors"
	"io"
	"io/fs"
)

var errTinyGo = errors.New("wasm-tools disabled under TinyGo")

type Instance struct{}

func New(ctx context.Context) (*Instance, error) {
	return &Instance{}, errTinyGo
}

func (w *Instance) Close(ctx context.Context) error {
	return errTinyGo
}

func (w *Instance) Run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, fsMap map[string]fs.FS, args ...string) error {
	return errTinyGo
}
