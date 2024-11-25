package wit

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"go.bytecodealliance.org/internal/relpath"
	"go.bytecodealliance.org/wit/clone"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var (
	update = flag.Bool("update", false, "update golden files")
	debug  = flag.Bool("debug", false, "debug errors by writing files")
)

func compareOrWrite(t *testing.T, path, golden, data string) {
	if *update {
		err := os.WriteFile(golden, []byte(data), 0o644)
		if err != nil {
			t.Error(err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Error(err)
		return
	}
	if string(want) != data {
		dmp := diffmatchpatch.New()
		dmp.PatchMargin = 3
		diffs := dmp.DiffMain(string(want), data, false)
		t.Errorf("value for %s did not match golden value %s:\n%v", path, golden, dmp.DiffPrettyText(diffs))
	}
}

const testdataPath = "../testdata"

func loadTestdata(f func(path string, res *Resolve) error) error {
	return relpath.Walk(testdataPath, func(path string) error {
		res, err := LoadJSON(path)
		if err != nil {
			return err
		}
		return f(path, res)
	}, "*.wit.json", "*.wit.md.json")
}

func TestClone(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			want := res.WIT(nil, "")
			c := clone.Clone(&clone.State{}, res)
			got := c.WIT(nil, "")
			if got != want {
				// t.Errorf("clone %p != original %p", c, res)

				dmp := diffmatchpatch.New()
				dmp.PatchMargin = 3
				diffs := dmp.DiffMain(want, got, false)
				t.Errorf("value for clone did not match expected:\n%v", dmp.DiffPrettyText(diffs))
			}
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

func TestGoldenWITFiles(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			data := res.WIT(nil, "")
			compareOrWrite(t, path, path+".golden.wit", data)
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

var canWasmTools = sync.OnceValue[bool](func() bool {
	// This is explicitly NOT using exec.LookPath so itfails to run on WebAssembly.
	// This disables tests that require wasm-tools.
	err := exec.Command("wasm-tools", "--version").Run()
	return err == nil
})

func TestGoldenWITRoundTrip(t *testing.T) {
	if testing.Short() {
		// t.Skip is not available in TinyGo, requires runtime.Goexit()
		return
	}
	if !canWasmTools() {
		t.Log("skipping test: wasm-tools not installed or cannot fork/exec (TinyGo)")
		return
	}
	err := loadTestdata(func(path string, res *Resolve) error {
		data := res.WIT(nil, "")
		t.Run(path, func(t *testing.T) {
			// Run the generated WIT through wasm-tools to generate JSON.
			cmd := exec.Command("wasm-tools", "component", "wit", "-j", "--all-features")
			cmd.Stdin = strings.NewReader(data)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			err := cmd.Run()
			if err != nil {
				t.Error(err, stderr.String())
				return
			}

			// Parse the JSON into a Resolve.
			res2, err := DecodeJSON(stdout)
			if err != nil {
				t.Error(err)
				return
			}

			// Convert back to WIT.
			data2 := res2.WIT(nil, "")
			if string(data2) != data {
				if *debug {
					os.WriteFile(path+".fail.wit", []byte(data2), os.ModePerm)
				}
				dmp := diffmatchpatch.New()
				dmp.PatchMargin = 3
				diffs := dmp.DiffMain(data, data2, false)
				t.Errorf("round-trip WIT for %s through wasm-tools did not match:\n%v", path, dmp.DiffPrettyText(diffs))
			}
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSizeAndAlign(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			for i, td := range res.TypeDefs {
				name := fmt.Sprintf("TypeDefs[%d]", i)
				if td.Name != nil {
					name += "#" + *td.Name
				}
				t.Run(name, func(t *testing.T) {
					defer func() {
						err := recover()
						if err != nil {
							t.Fatalf("panic: %v", err)
						}
					}()

					got, want := td.Size(), td.Kind.Size()
					if got != want {
						t.Errorf("(*TypeDef).Size(): got %d, expected %d", got, want)
					}

					got, want = td.Align(), td.Kind.Align()
					if got != want {
						t.Errorf("(*TypeDef).Align(): got %d, expected %d", got, want)
					}
				})
			}
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

// TestFunctionReturnsBorrow validates that no functions in the test data return borrowed handles.
func TestFunctionReturnsBorrow(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		// Skip two fixtures:
		// testdata/wit-parser/resources-multiple-returns-borrow.wit.json
		// testdata/wit-parser/resources-return-borrow.wit.json
		if strings.Contains(path, "-return") && strings.Contains(path, "-borrow") {
			return nil
		}
		t.Run(path, func(t *testing.T) {
			// TODO: when GOEXPERIMENT=rangefunc lands:
			// for f := range res.AllFunctions() {
			res.AllFunctions()(func(f *Function) bool {
				t.Run(f.Name, func(t *testing.T) {
					got, want := f.ReturnsBorrow(), false
					if got != want {
						t.Errorf("(*Function).ReturnsBorrow(): got %t, expected %t", got, want)
					}
				})
				return true
			})
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

// TestFunctionBaseName tests the [Function] BaseName method.
func TestFunctionBaseName(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			// TODO: when GOEXPERIMENT=rangefunc lands:
			// for f := range res.AllFunctions() {
			res.AllFunctions()(func(f *Function) bool {
				t.Run(f.Name, func(t *testing.T) {
					want, after, found := strings.Cut(f.Name, ".")
					if found {
						want = after
					}
					if f.IsConstructor() {
						want = "constructor"
					}
					got := f.BaseName()
					if got != want {
						t.Errorf("(*Function).BaseName(): got %s, expected %s", got, want)
					}
					if strings.Contains(got, ".") {
						t.Errorf("(*Function).BaseName(): %s contains \".\"", got)
					}
				})
				return true
			})
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

// TestFunctionNameConsistency tests to see if the names in the map[string]*Function in
// each [Interface] in a [Resolve] is identical to its Name field.
func TestFunctionNameConsistency(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			for i, face := range res.Interfaces {
				if face.Functions.Len() == 0 {
					continue
				}
				name := fmt.Sprintf("Interfaces[%d]", i)
				if face.Name != nil {
					name += "#" + *face.Name
				}
				t.Run(name, func(t *testing.T) {
					face.Functions.All()(func(name string, f *Function) bool {
						t.Run(name, func(t *testing.T) {
							if name != f.Name {
								t.Errorf("Interface.Functions[%q] != %q", name, f.Name)
							}
						})
						return true
					})
				})
			}

			for i, w := range res.Worlds {
				if w.Imports.Len() == 0 && w.Exports.Len() == 0 {
					continue
				}
				name := fmt.Sprintf("Worlds[%d]#%s", i, w.Name)
				t.Run(name, func(t *testing.T) {
					// A world can rename an imported function, so disable this
					// for name, item := range w.Imports {
					// 	f, ok := item.(*Function)
					// 	if !ok {
					// 		continue
					// 	}
					// 	t.Run(fmt.Sprintf("Imports[%q]==%q", name, f.Name), func(t *testing.T) {
					// 		if name != f.Name {
					// 			t.Errorf("Imports[%q] != %q", name, f.Name)
					// 		}
					// 	})
					// }

					// TODO: can a world rename an exported function?
					w.Exports.All()(func(name string, item WorldItem) bool {
						f, ok := item.(*Function)
						if !ok {
							return true
						}
						t.Run(fmt.Sprintf("Exports[%q]==%q", name, f.Name), func(t *testing.T) {
							if name != f.Name {
								t.Errorf("Exports[%q] != %q", name, f.Name)
							}
						})
						return true
					})
				})
			}
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

// TestConstructorResult validates that constructors return own<t>.
func TestConstructorResult(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			res.AllFunctions()(func(f *Function) bool {
				if !f.IsConstructor() {
					return true
				}
				t.Run(f.Name, func(t *testing.T) {
					want := f.Kind.(*Constructor).Type
					switch typ := f.Results[0].Type.(type) {
					default:
						t.Errorf("result[0].Type is not a *TypeDef")

					case *TypeDef:
						switch kind := typ.Kind.(type) {
						default:
							t.Errorf("result[0].Type.Kind is not a *Own")

						case *Own:
							got := kind.Type
							if want != got {
								t.Errorf("constructor result type own<%T> != %T", got, want)
							}
						}
					}
				})
				return true
			})
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

// TestPackageFieldIsNotNil tests to ensure the Package field of all [World] and [Interface]
// values in a [Resolve] are non-nil.
func TestPackageFieldIsNotNil(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			for i, face := range res.Interfaces {
				name := fmt.Sprintf("Interfaces[%d]", i)
				if face.Name != nil {
					name += "#" + *face.Name
				}
				t.Run(name, func(t *testing.T) {
					if face.Package == nil {
						t.Error("Package is nil")
					}
				})
			}
			for i, w := range res.Worlds {
				name := fmt.Sprintf("Worlds[%d]#%s", i, w.Name)
				t.Run(name, func(t *testing.T) {
					if w.Package == nil {
						t.Error("Package is nil")
					}
				})
			}
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

// TestInterfaceNameIsNotNil tests to ensure a non-nil Name field in all [Interface] values in each [Package].
func TestInterfaceNameIsNotNil(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			for _, pkg := range res.Packages {
				t.Run(pkg.Name.String(), func(t *testing.T) {
					pkg.Interfaces.All()(func(name string, face *Interface) bool {
						t.Run(name, func(t *testing.T) {
							if face.Name == nil {
								t.Error("nil name")
							}
						})

						return true
					})
				})
			}
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

// TestTypeDefNamesNotNil verifies that all [Record], [Variant], [Enum], and [Flags]
// types have a non-nil Name.
func TestTypeDefNamesNotNil(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			for i, td := range res.TypeDefs {
				switch td.Kind.(type) {
				case *Record, *Variant, *Enum, *Flags:
				default:
					continue
				}
				name := fmt.Sprintf("TypeDefs[%d]", i)
				if td.Name != nil {
					name += "#" + *td.Name
				}
				t.Run(name, func(t *testing.T) {
					if td.Name == nil {
						t.Error("Name is nil")
					}
				})
			}
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

// TestTypeDefRootOwnersHaveNames verifies that all root [TypeDef] owners have a name, if non-nil.
func TestTypeDefRootOwnersHaveNames(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			for i, td := range res.TypeDefs {
				name := fmt.Sprintf("TypeDefs[%d]", i)
				if td.Name != nil {
					name += "#" + *td.Name
				}
				t.Run(name, func(t *testing.T) {
					td = td.Root()
					switch owner := td.Owner.(type) {
					case *World:
						if owner.Name == "" {
							t.Error("Owner.Name is empty")
						}
					case *Interface:
						if owner.Name != nil && *owner.Name == "" {
							t.Error("*Owner.Name is empty")
						}
					}
				})
			}
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

// TestNoExportedTypeDefs verifies that any [TypeDef] instances in a [World] are
// referenced in Imports, and not Exports.
func TestNoExportedTypeDefs(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			for i, w := range res.Worlds {
				if w.Imports.Len() == 0 && w.Exports.Len() == 0 {
					continue
				}
				name := fmt.Sprintf("Worlds[%d]#%s", i, w.Name)
				t.Run(name, func(t *testing.T) {
					w.Exports.All()(func(name string, item WorldItem) bool {
						if _, ok := item.(*TypeDef); ok {
							t.Errorf("found TypeDef in World.Exports: %s", name)
						}
						return true
					})
				})
			}
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

// TestHasPointer verifies that the hasPointer method and HasPointer function return the same result.
func TestHasPointer(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			for _, td := range res.TypeDefs {
				a := td.hasPointer()
				b := HasPointer(td)
				if a != b {
					t.Errorf("td.hasPointer(): %t != HasPointer(td): %t (%s)", a, b, td.WIT(nil, td.TypeName()))
				}
			}
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

// TestHasBorrowOnNamedTypes verifies that HasBorrow returns true for all named WIT types that contain a borrow<T>.
func TestHasBorrowOnNamedTypes(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			for _, td := range res.TypeDefs {
				if td.Name == nil {
					// Skip anonymous types here. This is tested on `TestHasBorrow`.
					// https://github.com/bytecodealliance/go-modules/issues/167
					continue
				}
				wit := td.Kind.WIT(nil, "")
				a := strings.Contains(wit, "borrow<")
				b := HasBorrow(td)
				if a != b {
					t.Errorf("has \"borrow<\": %t != HasBorrow(td): %t (%s)", a, b, wit)
				}
			}
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

// TestHandlesAreResources verifies that all [Handle] types have an underlying [Resource] type.
func TestHandlesAreResources(t *testing.T) {
	err := loadTestdata(func(path string, res *Resolve) error {
		t.Run(path, func(t *testing.T) {
			for i, td := range res.TypeDefs {
				var handleType *TypeDef
				switch kind := td.Kind.(type) {
				case *Own:
					handleType = kind.Type
				case *Borrow:
					handleType = kind.Type
				default:
					continue
				}
				name := fmt.Sprintf("TypeDefs[%d]", i)
				if td.Name != nil {
					name += "#" + *td.Name
				}
				t.Run(name, func(t *testing.T) {
					switch kind := handleType.Root().Kind.(type) {
					case *Resource:
						// ok
					default:
						t.Errorf("non-resource type in handle: %s", kind.WIT(nil, ""))
					}
				})
			}
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
