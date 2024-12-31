package json_test

import (
	"encoding/json"
	"testing"

	wallclock "tests/generated/wasi/clocks/v0.2.0/wall-clock"
	"tests/generated/wasi/filesystem/v0.2.0/types"
)

func TestRecordJSON(t *testing.T) {
	var dt wallclock.DateTime
	_ = dt
}

func TestEnumMarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		v       any
		want    string
		wantErr bool
	}{
		{"nil", nil, `null`, false},
		{"descriptor-type(directory)", types.DescriptorTypeDirectory, `"directory"`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.v)
			if tt.wantErr && err == nil {
				t.Errorf("json.Marshal(%v): expected error, got nil error", tt.v)
				return
			} else if !tt.wantErr && err != nil {
				t.Errorf("json.Marshal(%v): expected no error, got error: %v", tt.v, err)
				return
			}
			if string(got) != tt.want {
				t.Errorf("json.Marshal(%v): %s, expected %s", tt.v, string(got), tt.want)
			}
		})
	}
}
