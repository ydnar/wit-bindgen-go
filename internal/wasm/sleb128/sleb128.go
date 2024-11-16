// Package sleb128 reads and writes signed LEB128 integers.
package sleb128

import (
	"io"
	"unsafe"
)

// Read reads a signed [LEB128] value from r.
// Returns the value, number of bytes read, and/or an error.
//
// [LEB128]: https://en.wikipedia.org/wiki/LEB128
func Read(r io.ByteReader) (v int64, n int, err error) {
	shift := 0
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, n, err
		}
		n++
		v |= int64(b&0x7f) << shift
		shift += 7
		if (b & 0x80) == 0 {
			if shift < int(unsafe.Sizeof(v)<<8) && (b&0x40) != 0 {
				v |= (^0 << shift)
			}
			break
		}
	}
	return v, n, nil
}

// Write writes v in signed [LEB128] format to w.
// Returns the number of bytes written and/or an error.
//
// Adapted from the Go standard library with the following copyright:
// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// [LEB128]: https://en.wikipedia.org/wiki/LEB128
func Write(w io.ByteWriter, v int64) (n int, err error) {
	more := true
	for more {
		c := uint8(v & 0x7f)
		s := uint8(v & 0x40)
		v >>= 7
		more = (v != 0 || s != 0) && (v != -1 || s == 0)
		if more {
			c |= 0x80
		}
		err = w.WriteByte(c)
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
