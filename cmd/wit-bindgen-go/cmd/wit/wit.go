package wit

import (
	"context"
	"fmt"
	"slices"

	"github.com/urfave/cli/v3"

	"go.bytecodealliance.org/internal/witcli"
	"go.bytecodealliance.org/wit"
	"go.bytecodealliance.org/wit/clone"
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
		res = filter(res, i)
	}

	if world := cmd.String("world"); world != "" {
		w := findWorld(res, world)
		if w == nil {
			return fmt.Errorf("world %s not found", world)
		}
		res = filter(res, w)
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

func filter(res *wit.Resolve, node wit.Node) *wit.Resolve {
	state := &clone.State{}
	res = clone.Clone(state, res)
	node = *clone.Clone(state, &node)

	packages := slices.Clone(res.Packages)
	res.Packages = nil
	for _, pkg := range packages {
		if !wit.DependsOn(node, pkg) && !wit.DependsOn(pkg, node) {
			continue
		}
		res.Packages = append(res.Packages, pkg)

		pkg.Worlds.All()(func(name string, w *wit.World) bool {
			if !wit.DependsOn(w, node) {
				pkg.Worlds.Delete(name)
				return true
			}

			w.Imports.All()(func(name string, i wit.WorldItem) bool {
				if !wit.DependsOn(node, i) {
					w.Imports.Delete(name)
				}
				return true
			})

			w.Exports.All()(func(name string, i wit.WorldItem) bool {
				if !wit.DependsOn(node, i) {
					w.Exports.Delete(name)
				}
				return true
			})

			return true
		})

		pkg.Interfaces.All()(func(name string, i *wit.Interface) bool {
			if !wit.DependsOn(node, i) {
				pkg.Interfaces.Delete(name)
			}
			return true
		})
	}

	// fmt.Printf("Resolve: %d %d %d %d\n\n",
	// 	len(res.Worlds), len(res.Interfaces), len(res.TypeDefs), len(res.Packages))

	return res
}
