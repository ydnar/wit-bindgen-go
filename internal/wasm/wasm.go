package wasm

import (
	"bufio"
	"io"

	"go.bytecodealliance.org/internal/wasm/section"
	"go.bytecodealliance.org/internal/wasm/uleb128"
)

const (
	Magic    = "\x00asm"
	Version1 = "\x01\x00\x00\x00"
)

// Write writes a binary [WebAssembly module] to w.
//
// [WebAssembly module]: https://webassembly.github.io/spec/core/binary/modules.html#binary-module
func Write(w io.Writer, sections []section.Section) error {
	err := WriteModuleHeader(w)
	if err != nil {
		return err
	}
	for _, s := range sections {
		contents, err := s.SectionContents()
		if err != nil {
			return err
		}
		_, err = WriteSectionHeader(w, s.SectionID(), uint64(len(contents)))
		if err != nil {
			return err
		}
		_, err = w.Write(contents)
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteModuleHeader writes a binary [WebAssembly module header] (version 1) to w.
//
// [WebAssembly module header]: https://webassembly.github.io/spec/core/binary/modules.html#binary-module
func WriteModuleHeader(w io.Writer) error {
	_, err := w.Write([]byte(Magic))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(Version1))
	return err
}

// WriteSectionHeader writes a binary [WebAssembly section header] to w.
// It returns the number of bytes written and/or an error.
//
// [WebAssembly section header]: https://webassembly.github.io/spec/core/binary/modules.html#sections
func WriteSectionHeader(w io.Writer, id section.ID, size uint64) (n int, err error) {
	bw := bufio.NewWriter(w)
	err = bw.WriteByte(byte(id))
	if err != nil {
		return 0, err
	}
	n, err = uleb128.Write(bw, size)
	if err != nil {
		return n + 1, err
	}
	return n + 1, bw.Flush()
}
