package generate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bytecodealliance/wasm-tools-go/internal/codec"
	"github.com/bytecodealliance/wasm-tools-go/internal/go/gen"
	"github.com/bytecodealliance/wasm-tools-go/internal/witcli"
	"github.com/bytecodealliance/wasm-tools-go/wit"
	"github.com/bytecodealliance/wasm-tools-go/wit/bindgen"
	"github.com/urfave/cli/v3"
)

// Command is the CLI command for wit.
var Command = &cli.Command{
	Name:    "generate",
	Aliases: []string{"go"},
	Usage:   "generate Go bindings from from WIT (WebAssembly Interface Types)",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "world",
			Aliases:  []string{"w"},
			Value:    "",
			OnlyOnce: true,
			Config:   cli.StringConfig{TrimSpace: true},
			Usage:    "WIT world to generate, otherwise generate all worlds",
		},
		&cli.StringFlag{
			Name:      "out",
			Aliases:   []string{"o"},
			Value:     ".",
			TakesFile: true,
			OnlyOnce:  true,
			Config:    cli.StringConfig{TrimSpace: true},
			Usage:     "output directory",
		},
		&cli.StringFlag{
			Name:     "package-root",
			Aliases:  []string{"p"},
			Value:    "",
			OnlyOnce: true,
			Config:   cli.StringConfig{TrimSpace: true},
			Usage:    "Go package root, e.g. github.com/org/repo/internal",
		},
		&cli.StringFlag{
			Name:     "cm",
			Value:    "",
			OnlyOnce: true,
			Config:   cli.StringConfig{TrimSpace: true},
			Usage:    "Import path for the Component Model utility package, e.g. github.com/bytecodealliance/wasm-tools-go/cm",
		},
		&cli.BoolFlag{
			Name:  "versioned",
			Usage: "emit versioned Go package(s) for each WIT version",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "do not write files; print to stdout",
		},
	},
	Action: action,
}

// Config is the configuration for the `generate` command.
type config struct {
	dryRun    bool
	out       string
	outPerm   os.FileMode
	pkgRoot   string
	world     string
	cm        string
	versioned bool
	forceWIT  bool
	path      string
}

func action(ctx context.Context, cmd *cli.Command) error {
	cfg, err := parseFlags(cmd)
	if err != nil {
		return err
	}

	res, err := witcli.LoadWIT(ctx, cfg.forceWIT, cfg.path)
	if err != nil {
		return err
	}

	packages, err := bindgen.Go(res,
		bindgen.GeneratedBy(cmd.Root().Name),
		bindgen.World(cfg.world),
		bindgen.PackageRoot(cfg.pkgRoot),
		bindgen.Versioned(cfg.versioned),
		bindgen.CMPackage(cfg.cm),
	)
	if err != nil {
		return err
	}

	if err := writeGoPackages(packages, cfg); err != nil {
		return err
	}

	return writeWITPackage(res, cfg)
}

func parseFlags(cmd *cli.Command) (*config, error) {
	dryRun := cmd.Bool("dry-run")
	out := cmd.String("out")

	info, err := os.Stat(out)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", out)
	}
	fmt.Fprintf(os.Stderr, "Output dir: %s\n", out)
	outPerm := info.Mode().Perm()

	pkgRoot := cmd.String("package-root")
	if !cmd.IsSet("package-root") {
		pkgRoot, err = gen.PackagePath(out)
		if err != nil {
			return nil, err
		}
	}
	fmt.Fprintf(os.Stderr, "Package root: %s\n", pkgRoot)

	path, err := witcli.LoadPath(cmd.Args().Slice()...)
	if err != nil {
		return nil, err
	}

	return &config{
		dryRun,
		out,
		outPerm,
		pkgRoot,
		cmd.String("world"),
		cmd.String("cm"),
		cmd.Bool("versioned"),
		cmd.Bool("force-wit"),
		path,
	}, nil
}

func writeGoPackages(packages []*gen.Package, cfg *config) error {
	fmt.Fprintf(os.Stderr, "Generated %d package(s)\n", len(packages))
	for _, pkg := range packages {
		if !pkg.HasContent() {
			fmt.Fprintf(os.Stderr, "Skipping empty package: %s\n", pkg.Path)
			continue
		}
		fmt.Fprintf(os.Stderr, "Generated package: %s\n", pkg.Path)

		for _, filename := range codec.SortedKeys(pkg.Files) {
			file := pkg.Files[filename]
			dir := filepath.Join(cfg.out, strings.TrimPrefix(file.Package.Path, cfg.pkgRoot))
			path := filepath.Join(dir, file.Name)

			if !file.HasContent() {
				fmt.Fprintf(os.Stderr, "Skipping empty file: %s\n", path)
				continue
			}

			if err := os.MkdirAll(dir, cfg.outPerm); err != nil {
				return err
			}

			content, err := file.Bytes()
			if err != nil {
				if content == nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "Error formatting file: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Generated file: %s\n", path)
			}

			if cfg.dryRun {
				fmt.Println(string(content))
				fmt.Println()
				continue
			}

			if err := os.WriteFile(path, content, cfg.outPerm); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeWITPackage(res *wit.Resolve, cfg *config) error {
	witDir := filepath.Join(cfg.out, "wit")
	if err := os.MkdirAll(witDir, cfg.outPerm); err != nil {
		return err
	}
	witFilePath := filepath.Join(witDir, "webassembly.wit.wasm")
	fmt.Fprintf(os.Stderr, "Generated WIT file: %s\n", witFilePath)

	wasmTools, err := exec.LookPath("wasm-tools")
	if err != nil {
		return err
	}

	var stderr bytes.Buffer
	wasmCmd := exec.Command(wasmTools, "component", "wit", "--wasm", "--all-features", "--output", witFilePath)
	wasmCmd.Stderr = &stderr
	wasmCmd.Stdin = bytes.NewReader([]byte(res.WIT(nil, "")))

	if err := wasmCmd.Run(); err != nil {
		fmt.Fprint(os.Stderr, stderr.String())
		return err
	}
	return nil
}
