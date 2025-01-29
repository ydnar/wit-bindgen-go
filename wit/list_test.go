package wit

import "testing"

// https://github.com/bytecodealliance/go-modules/issues/288
func TestListAlign(t *testing.T) {
	var l List
	got, want := l.Align(), uintptr(4)
	if got != want {
		t.Errorf("ABI alignment for list<T> is %d, expected %d", got, want)
	}
}
