package cabi

import (
	"sync"
	"unsafe"
)

// TODO: remove this or move it to package cm?

var (
	mu       sync.Mutex
	pointers = make(map[unsafe.Pointer]int)
)

// KeepAlive reference counts a pointer.
// TODO: prove this works.
func KeepAlive(ptr unsafe.Pointer) {
	mu.Lock()
	n := pointers[ptr]
	pointers[ptr] = n + 1
	mu.Unlock()
}

// Drop drops a reference to ptr.
// TODO: prove this works.
func Drop(ptr unsafe.Pointer) {
	mu.Lock()
	n := pointers[ptr]
	n -= 1
	// TODO: panic if n < 0?
	if n <= 0 {
		delete(pointers, ptr)
	}
	mu.Unlock()
}
