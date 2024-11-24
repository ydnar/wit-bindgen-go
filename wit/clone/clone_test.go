package clone

import (
	"fmt"
	"reflect"
	"testing"
)

type Example1 struct {
	s string
}

func (e *Example1) CloneWith(*State) Clonable {
	dst := *e
	return &dst
}

func init() {
}

func TestCloneRoundTrip(t *testing.T) {
	tests := []any{
		nil, error(nil), 0, uint16(0), uint64(123),
		"", "hello",
		Example1{"hello"},
		&Example1{"hello"},
	}
	state := &State{}
	for _, src := range tests {
		dst := Clone(state, &src)
		fmt.Printf("Clone(%T): %T\n", src, dst)
		if !reflect.DeepEqual(src, dst) {
			t.Errorf("Clone(%#v): %#v", src, dst)
		}
	}
	fmt.Printf("len(state.clones): %d\n", len(state.clones))
}

func TestCloneWideTypes(t *testing.T) {
	state := &State{}
	roundTrip(t, state, pointerTo(error(nil)))
	roundTrip(t, state, pointerTo(0))
	roundTrip(t, state, &struct{ a, b string }{})
	roundTrip(t, state, pointerTo(&struct{ a, b string }{}))
	roundTrip(t, state, &[3]uint64{})
	fmt.Printf("len(state.clones): %d\n", len(state.clones))
}

func pointerTo[T any](v T) *T {
	return &v
}

func roundTrip[T comparable](t *testing.T, state *State, v *T) {
	c := Clone(state, v)
	fmt.Printf("Clone(%T): %T\n", v, c)
	if !reflect.DeepEqual(v, c) {
		t.Errorf("Clone(%#v): %#v", v, c)
	}
}

func TestCloneInterface(t *testing.T) {
	var s fmt.Stringer = stringer("hello")
	report(s)
}

type stringer string

func (s stringer) String() string { return string(s) }

func report[T any](v T) {
	a := any(v)
	fmt.Printf("value of v: %T\n", a)
}

func TestCloneValue(t *testing.T) {
	v := &value{1, 2}
	c := Clone(&State{}, v)
	fmt.Printf("Clone(%v): %v\n", v, c)
}

type value struct {
	a, b int
}

func (v *value) CloneWith(state *State) Clonable {
	println("hello!")
	c := *v
	return &c
}
