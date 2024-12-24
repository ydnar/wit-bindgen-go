package wasmtools

import (
	"context"
	"io"
	"io/fs"
)

type instance interface {
	Close(ctx context.Context) error
	Run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, fsMap map[string]fs.FS, args ...string) error
}

var _ instance = &Instance{}
