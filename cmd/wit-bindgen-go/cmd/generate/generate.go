package generate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"go.bytecodealliance.org/internal/codec"
	"go.bytecodealliance.org/internal/go/gen"
	"go.bytecodealliance.org/internal/witcli"
	"go.bytecodealliance.org/wit/bindgen"
	"go.bytecodealliance.org/wit/logging"
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
			Usage:    "Import path for the Component Model utility package, e.g. go.bytecodealliance.org/cm",
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
	logger    logging.Logger
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
	cfg, err := parseFlags(ctx, cmd)
	if err != nil {
		return err
	}

	res, err := witcli.LoadWIT(ctx, cfg.path, cmd.Reader, cfg.forceWIT)
	if err != nil {
		return err
	}

	packages, err := bindgen.Go(res,
		bindgen.GeneratedBy(cmd.Root().Name),
		bindgen.Logger(cfg.logger),
		bindgen.World(cfg.world),
		bindgen.PackageRoot(cfg.pkgRoot),
		bindgen.Versioned(cfg.versioned),
		bindgen.CMPackage(cfg.cm),
	)
	if err != nil {
		return err
	}

	return writeGoPackages(ctx, cmd, cfg, packages)
}

func parseFlags(_ context.Context, cmd *cli.Command) (*config, error) {
	logger := witcli.Logger(cmd.Bool("verbose"), cmd.Bool("debug"))
	dryRun := cmd.Bool("dry-run")
	out := cmd.String("out")

	info, err := witcli.FindOrCreateDir(out)
	if err != nil {
		return nil, err
	}
	logger.Infof("Output dir: %s\n", out)
	outPerm := info.Mode().Perm()

	pkgRoot := cmd.String("package-root")
	if !cmd.IsSet("package-root") {
		pkgRoot, err = gen.PackagePath(out)
		if err != nil {
			return nil, err
		}
	}
	logger.Infof("Package root: %s\n", pkgRoot)

	path, err := witcli.LoadPath(cmd.Args().Slice()...)
	if err != nil {
		return nil, err
	}

	return &config{
		logger,
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

func writeGoPackages(_ context.Context, cmd *cli.Command, cfg *config, packages []*gen.Package) error {
	cfg.logger.Infof("Generated %d Go package(s)\n", len(packages))
	for _, pkg := range packages {
		if !pkg.HasContent() {
			cfg.logger.Debugf("Skipped empty package: %s\n", pkg.Path)
			continue
		}
		cfg.logger.Infof("Generated package: %s\n", pkg.Path)

		for _, filename := range codec.SortedKeys(pkg.Files) {
			file := pkg.Files[filename]
			dir := filepath.Join(cfg.out, strings.TrimPrefix(file.Package.Path, cfg.pkgRoot))
			path := filepath.Join(dir, file.Name)

			if !file.HasContent() {
				cfg.logger.Debugf("\tSkipping empty file: %s\n", path)
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
				cfg.logger.Errorf("\tError formatting file: %v\n", err)
			} else {
				cfg.logger.Infof("\t%s\n", path)
			}

			if cfg.dryRun {
				fmt.Fprintln(cmd.Writer, string(content))
				fmt.Fprintln(cmd.Writer)
				continue
			}

			if err := os.WriteFile(path, content, cfg.outPerm); err != nil {
				return err
			}
		}
	}
	return nil
}
