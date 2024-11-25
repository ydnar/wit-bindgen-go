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

	var deps []wit.Node

	if world := cmd.String("world"); world != "" {
		w := findWorld(res, world)
		if w == nil {
			return fmt.Errorf("world %s not found", world)
		}
		deps = append(deps, w)
	}

	if face := cmd.String("interface"); face != "" {
		i := findInterface(res, face)
		if i == nil {
			return fmt.Errorf("interface %s not found", face)
		}
		deps = append(deps, i)
	}

	res = filter(res, deps...)

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

func filter(res *wit.Resolve, deps ...wit.Node) *wit.Resolve {
	if len(deps) == 0 {
		return res
	}

	state := &clone.State{}
	res = clone.Clone(state, res)
	deps = clone.Slice(state, deps)

	packages := slices.Clone(res.Packages)
	res.Packages = nil
	for _, pkg := range packages {
		if !dependsOn(pkg, deps...) {
			continue
		}
		res.Packages = append(res.Packages, pkg)

		pkg.Worlds.All()(func(name string, w *wit.World) bool {
			if !dependsOn(w, deps...) {
				pkg.Worlds.Delete(name)
			}
			return true
		})

		pkg.Interfaces.All()(func(name string, i *wit.Interface) bool {
			if !dependsOn(i, deps...) {
				pkg.Interfaces.Delete(name)
			}
			return true
		})
	}

	// fmt.Printf("Resolve: %d %d %d %d\n\n",
	// 	len(res.Worlds), len(res.Interfaces), len(res.TypeDefs), len(res.Packages))

	return res
}

func dependsOn(node wit.Node, deps ...wit.Node) bool {
	for _, dep := range deps {
		// fmt.Printf("Does %T depend on %T?\n", node, dep)
		if wit.DependsOn(node, dep) || wit.DependsOn(dep, node) {
			return true
		}
	}
	return false
}
