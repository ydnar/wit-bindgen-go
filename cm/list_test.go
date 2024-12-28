package cm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestListMethods(t *testing.T) {
	want := []byte("hello world")
	type myList List[uint8]
	l := myList(ToList(want))
	got := l.Slice()
	if !bytes.Equal(want, got) {
		t.Errorf("got (%s) != want (%s)", string(got), string(want))
	}
}

func TestListMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		w    listTester
	}{
		{
			name: "encode error",
			w:    newListEncoder(``, []errorEntry{{}}, true),
		},
		{
			name: "f32 nan",
			w:    newListEncoder(``, []float32{float32(math.NaN())}, true),
		},
		{
			name: "f64 nan",
			w:    newListEncoder(``, []float64{float64(math.NaN())}, true),
		},
		{
			name: "nil",
			w:    newListEncoder[string](`[]`, nil, false),
		},
		{
			name: "empty",
			w:    newListEncoder(`[]`, []string{}, false),
		},
		{
			name: "bool",
			w:    newListEncoder(`[true,false]`, []bool{true, false}, false),
		},
		{
			name: "string",
			w:    newListEncoder(`["one","two","three"]`, []string{"one", "two", "three"}, false),
		},
		{
			name: "char",
			w:    newListEncoder(`[104,105,127942]`, []rune{'h', 'i', '🏆'}, false),
		},
		{
			name: "s8",
			w:    newListEncoder(`[123,-123,127]`, []int8{123, -123, math.MaxInt8}, false),
		},
		{
			name: "u8",
			w:    newListEncoder(`[123,0,255]`, []uint8{123, 0, math.MaxUint8}, false),
		},
		{
			name: "s16",
			w:    newListEncoder(`[123,-123,32767]`, []int16{123, -123, math.MaxInt16}, false),
		},
		{
			name: "u16",
			w:    newListEncoder(`[123,0,65535]`, []uint16{123, 0, math.MaxUint16}, false),
		},
		{
			name: "s32",
			w:    newListEncoder(`[123,-123,2147483647]`, []int32{123, -123, math.MaxInt32}, false),
		},
		{
			name: "u32",
			w:    newListEncoder(`[123,0,4294967295]`, []uint32{123, 0, math.MaxUint32}, false),
		},
		{
			name: "s64",
			w:    newListEncoder(`[123,-123,9223372036854775807]`, []int64{123, -123, math.MaxInt64}, false),
		},
		{
			name: "u64",
			w:    newListEncoder(`[123,0,18446744073709551615]`, []uint64{123, 0, math.MaxUint64}, false),
		},
		{
			name: "f32",
			w:    newListEncoder(`[1.01,2,3.4028235e+38]`, []float32{1.01, 2, math.MaxFloat32}, false),
		},
		{
			name: "f64",
			w:    newListEncoder(`[1.01,2,1.7976931348623157e+308]`, []float64{1.01, 2, math.MaxFloat64}, false),
		},
		{
			name: "struct",
			w:    newListEncoder(`[{"name":"joe","age":10},{"name":"jane","age":20}]`, []testEntry{{Name: "joe", Age: 10}, {Name: "jane", Age: 20}}, false),
		},
		{
			name: "list",
			w:    newListEncoder(`[["one","two","three"],["four","five","six"]]`, []List[string]{ToList([]string{"one", "two", "three"}), ToList([]string{"four", "five", "six"})}, false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// NOTE(lxf): skip marshal errors in tinygo as it uses 'defer'
			// needs tinygo 0.35-dev
			if tt.w.wantErr() && runtime.Compiler == "tinygo" && strings.Contains(runtime.GOARCH, "wasm") {
				return
			}

			data, err := json.Marshal(tt.w.outer())
			if err != nil {
				if tt.w.wantErr() {
					return
				}

				t.Fatal(err)
			}

			if tt.w.wantErr() {
				t.Fatalf("expect error, but got none. got (%s)", string(data))
			}

			if got, want := data, []byte(tt.w.rawData()); !bytes.Equal(got, want) {
				t.Errorf("got (%v) != want (%v)", string(got), string(want))
			}
		})
	}
}

func TestListUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		w    listTester
	}{
		{
			name: "decode error",
			w:    newListDecoder(`["joe"]`, []errorEntry{}, true),
		},
		{
			name: "invalid json",
			w:    newListDecoder(`[joe]`, []string{}, true),
		},
		{
			name: "incompatible type",
			w:    newListDecoder(`[123,456]`, []string{}, true),
		},
		{
			name: "incompatible bool",
			w:    newListDecoder(`["true","false"]`, []bool{true, false}, true),
		},
		{
			name: "incompatible s32",
			w:    newListDecoder(`["123","-123","2147483647"]`, []int32{}, true),
		},
		{
			name: "incompatible u32",
			w:    newListDecoder(`["123","0","4294967295"]`, []uint32{}, true),
		},

		{
			name: "null",
			w:    newListDecoder[string](`null`, nil, false),
		},
		{
			name: "empty",
			w:    newListDecoder(`[]`, []string{}, false),
		},
		{
			name: "bool",
			w:    newListDecoder(`[true,false]`, []bool{true, false}, false),
		},
		{
			name: "string",
			w:    newListDecoder(`["one","two","three"]`, []string{"one", "two", "three"}, false),
		},
		{
			name: "char",
			w:    newListDecoder(`[104,105,127942]`, []rune{'h', 'i', '🏆'}, false),
		},
		{
			name: "s8",
			w:    newListDecoder(`[123,-123,127]`, []int8{123, -123, math.MaxInt8}, false),
		},
		{
			name: "u8",
			w:    newListDecoder(`[123,0,255]`, []uint8{123, 0, math.MaxUint8}, false),
		},
		{
			name: "s16",
			w:    newListDecoder(`[123,-123,32767]`, []int16{123, -123, math.MaxInt16}, false),
		},
		{
			name: "u16",
			w:    newListDecoder(`[123,0,65535]`, []uint16{123, 0, math.MaxUint16}, false),
		},
		{
			name: "s32",
			w:    newListDecoder(`[123,-123,2147483647]`, []int32{123, -123, math.MaxInt32}, false),
		},
		{
			name: "u32",
			w:    newListDecoder(`[123,0,4294967295]`, []uint32{123, 0, math.MaxUint32}, false),
		},
		{
			name: "s64",
			w:    newListDecoder(`[123,-123,9223372036854775807]`, []int64{123, -123, math.MaxInt64}, false),
		},
		{
			name: "u64",
			w:    newListDecoder(`[123,0,18446744073709551615]`, []uint64{123, 0, math.MaxUint64}, false),
		},
		{
			name: "f32",
			w:    newListDecoder(`[1.01,2,3.4028235e+38]`, []float32{1.01, 2, math.MaxFloat32}, false),
		},
		{
			name: "f32 nan",
			w:    newListDecoder(`[null]`, []float32{0}, false),
		},
		{
			name: "f64",
			w:    newListDecoder(`[1.01,2,1.7976931348623157e+308]`, []float64{1.01, 2, math.MaxFloat64}, false),
		},
		{
			name: "f64 nan",
			w:    newListDecoder(`[null]`, []float64{0}, false),
		},
		{
			name: "struct",
			w:    newListDecoder(`[{"name":"joe","age":10},{"name":"jane","age":20}]`, []testEntry{{Name: "joe", Age: 10}, {Name: "jane", Age: 20}}, false),
		},
		{
			name: "list",
			w:    newListDecoder(`[["one","two","three"],["four","five","six"]]`, []List[string]{ToList([]string{"one", "two", "three"}), ToList([]string{"four", "five", "six"})}, false),
		},
		// tuple, result, option, and variant needs json implementation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(tt.w.rawData()), tt.w.outer())
			if err != nil {
				if tt.w.wantErr() {
					return
				}

				t.Fatal(err)
			}

			if tt.w.wantErr() {
				t.Fatalf("expect error, but got none. got (%v)", tt.w.outerSlice())
			}

			if got, want := tt.w.outerSlice(), tt.w.inner(); !reflect.DeepEqual(got, want) {
				t.Errorf("got (%v) != want (%v)", got, want)
			}
		})
	}
}

type listTester interface {
	outer() any
	inner() any
	outerSlice() any
	wantErr() bool
	rawData() string
}

type listWrapper[T comparable] struct {
	raw       string
	outerList List[T]
	innerList []T
	err       bool
}

func (w *listWrapper[T]) wantErr() bool {
	return w.err
}

func (w *listWrapper[T]) outer() any {
	return &w.outerList
}

func (w *listWrapper[T]) outerSlice() any {
	return w.outerList.Slice()
}

func (w *listWrapper[T]) inner() any {
	return w.innerList
}

func (w *listWrapper[T]) rawData() string {
	return w.raw
}

func newListEncoder[T comparable](raw string, want []T, wantErr bool) *listWrapper[T] {
	return &listWrapper[T]{raw: raw, outerList: ToList(want), err: wantErr}
}

func newListDecoder[T comparable](raw string, want []T, wantErr bool) *listWrapper[T] {
	return &listWrapper[T]{raw: raw, innerList: want, err: wantErr}
}

type testEntry struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type errorEntry struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (errorEntry) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("can't encode")
}

func (*errorEntry) UnmarshalJSON(_ []byte) error {
	return fmt.Errorf("can't decode")
}
