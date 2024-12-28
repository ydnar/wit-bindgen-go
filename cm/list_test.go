package cm

import (
	"bytes"
	"encoding/json"
	"errors"
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
			w:    listMarshalTest(``, []errorEntry{{}}, true),
		},
		{
			name: "f32 nan",
			w:    listMarshalTest(``, []float32{float32(math.NaN())}, true),
		},
		{
			name: "f64 nan",
			w:    listMarshalTest(``, []float64{float64(math.NaN())}, true),
		},
		{
			name: "nil",
			w:    listMarshalTest[string](`[]`, nil, false),
		},
		{
			name: "empty",
			w:    listMarshalTest(`[]`, []string{}, false),
		},
		{
			name: "bool",
			w:    listMarshalTest(`[true,false]`, []bool{true, false}, false),
		},
		{
			name: "string",
			w:    listMarshalTest(`["one","two","three"]`, []string{"one", "two", "three"}, false),
		},
		{
			name: "char",
			w:    listMarshalTest(`[104,105,127942]`, []rune{'h', 'i', '🏆'}, false),
		},
		{
			name: "s8",
			w:    listMarshalTest(`[123,-123,127]`, []int8{123, -123, math.MaxInt8}, false),
		},
		{
			name: "u8",
			w:    listMarshalTest(`[123,0,255]`, []uint8{123, 0, math.MaxUint8}, false),
		},
		{
			name: "s16",
			w:    listMarshalTest(`[123,-123,32767]`, []int16{123, -123, math.MaxInt16}, false),
		},
		{
			name: "u16",
			w:    listMarshalTest(`[123,0,65535]`, []uint16{123, 0, math.MaxUint16}, false),
		},
		{
			name: "s32",
			w:    listMarshalTest(`[123,-123,2147483647]`, []int32{123, -123, math.MaxInt32}, false),
		},
		{
			name: "u32",
			w:    listMarshalTest(`[123,0,4294967295]`, []uint32{123, 0, math.MaxUint32}, false),
		},
		{
			name: "s64",
			w:    listMarshalTest(`[123,-123,9223372036854775807]`, []int64{123, -123, math.MaxInt64}, false),
		},
		{
			name: "u64",
			w:    listMarshalTest(`[123,0,18446744073709551615]`, []uint64{123, 0, math.MaxUint64}, false),
		},
		{
			name: "f32",
			w:    listMarshalTest(`[1.01,2,3.4028235e+38]`, []float32{1.01, 2, math.MaxFloat32}, false),
		},
		{
			name: "f64",
			w:    listMarshalTest(`[1.01,2,1.7976931348623157e+308]`, []float64{1.01, 2, math.MaxFloat64}, false),
		},
		{
			name: "struct",
			w:    listMarshalTest(`[{"name":"joe","age":10},{"name":"jane","age":20}]`, []testEntry{{Name: "joe", Age: 10}, {Name: "jane", Age: 20}}, false),
		},
		{
			name: "list",
			w:    listMarshalTest(`[["one","two","three"],["four","five","six"]]`, []List[string]{ToList([]string{"one", "two", "three"}), ToList([]string{"four", "five", "six"})}, false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// NOTE(lxf): skip marshal errors in tinygo as it uses 'defer'
			// needs tinygo 0.35-dev
			if tt.w.WantErr() && runtime.Compiler == "tinygo" && strings.Contains(runtime.GOARCH, "wasm") {
				return
			}

			data, err := json.Marshal(tt.w.List())
			if err != nil {
				if tt.w.WantErr() {
					return
				}

				t.Error(err)
				return
			}

			if tt.w.WantErr() {
				t.Errorf("expected error, but got none. got (%s)", string(data))
				return
			}

			if got, want := data, tt.w.JSON(); !bytes.Equal(got, want) {
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
			w:    listUnmarshalTest(`["joe"]`, []errorEntry{}, true),
		},
		{
			name: "invalid json",
			w:    listUnmarshalTest(`[joe]`, []string{}, true),
		},
		{
			name: "incompatible type",
			w:    listUnmarshalTest(`[123,456]`, []string{}, true),
		},
		{
			name: "incompatible bool",
			w:    listUnmarshalTest(`["true","false"]`, []bool{true, false}, true),
		},
		{
			name: "incompatible s32",
			w:    listUnmarshalTest(`["123","-123","2147483647"]`, []int32{}, true),
		},
		{
			name: "incompatible u32",
			w:    listUnmarshalTest(`["123","0","4294967295"]`, []uint32{}, true),
		},

		{
			name: "null",
			w:    listUnmarshalTest[string](`null`, nil, false),
		},
		{
			name: "empty",
			w:    listUnmarshalTest(`[]`, []string{}, false),
		},
		{
			name: "bool",
			w:    listUnmarshalTest(`[true,false]`, []bool{true, false}, false),
		},
		{
			name: "string",
			w:    listUnmarshalTest(`["one","two","three"]`, []string{"one", "two", "three"}, false),
		},
		{
			name: "char",
			w:    listUnmarshalTest(`[104,105,127942]`, []rune{'h', 'i', '🏆'}, false),
		},
		{
			name: "s8",
			w:    listUnmarshalTest(`[123,-123,127]`, []int8{123, -123, math.MaxInt8}, false),
		},
		{
			name: "u8",
			w:    listUnmarshalTest(`[123,0,255]`, []uint8{123, 0, math.MaxUint8}, false),
		},
		{
			name: "s16",
			w:    listUnmarshalTest(`[123,-123,32767]`, []int16{123, -123, math.MaxInt16}, false),
		},
		{
			name: "u16",
			w:    listUnmarshalTest(`[123,0,65535]`, []uint16{123, 0, math.MaxUint16}, false),
		},
		{
			name: "s32",
			w:    listUnmarshalTest(`[123,-123,2147483647]`, []int32{123, -123, math.MaxInt32}, false),
		},
		{
			name: "u32",
			w:    listUnmarshalTest(`[123,0,4294967295]`, []uint32{123, 0, math.MaxUint32}, false),
		},
		{
			name: "s64",
			w:    listUnmarshalTest(`[123,-123,9223372036854775807]`, []int64{123, -123, math.MaxInt64}, false),
		},
		{
			name: "u64",
			w:    listUnmarshalTest(`[123,0,18446744073709551615]`, []uint64{123, 0, math.MaxUint64}, false),
		},
		{
			name: "f32",
			w:    listUnmarshalTest(`[1.01,2,3.4028235e+38]`, []float32{1.01, 2, math.MaxFloat32}, false),
		},
		{
			name: "f32 nan",
			w:    listUnmarshalTest(`[null]`, []float32{0}, false),
		},
		{
			name: "f64",
			w:    listUnmarshalTest(`[1.01,2,1.7976931348623157e+308]`, []float64{1.01, 2, math.MaxFloat64}, false),
		},
		{
			name: "f64 nan",
			w:    listUnmarshalTest(`[null]`, []float64{0}, false),
		},
		{
			name: "struct",
			w:    listUnmarshalTest(`[{"name":"joe","age":10},{"name":"jane","age":20}]`, []testEntry{{Name: "joe", Age: 10}, {Name: "jane", Age: 20}}, false),
		},
		{
			name: "list",
			w:    listUnmarshalTest(`[["one","two","three"],["four","five","six"]]`, []List[string]{ToList([]string{"one", "two", "three"}), ToList([]string{"four", "five", "six"})}, false),
		},
		// tuple, result, option, and variant needs json implementation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal(tt.w.JSON(), tt.w.List())
			if err != nil {
				if tt.w.WantErr() {
					return
				}

				t.Error(err)
				return
			}

			if tt.w.WantErr() {
				t.Errorf("expected error, but got none. got (%v)", tt.w.Slice())
				return
			}

			if got, want := tt.w.Slice(), tt.w.WantSlice(); !reflect.DeepEqual(got, want) {
				t.Errorf("got (%v) != want (%v)", got, want)
			}
		})
	}
}

type listTester interface {
	List() any
	WantSlice() any
	Slice() any
	WantErr() bool
	JSON() []byte
}

type listWrapper[T comparable] struct {
	json    string
	list    List[T]
	slice   []T
	wantErr bool
}

func (w *listWrapper[T]) WantErr() bool {
	return w.wantErr
}

func (w *listWrapper[T]) List() any {
	return &w.list
}

func (w *listWrapper[T]) Slice() any {
	return w.list.Slice()
}

func (w *listWrapper[T]) WantSlice() any {
	return w.slice
}

func (w *listWrapper[T]) JSON() []byte {
	return []byte(w.json)
}

func listMarshalTest[T comparable](json string, want []T, wantErr bool) *listWrapper[T] {
	return &listWrapper[T]{json: json, list: ToList(want), wantErr: wantErr}
}

func listUnmarshalTest[T comparable](json string, want []T, wantErr bool) *listWrapper[T] {
	return &listWrapper[T]{json: json, slice: want, wantErr: wantErr}
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
	return nil, errors.New("MarshalJSON")
}

func (*errorEntry) UnmarshalJSON(_ []byte) error {
	return errors.New("UnmarshalJSON")
}
