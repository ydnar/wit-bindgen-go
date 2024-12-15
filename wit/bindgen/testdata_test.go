// TODO: remove this once TinyGo adds runtime.Frame.Entry and reflect.StringHeader.Len is type int
//go:build !tinygo

package bindgen

import (
	"flag"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"golang.org/x/tools/go/packages"

	"go.bytecodealliance.org/internal/codec"
	"go.bytecodealliance.org/internal/go/gen"
	"go.bytecodealliance.org/internal/relpath"
	"go.bytecodealliance.org/wit"
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
func validateGeneratedGo(t *testing.T, res *wit.Resolve, origin string) {
	if !canGo() {
		t.Log("skipping test: cannot run go command")
		return
	}

	dir := path.Join(generatedPath, origin)
	err := os.MkdirAll(dir, fs.ModePerm)
	if err != nil {
		t.Error(err)
		return
	}

	out, err := relpath.Abs(dir)
	if err != nil {
		t.Error(err)
		return
	}

	pkgPath, err := gen.PackagePath(out)
	if err != nil {
		t.Error(err)
		return
	}

	pkgs, err := Go(res,
		GeneratedBy("test"),
		PackageRoot(pkgPath),
		Versioned(true),
	)
	if err != nil {
		t.Error(err)
		return
	}

	pkgMap := make(map[string]*gen.Package)

	cfg := &packages.Config{
		Mode:    packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedTypesSizes,
		Dir:     out,
		Fset:    token.NewFileSet(),
		Overlay: make(map[string][]byte),
	}
	// cfg.ParseFile = func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
	// 	if _, ok := cfg.Overlay[filename]; !ok && strings.Contains(filename, pkgPath) {
	// 		return nil, nil
	// 	}
	// 	return parser.ParseFile(fset, filename, src, parser.AllErrors|parser.ParseComments)
	// }

	for _, pkg := range pkgs {
		if !pkg.HasContent() {
			continue
		}
		pkgMap[pkg.Path] = pkg
		dir := filepath.Join(out, strings.TrimPrefix(pkg.Path, pkgPath))
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
		count := len(goPkg.OtherFiles) + len(goPkg.IgnoredFiles)
		// t.Logf("Go package: %s %t", goPkg.PkgPath, goPkg.Types.Complete())
		for _, f := range goPkg.GoFiles {
			count++
			base := filepath.Base(f)
			// t.Logf("Go file: %s", base)
			if pkg.Files[base] == nil {
				t.Errorf("unknown file in package %s: %s", pkg.Path, base)
			}
		}
		if count > len(pkg.Files) {
			t.Errorf("%d files in package %s (expected %d):\n%s", count, pkg.Path, len(pkg.Files),
				strings.Join(append(append(goPkg.GoFiles, goPkg.OtherFiles...), goPkg.IgnoredFiles...), "\n"))
		}

		// Verify generated names
		if len(goPkg.TypesInfo.Defs) == 0 {
			t.Errorf("package %s has no TypesInfo.Defs", pkg.Path)
		}
		for id, def := range goPkg.TypesInfo.Defs {
			if def == nil || def.Parent() != goPkg.Types.Scope() {
				continue
			}
			name := id.String()
			// t.Logf("Def: %s %T", name, def.Type())
			if name != "init" && !pkg.HasName(name) {
				t.Errorf("name %s not found in generated package %s", name, pkg.Path)
			}
		}

		// Write the package to disk if it has errors
		if *writeGoFiles {
			for _, file := range pkg.Files {
				writeFile(t, out, pkgPath, file)
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
			origin := strings.TrimSuffix(strings.TrimPrefix(path, testdataPath), ".wit.json")
			validateGeneratedGo(t, res, origin)
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
