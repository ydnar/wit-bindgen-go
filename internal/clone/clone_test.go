package clone

import (
	"fmt"
	"reflect"
	"testing"
)

type Example1 struct {
	s string
}

func (e *Example1) CloneWith(*State) *Example1 {
	dst := *e
	return &dst
}

func init() {
}

func TestCloneExample1(t *testing.T) {
	src := &Example1{"hello"}
	dst := Clone(&State{}, src)
	if !reflect.DeepEqual(src, dst) {
		t.Errorf("Clone: %#v != %#v", dst, src)
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
