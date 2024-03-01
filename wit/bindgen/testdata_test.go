// TODO: remove this once TinyGo adds runtime.Frame.Entry and reflect.StringHeader.Len is type int
//go:build !tinygo

package bindgen

import (
	"flag"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/ydnar/wasm-tools-go/internal/codec"
	"github.com/ydnar/wasm-tools-go/internal/go/gen"
	"github.com/ydnar/wasm-tools-go/internal/relpath"
	"github.com/ydnar/wasm-tools-go/wit"
	"golang.org/x/tools/go/packages"
)

var writeGoFiles = flag.Bool("write", false, "write generated Go files")

const (
	testdataPath  = "../../testdata"
	generatedPath = "../../generated"
)

func loadTestdata(f func(path string, res *wit.Resolve) error) error {
	return relpath.Walk(testdataPath, func(path string) error {
		res, err := wit.LoadJSON(path)
		if err != nil {
			return err
		}
		return f(path, res)
	}, "*.wit.json")
}

func writeFile(t *testing.T, dir, pkgPath string, file *gen.File) {
	dir = filepath.Join(dir, strings.TrimPrefix(file.Package.Path, pkgPath))
	err := os.MkdirAll(dir, fs.ModePerm)
	if err != nil {
		t.Error(err)
		return
	}

	path := filepath.Join(dir, file.Name)

	b, err := file.Bytes()
	if err != nil {
		if b == nil {
			t.Error(err)
			return
		}
	}

	f, err := os.Create(path)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = f.Write(b)
	f.Close()

	if err != nil {
		t.Errorf("error writing %s: %v", path, err)
	} else {
		t.Logf("wrote %s", path)
	}
}

var canGo = sync.OnceValue[bool](func() bool {
	err := exec.Command("go", "version").Run()
	return err == nil
})

// validateGeneratedGo loads the Go package(s) generated
func validateGeneratedGo(t *testing.T, res *wit.Resolve) {
	if !canGo() {
		t.Log("skipping test: can't run go (TinyGo without fork?)")
		return
	}

	generated, err := relpath.Abs(generatedPath)
	if err != nil {
		t.Error(err)
		return
	}

	testdata, err := relpath.Abs(testdataPath)
	if err != nil {
		t.Error(err)
		return
	}

	pkgPath, err := gen.PackagePath(testdata)
	if err != nil {
		t.Error(err)
		return
	}

	pkgs, err := Go(res,
		GeneratedBy("test"),
		PackageRoot(pkgPath),
	)
	if err != nil {
		t.Error(err)
		return
	}

	pkgMap := make(map[string]*gen.Package)

	cfg := &packages.Config{
		Mode:    packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedTypesSizes,
		Dir:     testdata,
		Fset:    token.NewFileSet(),
		Overlay: make(map[string][]byte),
	}

	for _, pkg := range pkgs {
		if !pkg.HasContent() {
			continue
		}
		pkgMap[pkg.Path] = pkg
		dir := filepath.Join(testdata, strings.TrimPrefix(pkg.Path, pkgPath))
		// cfg.Overlay[dir] = nil
		for _, file := range pkg.Files {
			path := filepath.Join(dir, file.Name)
			src, err := file.Bytes()
			if err != nil {
				t.Error(err)
			}
			cfg.Overlay[path] = src // Keep unformatted file for more testing
		}
	}

	goPackages, err := packages.Load(cfg, codec.Keys(pkgMap)...)
	if err != nil {
		t.Error(err)
		return
	}

	for _, goPkg := range goPackages {
		pkg := pkgMap[goPkg.PkgPath]
		if pkg == nil {
			t.Logf("Skipped package: %s", goPkg.PkgPath)
			continue
		}

		// Check for errors
		for _, err := range goPkg.Errors {
			if err.Kind == 1 && err.Pos == "" {
				continue
			}
			t.Error(err)
		}
		for _, err := range goPkg.TypeErrors {
			t.Error(err)
		}

		// Verify number of files
		count := len(goPkg.OtherFiles)
		// t.Logf("Go package: %s %t", goPkg.PkgPath, goPkg.Types.Complete())
		for _, f := range goPkg.GoFiles {
			count++
			base := filepath.Base(f)
			// t.Logf("Go file: %s", base)
			if pkg.Files[base] == nil {
				t.Errorf("unknown file in package %s: %s", pkg.Path, base)
			}
		}
		if count != len(pkg.Files) {
			t.Errorf("%d files in package %s; expected %d:\n%s", count, pkg.Path, len(pkg.Files),
				strings.Join(append(goPkg.GoFiles, goPkg.OtherFiles...), "\n"))
		}

		// Verify generated names
		if len(goPkg.TypesInfo.Defs) == 0 {
			t.Errorf("package %s has no TypesInfo.Defs", pkg.Path)
		}
		for id, def := range goPkg.TypesInfo.Defs {
			if def == nil || def.Parent() != goPkg.Types.Scope() {
				continue
			}
			// t.Logf("Def: %s", id.String())
			if !pkg.HasName(id.String()) {
				t.Errorf("name %s not found in generated package %s", id.String(), pkg.Path)
			}
		}

		// Write the package to disk if it has errors
		if *writeGoFiles {
			for _, file := range pkg.Files {
				writeFile(t, generated, pkgPath, file)
			}
		}
	}
}

func TestGenerateTestdata(t *testing.T) {
	if testing.Short() {
		// t.Skip is not available in TinyGo, requires runtime.Goexit()
		return
	}
	err := loadTestdata(func(path string, res *wit.Resolve) error {
		t.Run(path, func(t *testing.T) {
			validateGeneratedGo(t, res)
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
