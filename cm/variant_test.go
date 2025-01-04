package cm

import (
	"runtime"
	"strings"
	"testing"
	"unsafe"
)

func TestVariantLayout(t *testing.T) {
	// 8 on 64-bit, 4 on 32-bit
	ptrSize := unsafe.Sizeof(uintptr(0))

	tests := []struct {
		name   string
		v      VariantDebug
		size   uintptr
		offset uintptr
	}{
		{"variant { string; string }", Variant[uint8, string, string]{}, sizePlusAlignOf[string](), ptrSize},
		{"variant { bool; string }", Variant[uint8, string, bool]{}, sizePlusAlignOf[string](), ptrSize},
		{"variant { string; _ }", Variant[uint8, string, string]{}, sizePlusAlignOf[string](), ptrSize},
		{"variant { _; _ }", Variant[uint8, string, struct{}]{}, sizePlusAlignOf[string](), ptrSize},
		{"variant { u64; u64 }", Variant[uint8, uint64, uint64]{}, 16, alignOf[uint64]()},
		{"variant { u32; u64 }", Variant[uint8, uint64, uint32]{}, 16, alignOf[uint64]()},
		{"variant { u64; u32 }", Variant[uint8, uint64, uint32]{}, 16, alignOf[uint64]()},
		{"variant { u8; u64 }", Variant[uint8, uint64, uint8]{}, 16, alignOf[uint64]()},
		{"variant { u64; u8 }", Variant[uint8, uint64, uint8]{}, 16, alignOf[uint64]()},
		{"variant { u8; u32 }", Variant[uint8, uint32, uint8]{}, 8, alignOf[uint32]()},
		{"variant { u32; u8 }", Variant[uint8, uint32, uint8]{}, 8, alignOf[uint32]()},
		{"variant { [9]u8, u64 }", Variant[uint8, [9]byte, uint64]{}, 24, alignOf[uint64]()},
	}

	for _, tt := range tests {
		typ := typeName(tt.v)
		t.Run(tt.name, func(t *testing.T) {
			if got, want := tt.v.Size(), tt.size; got != want {
				t.Errorf("(%s).Size(): %v, expected %v", typ, got, want)
			}
			if got, want := tt.v.DataOffset(), tt.offset; got != want {
				t.Errorf("(%s).DataOffset(): %v, expected %v", typ, got, want)
			}
		})
	}
}

func TestGetValidates(t *testing.T) {
	if runtime.Compiler == "tinygo" && strings.Contains(runtime.GOARCH, "wasm") {
		return
	}
	defer func() {
		if recover() == nil {
			t.Errorf("Get did not panic")
		}
	}()
	var v Variant[uint8, uint8, uint8]
	_ = Case[string](&v, 0)
}

func TestNewVariantValidates(t *testing.T) {
	if runtime.Compiler == "tinygo" && strings.Contains(runtime.GOARCH, "wasm") {
		return
	}
	defer func() {
		if recover() == nil {
			t.Errorf("NewVariant did not panic")
		}
	}()
	_ = NewVariant[uint8, uint8, uint8](0, "hello world")
}
