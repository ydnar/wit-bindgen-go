package json_test

import (
	"encoding/json"
	"reflect"
	"testing"

	wallclock "tests/generated/wasi/clocks/v0.2.0/wall-clock"
	"tests/generated/wasi/filesystem/v0.2.0/types"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		into    any
		want    any
		wantErr bool
	}{
		{
			"nil",
			`null`,
			ptr(ptr("")),
			ptr((*string)(nil)),
			false,
		},
		{
			"descriptor-type(block-device)",
			`"block-device"`,
			ptr(types.DescriptorType(0)),
			ptr(types.DescriptorTypeBlockDevice),
			false,
		},
		{
			"descriptor-type(directory)",
			`"directory"`,
			ptr(types.DescriptorType(0)),
			ptr(types.DescriptorTypeDirectory),
			false,
		},
		{
			"datetime",
			`{"seconds":1,"nanoseconds":2}`,
			&wallclock.DateTime{},
			&wallclock.DateTime{Seconds: 1, Nanoseconds: 2},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(tt.json), &tt.into)
			if tt.wantErr && err == nil {
				t.Errorf("json.Unmarshal(%q): expected error, got nil error", tt.json)
				return
			} else if !tt.wantErr && err != nil {
				t.Errorf("json.Unmarshal(%q): expected no error, got error: %v", tt.json, err)
				return
			}
			if !reflect.DeepEqual(tt.want, tt.into) {
				t.Errorf("json.Unmarshal(%q): resulting value different (%v != %v)", tt.json, tt.into, tt.want)
				return
			}
			got, err := json.Marshal(tt.into)
			if err != nil {
				t.Error(err)
				return
			}
			if string(got) != tt.json {
				t.Errorf("json.Marshal(%v): %s, expected %s", tt.into, string(got), tt.json)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
