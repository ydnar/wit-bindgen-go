package sleb128

import (
	"bytes"
	"math"
	"testing"
)

func TestReadWrite(t *testing.T) {
	tests := []int64{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		-1, -2, -3, -4, -5, -6, -7, -8, -9, -10,
		1 << 7, 1 << 8, 1 << 9,
		-1 << 7, -1 << 8, -1 << 9,
		math.MinInt64, math.MaxInt64,
	}
	for _, want := range tests {
		got, b, err := roundTrip(want)
		if err != nil {
			t.Errorf("roundTrip(%d): error: %v", want, err)
			continue
		}
		if got != want {
			t.Errorf("roundTrip(%d): got %d (%x)", want, got, b)
		}
	}
}

func roundTrip(v int64) (int64, []byte, error) {
	var buf bytes.Buffer
	_, err := Write(&buf, v)
	b := buf.Bytes()
	if err != nil {
		return 0, b, err
	}
	v, _, err = Read(bytes.NewReader(b))
	return v, b, err
}
