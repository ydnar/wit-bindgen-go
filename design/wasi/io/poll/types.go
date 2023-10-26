package poll

import "github.com/ydnar/wasm-tools-go/cabi"

// Pollable represents the Component Model type "wasi:io/poll.pollable".
type Pollable interface {
	Block()
	Ready() bool
	cabi.Resource[Pollable]
}
