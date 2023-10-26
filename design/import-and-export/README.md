# import-and-export

This design example contains a subset of the `wasi:clocks` package with a world that both *imports* and *exports* the `wasi:clocks/wall-clock` interface. The purpose of this example is to sketch out a mechanism for generating Go code whereby an interface (and types) can simultaneously be imported and exported.

The source WIT file is [world.wit](./world.wit).
