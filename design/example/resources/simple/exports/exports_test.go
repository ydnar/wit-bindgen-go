package exports_test

import (
	"strconv"
	"testing"

	"github.com/ydnar/wasm-tools-go/cm"
)

func TestExportNumber(t *testing.T) {
	// exports.ExportNumber[number](numberExports{})
	// exports.ExportNumber[numberStruct](numberStructExports{})
}

func TestNumberResourceRep(t *testing.T) {
	// _ = exports.NumberResourceRep[number, number]
	// _ = exports.NumberResourceRep[*numberStruct, numberStruct]
}

// Value representation
type number int32

func (n number) Value() int32        { return int32(n) }
func (n number) String() string      { return strconv.Itoa(int(n)) }
func (n number) ResourceDestructor() {}
func (n number) ResourceRep() cm.Rep { return 0 /* FIXME */ }

type numberExports struct{}

func (exports numberExports) Constructor(value int32) number { return number(value) }
func (exports numberExports) Merge(a number, b number) number {
	return exports.Constructor(a.Value() + b.Value())
}

// Pointer representation
type numberStruct struct {
	value int32
}

func (n *numberStruct) Value() int32        { return n.value }
func (n *numberStruct) String() string      { return strconv.Itoa(int(n.value)) }
func (n *numberStruct) ResourceDestructor() {}
func (n *numberStruct) ResourceRep() cm.Rep { return 0 /* FIXME */ }

type numberStructExports struct{}

func (exports numberStructExports) Constructor(value int32) *numberStruct {
	return &numberStruct{value}
}

func (exports numberStructExports) Merge(a, b *numberStruct) *numberStruct {
	return exports.Constructor(a.Value() + b.Value())
}
