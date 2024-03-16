package exports_test

import (
	"strconv"

	"github.com/ydnar/wasm-tools-go/cm"
	"github.com/ydnar/wasm-tools-go/design/example/resources/simple/exports"
)

// Value representation
var (
	_ exports.Interface       = interface1{}
	_ exports.NumberInterface = number1(0)
)

type interface1 struct{}

func (i interface1) Number(rep cm.Rep) exports.NumberInterface {
	return cm.AsRep[number1](rep)
}

func (i interface1) NewNumber(value int32) exports.Number {
	return exports.NumberResourceNew(number1(value))
}

func (i interface1) NumberMerge(a exports.Number, b exports.Number) exports.Number {
	return i.NewNumber(a.ResourceRep().Value() + b.ResourceRep().Value())
}

func (i interface1) NumberChoose(a exports.NumberInterface, b exports.NumberInterface) exports.Number {
	return i.NewNumber(a.Value() + b.Value())
}

type number1 int32

func (n number1) Value() int32        { return int32(n) }
func (n number1) String() string      { return strconv.Itoa(int(n)) }
func (n number1) ResourceDestructor() {}
func (n number1) ResourceRep() cm.Rep { return 0 /* FIXME */ }

// Pointer representation
var (
	_ exports.Interface       = interface2{}
	_ exports.NumberInterface = &number2{}
)

type interface2 struct{}

func (i interface2) Number(rep cm.Rep) exports.NumberInterface {
	return cm.AsRep[*number2](rep)
}

func (i interface2) NewNumber(value int32) exports.Number {
	return exports.NumberResourceNew(number1(value))
}

func (i interface2) NumberMerge(a exports.Number, b exports.Number) exports.Number {
	return i.NewNumber(a.ResourceRep().Value() + b.ResourceRep().Value())
}

func (i interface2) NumberChoose(a exports.NumberInterface, b exports.NumberInterface) exports.Number {
	return i.NewNumber(a.Value() + b.Value())
}

type number2 struct {
	value int32
}

func (n *number2) Value() int32        { return n.value }
func (n *number2) String() string      { return strconv.Itoa(int(n.value)) }
func (n *number2) ResourceDestructor() {}
func (n *number2) ResourceRep() cm.Rep { return 0 /* FIXME */ }
