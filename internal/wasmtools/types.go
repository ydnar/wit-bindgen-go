package wasmtools

import (
	"context"
	"io"
	"io/fs"
)

// Runner is an interface for running Wasm modules.
type Runner interface {
	Run(ctx context.Context, args []string, stdin io.Reader, fsMap map[fs.FS]string, name *string) (stdout io.Reader, stderr io.Reader, err error)
}
