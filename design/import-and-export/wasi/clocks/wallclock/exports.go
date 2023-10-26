//go:build wasm

package wallclock

// Instance is the sole global instance of the
// WIT interface "wasi:clocks/wall-clock".
// Assign it to accept calls to this interface.
var Instance Interface

// FIXME: correct type for struct return values
//
//go:wasmexport wasi:clocks/wall-clock now
func wasmexport_now() DateTime {
	return Instance.Now()
}

// FIXME: correct type for struct return values
//
//go:wasmexport wasi:clocks/wall-clock resolution
func wasmexport_resolution() DateTime {
	return Instance.Resolution()
}
