package uleb128

import (
	"bytes"
	"math"
	"testing"
)

func TestReadWrite(t *testing.T) {
	var i uint64
	for {
		got, b, err := roundTrip(i)
		if err != nil {
			t.Errorf("roundTrip(%d): error: %v", i, err)
			continue
		}
		if got != i {
			t.Errorf("roundTrip(%d): got %d (%x)", i, got, b)
		}
		switch {
		case i < 1<<17:
			i++
		case i < 1<<22:
			i += 99991
		case i < 1<<36:
			i += 999331
		case i == math.MaxUint64:
			return
		default:
			i = math.MaxUint64
		}
	}
}

func roundTrip(v uint64) (uint64, []byte, error) {
	var buf bytes.Buffer
	_, err := Write(&buf, v)
	b := buf.Bytes()
	if err != nil {
		return 0, b, err
	}
	v, _, err = Read(bytes.NewReader(b))
	return v, b, err
}
