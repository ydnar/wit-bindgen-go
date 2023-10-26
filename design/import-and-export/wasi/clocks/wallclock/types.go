package wallclock

// Interface is the Go implementation of WIT interface "wasi:clocks/wall-clock".
type Interface interface {
	Now() DateTime
	Resolution() DateTime
}

// DateTime is a Go implementation of WIT type "wasi:clocks/wall-clock.datetime".
type DateTime struct {
	Seconds      uint64
	Nanonseconds uint32
}
