package wasm

import (
	"bytes"

	"go.bytecodealliance.org/internal/wasm/uleb128"
)

// SectionID represents a WebAssembly [section SectionID].
//
// [section SectionID]: https://webassembly.github.io/spec/core/binary/modules.html#sections
type SectionID uint8

const (
	SectionCustom    SectionID = 0
	SectionType      SectionID = 1
	SectionImport    SectionID = 2
	SectionFunction  SectionID = 3
	SectionTable     SectionID = 4
	SectionMemory    SectionID = 5
	SectionGlobal    SectionID = 6
	SectionExport    SectionID = 7
	SectionStart     SectionID = 8
	SectionElement   SectionID = 9
	SectionCode      SectionID = 10
	SectionData      SectionID = 11
	SectionDataCount SectionID = 12
)

// Section represents an abstract [WebAssembly section].
//
// [WebAssembly section]: https://webassembly.github.io/spec/core/binary/modules.html#sections
type Section interface {
	// SectionID returns the section ID of this section.
	SectionID() SectionID

	// SectionContents returns the section contents as a byte slice.
	SectionContents() ([]byte, error)
}

// CustomSection represents a [WebAssembly custom section].
//
// [WebAssembly custom section]: https://webassembly.github.io/spec/core/binary/modules.html#binary-customsec
type CustomSection struct {
	Name     string
	Contents []byte
}

// SectionID implements the [Section] interface.
func (*CustomSection) SectionID() SectionID {
	return SectionCustom
}

// SectionContents implements the [Section] interface.
func (s *CustomSection) SectionContents() ([]byte, error) {
	var buf bytes.Buffer
	_, err := WriteString(&buf, s.Name)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(s.Contents)
	return buf.Bytes(), err
}

type LinkingSection struct{}

// SectionID implements the [Section] interface.
func (*LinkingSection) SectionID() SectionID {
	return SectionCustom
}

// SectionContents implements the [Section] interface.
func (s *LinkingSection) SectionContents() ([]byte, error) {
	var buf bytes.Buffer
	custom := &CustomSection{Name: "linking"}
	contents, err := custom.SectionContents()
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(contents)
	if err != nil {
		return nil, err
	}
	_, err = uleb128.Write(&buf, 2) // linking section version 2
	return buf.Bytes(), err
}
