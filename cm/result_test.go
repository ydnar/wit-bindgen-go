package cm

import (
	"runtime"
	"testing"
	"unsafe"
)

var (
	_ resulter[string, bool] = &Result[string, string, bool]{}
	_ resulter[bool, string] = &Result[string, bool, string]{}
)

type resulter[OK, Err any] interface {
	IsOK() bool
	IsErr() bool
	OK() *OK
	Err() *Err
	Result() (OK, Err, bool)
}

func TestResultOKOrErr(t *testing.T) {
	r1 := OK[Result[string, string, struct{}]]("hello")
	if ok := r1.OK(); ok == nil {
		t.Errorf("OK(): %v, expected non-nil OK", ok)
	}
	if err := r1.Err(); err != nil {
		t.Errorf("Err(): %v, expected nil Err", err)
	}

	r2 := Err[Result[bool, struct{}, bool]](true)
	if ok := r2.OK(); ok != nil {
		t.Errorf("OK(): %v, expected nil OK", ok)
	}
	if err := r2.Err(); err == nil {
		t.Errorf("Err(): %v, expected non-nil Err", err)
	}
}

func TestResultResult(t *testing.T) {
	ok, err, isErr := OK[Result[string, string, int]]("hello").Result()
	if got, want := ok, "hello"; got != want {
		t.Errorf("Result(): ok = %v; expected %v", got, want)
	}
	if got, want := err, 0; got != want {
		t.Errorf("Result(): err = %v; expected %v", got, want)
	}
	if got, want := isErr, false; got != want {
		t.Errorf("Result(): isErr = %v; expected %v", got, want)
	}
	ok, err, isErr = Err[Result[string, string, int]](42).Result()
	if got, want := ok, ""; got != want {
		t.Errorf("Result(): ok = %v; expected %v", got, want)
	}
	if got, want := err, 42; got != want {
		t.Errorf("Result(): err = %v; expected %v", got, want)
	}
	if got, want := isErr, true; got != want {
		t.Errorf("Result(): isErr = %v; expected %v", got, want)
	}
}

func TestResultLayout(t *testing.T) {
	// 8 on 64-bit, 4 on 32-bit
	ptrSize := unsafe.Sizeof(uintptr(0))

	tests := []struct {
		name   string
		r      ResultDebug
		size   uintptr
		offset uintptr
	}{
		{"result", BoolResult(false), 1, 0},
		{"ok", BoolResult(ResultOK), 1, 0},
		{"err", BoolResult(ResultErr), 1, 0},

		{"result<string, string>", Result[string, string, string]{}, sizePlusAlignOf[string](), ptrSize},
		{"result<bool, string>", Result[string, bool, string]{}, sizePlusAlignOf[string](), ptrSize},
		{"result<string, _>", Result[string, string, struct{}]{}, sizePlusAlignOf[string](), ptrSize},
		{"result<_, string>", Result[string, struct{}, string]{}, sizePlusAlignOf[string](), ptrSize},
		{"result<u64, u64>", Result[uint64, uint64, uint64]{}, 16, alignOf[uint64]()},
		{"result<u32, u64>", Result[uint64, uint32, uint64]{}, 16, alignOf[uint64]()},
		{"result<u64, u32>", Result[uint64, uint64, uint32]{}, 16, alignOf[uint64]()},
		{"result<u8, u64>", Result[uint64, uint8, uint64]{}, 16, alignOf[uint64]()},
		{"result<u64, u8>", Result[uint64, uint64, uint8]{}, 16, alignOf[uint64]()},
		{"result<u8, u32>", Result[uint32, uint8, uint32]{}, 8, alignOf[uint32]()},
		{"result<u32, u8>", Result[uint32, uint32, uint8]{}, 8, alignOf[uint32]()},
		{"result<[9]u8, u64>", Result[[9]byte, [9]byte, uint64]{}, 24, alignOf[uint64]()},

		{"result<string, _>", Result[string, string, struct{}]{}, sizePlusAlignOf[string](), ptrSize},
		{"result<string, _>", Result[string, string, struct{}]{}, sizePlusAlignOf[string](), ptrSize},
		{"result<string, bool>", Result[string, string, bool]{}, sizePlusAlignOf[string](), ptrSize},
		{"result<[9]u8, u64>", Result[[9]byte, [9]byte, uint64]{}, 24, alignOf[uint64]()},

		{"result<_, string>", Result[string, struct{}, string]{}, sizePlusAlignOf[string](), ptrSize},
		{"result<_, string>", Result[string, struct{}, string]{}, sizePlusAlignOf[string](), ptrSize},
		{"result<bool, string>", Result[string, bool, string]{}, sizePlusAlignOf[string](), ptrSize},
		{"result<u64, [9]u8>", Result[[9]byte, uint64, [9]byte]{}, 24, alignOf[uint64]()},
	}

	for _, tt := range tests {
		typ := typeName(tt.r)
		t.Run(tt.name, func(t *testing.T) {
			if got, want := tt.r.Size(), tt.size; got != want {
				t.Errorf("(%s).Size(): %v, expected %v", typ, got, want)
			}
			if got, want := tt.r.DataOffset(), tt.offset; got != want {
				t.Errorf("(%s).DataOffset(): %v, expected %v", typ, got, want)
			}
		})
	}
}

func TestAltResult1(t *testing.T) {
	type alt1[Shape, OK, Err any] struct {
		_     [0]OK
		_     [0]Err
		data  Shape
		isErr bool
	}

	equalSize(t, result[uint8, struct{}, uint8]{}, alt1[uint8, struct{}, uint8]{})
	equalSize(t, result[uint16, struct{}, uint16]{}, alt1[uint16, struct{}, uint16]{})
	equalSize(t, result[uint32, struct{}, uint32]{}, alt1[uint32, struct{}, uint32]{})
	equalSize(t, result[uint64, struct{}, uint64]{}, alt1[uint64, struct{}, uint64]{})
	equalSize(t, result[uint64, [5]uint8, uint64]{}, alt1[uint64, [5]uint8, uint64]{})
	equalSize(t, result[uint64, [6]uint8, uint64]{}, alt1[uint64, [6]uint8, uint64]{})

	// Go adds padding to structs with zero-length trailing fields.
	// TinyGo does not.
	if runtime.Compiler != "tinygo" {
		unequalSize(t, result[struct{}, struct{}, struct{}]{}, alt1[struct{}, struct{}, struct{}]{})

		// zero-length arrays have alignment of their associated type
		// TODO: document why zero-length arrays are not allowed as result or variant associated types
		unequalSize(t, result[[0]uint64, [0]uint64, struct{}]{}, alt1[[0]uint64, [0]uint64, struct{}]{})
	}
}

func equalSize[A, B any](t *testing.T, a A, b B) {
	if unsafe.Sizeof(a) != unsafe.Sizeof(b) {
		t.Errorf("unsafe.Sizeof(%T) (%d) != unsafe.Sizeof(%T) (%d)", a, unsafe.Sizeof(a), b, unsafe.Sizeof(b))
	}
}

func unequalSize[A, B any](t *testing.T, a A, b B) {
	if unsafe.Sizeof(a) == unsafe.Sizeof(b) {
		t.Errorf("unsafe.Sizeof(%T) (%d) == unsafe.Sizeof(%T) (%d)", a, unsafe.Sizeof(a), b, unsafe.Sizeof(b))
	}
}

func BenchmarkResultInlines(b *testing.B) {
	var ok *struct{}
	var err *string
	r1 := Err[Result[string, struct{}, string]]("hello")
	for i := 0; i < b.N; i++ {
		ok = r1.OK()
	}
	_ = ok
	_ = err
}

func TestIssue95String(t *testing.T) {
	type (
		magic struct {
			data *byte
			len  [unsafe.Sizeof(uintptr(0))]byte
		}
		stringVariant Variant[uint8, string, string]
		// stringVariant Variant[uint8, [unsafe.Sizeof("")]byte, string]
		// stringVariant Variant[uint8, magic, string]
		// stringResult Result[stringVariant, string, stringVariant]
		stringResult Result[[unsafe.Sizeof(*(*stringVariant)(nil))]byte, string, stringVariant]
	)

	want := "hello"
	res := OK[stringResult](want)
	got := *res.OK()
	// fmt.Printf("unsafe.Sizeof(res): %d\n", unsafe.Sizeof(res))
	// fmt.Printf("got: %v (%d) want: %v (%d)\n",
	// 	unsafe.StringData(got), len(got), unsafe.StringData(want), len(want))
	if got != want {
		t.Errorf("*res.OK(): %v, expected %v", got, want)
	}
}

func TestIssue95Uint64(t *testing.T) {
	type (
		uint64Variant Variant[uint8, uint64, uint64]
		// uint64Variant Variant[uint8, [unsafe.Sizeof(uint64(0))]byte, uint64]
		// uint64Result Result[uint64Variant, uint64, uint64Variant]
		uint64Result Result[[unsafe.Sizeof(uint64Variant{})]byte, uint64, uint64Variant]
	)

	want := uint64(123)
	res := OK[uint64Result](want)
	got := *res.OK()
	// fmt.Printf("unsafe.Sizeof(res): %d\n", unsafe.Sizeof(res))
	// fmt.Printf("got: %v want: %v\n", got, want)
	if got != want {
		t.Errorf("*res.OK(): %v, expected %v", got, want)
	}
}

func TestIssue95Struct(t *testing.T) {
	type (
		// structResult  Result[structVariant, stringStruct, structVariant]
		stringStruct struct {
			_ HostLayout
			// i int
			s string
		}
		structVariant Variant[uint8, stringStruct, stringStruct]
		// structVariant Variant[uint8, [1]stringStruct, [2]stringStruct]
		// structResult Result[structVariant, stringStruct, structVariant]
		structResult Result[[unsafe.Sizeof(*(*structVariant)(nil))]byte, stringStruct, structVariant]
		// structResult Result[[2]uintptr, stringStruct, structVariant]
	)

	want := stringStruct{s: "hello"}
	res := OK[structResult](want)
	got := *res.OK()
	// fmt.Printf("unsafe.Sizeof(res): %d\n", unsafe.Sizeof(res))
	// fmt.Printf("got: %v want: %v\n", got, want)
	if got != want {
		t.Errorf("*res.OK(): %v, expected %v", got, want)
	}
}

func TestIssue95BoolInt64(t *testing.T) {
	type boolInt64Result Result[int64, bool, int64]
	want := int64(1234567890)
	res := Err[boolInt64Result](1234567890)
	got := *res.Err()
	// fmt.Printf("unsafe.Sizeof(res): %d\n", unsafe.Sizeof(res))
	// fmt.Printf("got: %v want: %v\n", got, want)
	if got != want {
		t.Errorf("*res.OK(): %v, expected %v", got, want)
	}
}
