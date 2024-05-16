package bstio

import (
	"io"

	"github.com/devmodules/bst/bsterr"
)

// ReadNullableFlag reads the nullable flag from the reader.
// It literally reads a byte from the reader, reverse it (if descending order flag is true)
// and return it.
func ReadNullableFlag(r io.Reader, descending bool) (byte, error) {
	// 1. Read a nullable flag byte.
	nf, err := ReadByte(r)
	if err != nil {
		return 0, bsterr.Err(bsterr.CodeDecodingBinaryValue, "failed to read nullable flag byte").
			WithDetail("error", err.Error())
	}
	if descending {
		nf = ^nf
	}
	return nf, nil
}

// Constant values for the nullable flags.
const (
	NullableIsNull        = byte(0x00)
	NullableIsNotNull     = byte(0x01)
	NullableIsNullDesc    = ^NullableIsNull
	NullableIsNotNullDesc = ^NullableIsNotNull
)
