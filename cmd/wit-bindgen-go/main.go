package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"go.bytecodealliance.org/cmd/wit-bindgen-go/cmd/generate"
	"go.bytecodealliance.org/cmd/wit-bindgen-go/cmd/wit"
	"go.bytecodealliance.org/internal/witcli"
)

func main() {
	err := Command.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

var Command = &cli.Command{
	Name:  "wit-bindgen-go",
	Usage: "inspect or manipulate WebAssembly Interface Types for Go",
	Commands: []*cli.Command{
		generate.Command,
		wit.Command,
		version,
	},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "version",
			Usage:       "print the version",
			HideDefault: true,
			Local:       true,
		},
		&cli.BoolFlag{
			Name:  "force-wit",
			Usage: "force loading WIT via wasm-tools",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "print verbose logging messages",
		},
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"vv"},
			Usage:   "print debug logging messages",
		},
	},
	Action: action,
}

func action(ctx context.Context, cmd *cli.Command) error {
	if cmd.Bool("version") {
		return version.Run(ctx, nil)
	}
	return cli.ShowAppHelp(cmd)
}

var version = &cli.Command{
	Name:  "version",
	Usage: "print the version",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		fmt.Printf("%s version %s\n", cmd.Root().Name, witcli.Version())
		return nil
	},
}
