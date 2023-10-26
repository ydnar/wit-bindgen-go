//go:build wasm

package imports

import (
	"unsafe"

	"github.com/ydnar/wasm-tools-go/cabi"
	"github.com/ydnar/wasm-tools-go/design/wasi/io/poll"
)

// Pollable represents the Component Model type "wasi:io/poll.pollable".
type Pollable cabi.Handle[poll.Pollable]

var _ poll.Pollable = Pollable(0)

// Poll imports Component Model func "wasi:io/poll.poll".
func Poll(in []poll.Pollable) []uint32 {
	in_ := make([]Pollable, len(in))
	for i := range in {
		in_[i] = Pollable(in[i].ResourceHandle())
	}
	ptr, size := wasmimport_poll(unsafe.SliceData(in_), uint32(len(in_)))
	return unsafe.Slice(ptr, size)
}

func (self Pollable) ResourceHandle() cabi.Handle[poll.Pollable] {
	return cabi.Handle[poll.Pollable](self)
}

//go:wasmimport wasi:io/poll poll
func wasmimport_poll(data *Pollable, size uint32) (*uint32, int32)

// Block imports Component Model method "wasi:io/poll.pollable.block".
func (self Pollable) Block() {
	wasmimport_pollable_block(self)
}

//go:wasmimport wasi:io/poll [method]pollable.block
func wasmimport_pollable_block(self Pollable)

// Ready imports Component Model method "wasi:io/poll.pollable.ready".
func (self Pollable) Ready() bool {
	return wasmimport_pollable_ready(self)
}

//go:wasmimport wasi:io/poll [method]pollable.ready
func wasmimport_pollable_ready(self Pollable) bool
