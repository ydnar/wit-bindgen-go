package wit

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"go.bytecodealliance.org/internal/witcli"
	"go.bytecodealliance.org/wit"
)

// Command is the CLI command for wit.
var Command = &cli.Command{
	Name:  "wit",
	Usage: "reverses a WIT JSON file into WIT syntax",
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
			Name:     "interface",
			Aliases:  []string{"i"},
			Value:    "",
			OnlyOnce: true,
			Config:   cli.StringConfig{TrimSpace: true},
			Usage:    "WIT interface to generate, otherwise generate all interfaces",
		},
	},
	Action: action,
}

func action(ctx context.Context, cmd *cli.Command) error {
	path, err := witcli.LoadPath(cmd.Args().Slice()...)
	if err != nil {
		return err
	}

	res, err := witcli.LoadWIT(ctx, path, cmd.Reader, cmd.Bool("force-wit"))
	if err != nil {
		return err
	}

	if face := cmd.String("interface"); face != "" {
		i := findInterface(res, face)
		if i == nil {
			return fmt.Errorf("interface %s not found", face)
		}
		res.ConstrainTo(i)
	}

	if world := cmd.String("world"); world != "" {
		w := findWorld(res, world)
		if w == nil {
			return fmt.Errorf("world %s not found", world)
		}
		res.ConstrainTo(w)
	}

	fmt.Print(res.WIT(nil, ""))
	return nil
}

func findWorld(r *wit.Resolve, pattern string) *wit.World {
	for _, w := range r.Worlds {
		if w.Match(pattern) {
			return w
		}
	}
	return nil
}

func findInterface(r *wit.Resolve, pattern string) *wit.Interface {
	for _, i := range r.Interfaces {
		if i.Match(pattern) {
			return i
		}
	}
	return nil
}
