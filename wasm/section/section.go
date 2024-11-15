package section

// ID represents a WebAssembly [section ID].
//
// section ID: https://webassembly.github.io/spec/core/binary/modules.html#sections
type ID uint8

const (
	IDCustom    ID = 0
	IDType      ID = 1
	IDImport    ID = 2
	IDFunction  ID = 3
	IDTable     ID = 4
	IDMemory    ID = 5
	IDGlobal    ID = 6
	IDExport    ID = 7
	IDStart     ID = 8
	IDElement   ID = 9
	IDCode      ID = 10
	IDData      ID = 11
	IDDataCount ID = 12
)
