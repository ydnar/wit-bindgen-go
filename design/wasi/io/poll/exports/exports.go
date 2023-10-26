package exports

import (
	"unsafe"

	"github.com/ydnar/wasm-tools-go/cabi"
	"github.com/ydnar/wasm-tools-go/cm"
)

// Interface implements the Component Model interface "wasi:io/poll".
type Interface interface {
	Poll(in cm.List[cabi.Borrow[Pollable]]) cm.List[uint32]
	Pollable() interface {
		// constructor, static functions would go here
	}
}

// Export registers a concrete implementation of "wasi:io/poll".
func Export(i Interface) {
	impl = i
}

// TODO: make a default implementation that panics with a helpful message on all function calls.
var impl Interface

//go:wasmexport wasi:io/poll#poll
func poll(in cm.List[cabi.Borrow[Pollable]], result *cm.List[uint32]) {
	*result = impl.Poll(in)
	// sData := unsafe.SliceData(s.Slice())
	// cabi.KeepAlive(unsafe.Pointer(sData))
}

//go:wasmexport wasi:io/poll#cabi_post_poll
func cabi_post_poll(result *cm.List[uint32]) {
	// Is this necessary if the Go GC runs after the wasmexport call?
	cabi.Drop(unsafe.Pointer(result.Data()))
}

//go:wasmexport wasi:io/poll#[method]pollable.block
func method_pollable_block(self cabi.Borrow[Pollable]) {
	self.Rep().Block()
}

//go:wasmexport wasi:io/poll#[method]pollable.ready
func method_pollable_ready(self cabi.Borrow[Pollable]) bool {
	return self.Rep().Ready()
}

// Pollable represents the Component Model type "wasi:io/poll.pollable".
type Pollable interface {
	Block()
	Ready() bool
	cabi.Resource[Pollable]
}
