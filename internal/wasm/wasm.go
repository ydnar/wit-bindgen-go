package wasm

import (
	"bufio"
	"io"

	"go.bytecodealliance.org/wasm/section"
	"go.bytecodealliance.org/wasm/uleb128"
)

const (
	Magic    = "\x00asm"
	Version1 = "\x01\x00\x00\x00"
)

// WriteModuleHeader writes a binary [WebAssembly module header], version 1.
//
// WebAssembly module header: https://webassembly.github.io/spec/core/binary/modules.html#binary-module
func WriteModuleHeader(w io.Writer) error {
	_, err := w.Write([]byte(Magic))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(Version1))
	return err
}

// WriteSectionHeader writes a binary [WebAssembly section header].
//
// WebAssembly section header: https://webassembly.github.io/spec/core/binary/modules.html#sections
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
