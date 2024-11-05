package wit

import (
	"context"
	"fmt"

	"go.bytecodealliance.org/internal/witcli"
	"go.bytecodealliance.org/wit"
	"github.com/urfave/cli/v3"
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
	},
	Action: action,
}

func action(ctx context.Context, cmd *cli.Command) error {
	path, err := witcli.LoadPath(cmd.Args().Slice()...)
	if err != nil {
		return err
	}
	res, err := witcli.LoadWIT(ctx, cmd.Bool("force-wit"), path)
	if err != nil {
		return err
	}
	var w *wit.World
	world := cmd.String("world")
	if world != "" {
		w = findWorld(res, world)
		if w == nil {
			return fmt.Errorf("world %s not found", world)
		}
	}
	fmt.Print(res.WIT(w, ""))
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
