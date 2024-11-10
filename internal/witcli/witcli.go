package witcli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"go.bytecodealliance.org/internal/oci"
	"go.bytecodealliance.org/wit"
)

// LoadWIT loads a single [wit.Resolve].
// If path is a OCI path, it pulls from the OCI registry and load WIT
// from the buffer.
// If path == "" or "-", then it reads from stdin.
// If the resolved path doesn’t end in ".json", it will attempt to load
// WIT indirectly by processing the input through wasm-tools.
// If forceWIT is true, it will always process input through wasm-tools.
func LoadWIT(ctx context.Context, path string, r io.Reader, forceWIT bool) (*wit.Resolve, error) {
	if oci.IsOCIPath(path) {
		fmt.Fprintf(os.Stderr, "Fetching OCI artifact %s\n", path)
		if b, err := oci.PullWIT(ctx, path); err != nil {
			return nil, err
		} else {
			return wit.DecodeWIT(bytes.NewReader(b))
		}
	}
	forceReader := path == "" || path == "-"
	if forceWIT || (!forceReader && !strings.HasSuffix(path, ".json")) {
		if forceReader {
			return wit.DecodeWIT(r)
		}
		return wit.LoadWIT(path)
	}
	if forceReader {
		return wit.DecodeJSON(r)
	}
	return wit.LoadJSON(path)
}

// LoadPath parses paths and returns the first path.
// If paths is empty, returns "-".
// If paths has more than one element, returns an error.
func LoadPath(paths ...string) (string, error) {
	var path string
	switch len(paths) {
	case 0:
		path = "-"
	case 1:
		path = paths[0]
	default:
		return "", fmt.Errorf("found %d path arguments, expecting 0 or 1", len(paths))
	}
	return path, nil
}
