package section

// ID represents a WebAssembly [section ID].
//
// [section ID]: https://webassembly.github.io/spec/core/binary/modules.html#sections
type ID uint8

const (
	Custom    ID = 0
	Type      ID = 1
	Import    ID = 2
	Function  ID = 3
	Table     ID = 4
	Memory    ID = 5
	Global    ID = 6
	Export    ID = 7
	Start     ID = 8
	Element   ID = 9
	Code      ID = 10
	Data      ID = 11
	DataCount ID = 12
)
