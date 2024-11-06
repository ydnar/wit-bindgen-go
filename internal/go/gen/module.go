package gen

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"go.bytecodealliance.org/internal/relpath"
	"golang.org/x/mod/modfile"
)

// PackagePath returns the Go module path and optional package directory path(s)
// for the given directory path dir. Returns an error if dir or its parent directories
// do not contain a go.mod file.
func PackagePath(dir string) (string, error) {
	dir, err := relpath.Abs(dir)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", dir)
	}

	var file string
	var subdirs string
	for {
		// Find a go.mod file in dir
		file = filepath.Join(dir, "go.mod")
		info, err := os.Stat(file)
		if err != nil {
			if dir == "/" {
				return "", errors.New("unable to locate a go.mod file")
			}
			// Pop up to parent dir
			var rest string
			dir, rest = filepath.Split(dir)
			if dir == "" {
				return "", errors.New("unable to locate a go.mod file")
			}
			dir = filepath.Clean(dir)
			subdirs = path.Join(rest, subdirs)
			continue
		}
		if info.IsDir() {
			return "", fmt.Errorf("unexpected directory: %s", file)
		}
		break
	}

	// Read the go.mod file
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("unable to open %s", file)
	}
	mod, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		return "", err
	}

	// Parse it
	modpath := modfile.ModulePath(mod)
	if modpath == "" {
		return "", fmt.Errorf("no module path in %s", file)
	}
	return path.Join(modpath, subdirs), nil
}
