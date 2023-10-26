//go:build wasm

package wallclock

// Now returns a DateTime for the current wall clock, corresponding to
// the Component Model function "wasi:clocks/wall-clock.now".
func Now() DateTime {
	return wasmimport_now()
}

// FIXME: correct type for struct return values
//
//go:wasmimport wasi:clocks/wall-clock now
func wasmimport_now() DateTime

// Resolution returns the resolution of the current wall clock, corresponding to
// the Component Model function "wasi:clocks/wall-clock.resolution".
func Resolution() DateTime {
	return wasmimport_resolution()
}

// FIXME: correct type for struct return values
//
//go:wasmimport wasi:clocks/wall-clock resolution
func wasmimport_resolution() DateTime
