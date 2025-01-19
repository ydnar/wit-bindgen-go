package issue284

import (
	"fmt"
	"testing"

	"tests/generated/issues/issue284/i"

	"go.bytecodealliance.org/cm"
)

func TestIssue284(t *testing.T) {
	tests := []struct {
		isErr bool
		ok    bool
		err   int8
	}{
		{false, false, 0},
		{false, true, 0},
		{true, false, 0},
		{true, false, 1},
		{true, false, 5},
		{true, false, 126},
		{true, false, 127},
	}
	for _, tt := range tests {
		if !tt.isErr {
			t.Run(fmt.Sprintf("cm.OK[i.BoolS8Result](%t)", tt.ok), func(t *testing.T) {
				r := cm.OK[i.BoolS8Result](tt.ok)
				ok, _, isErr := r.Result()
				if isErr != tt.isErr {
					t.Errorf("isErr == %t, expected %t", isErr, tt.isErr)
				}
				if ok != tt.ok {
					t.Errorf("ok == %t, expected %t", ok, tt.ok)
				}
			})
		} else {
			t.Run(fmt.Sprintf("cm.Err[i.BoolS8Result](%d)", tt.err), func(t *testing.T) {
				r := cm.Err[i.BoolS8Result](tt.err)
				_, err, isErr := r.Result()
				if isErr != tt.isErr {
					t.Errorf("isErr == %t, expected %t", isErr, tt.isErr)
				}
				if err != tt.err {
					t.Errorf("err == %d, expected %d", err, tt.err)
				}
			})
		}
	}
}
