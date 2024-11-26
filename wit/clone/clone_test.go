package clone

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCloneRoundTrip(t *testing.T) {
	state := &State{}
	roundTrip(t, state, (*byte)(nil))
	roundTrip(t, state, pointerTo(0))
	roundTrip(t, state, pointerTo(uint16(0)))
	roundTrip(t, state, pointerTo(uint64(123)))
	roundTrip(t, state, pointerTo(""))
	roundTrip(t, state, pointerTo("hello"))
	roundTrip(t, state, &Example1{"hello"})
	roundTrip(t, state, pointerTo(fmt.Stringer(&Example1{"hello"})))
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

func roundTrip[T any](t *testing.T, state *State, v *T) {
	c := Clone(state, v)
	// fmt.Printf("Clone(%T): %T\n", v, c)
	if !reflect.DeepEqual(v, c) {
		t.Errorf("Clone(%#v): %#v", v, c)
	}
}

type Example1 struct {
	s string
}

func (e *Example1) String() string {
	return e.s
}

func (e *Example1) DeepClone(*State) Clonable {
	dst := *e
	return &dst
}

type Stringer string

func (s Stringer) String() string { return string(s) }

func report[T any](v T) {
	a := any(v)
	fmt.Printf("type of v: %T\n", a)
}

type Ints struct {
	a, b int
}

func (i *Ints) DeepClone(state *State) Clonable {
	c := *i
	return &c
}
