package bstio

import (
	"io"

	"github.com/devmodules/bst/bsterr"
)

// ReadBool reads a boolean binary value from the given reader.
// The desc flag indicates if the value is encoded in descending order.
func ReadBool(r io.Reader, desc bool) (bool, int, error) {
	b, err := ReadByte(r)
	if err != nil {
		return false, 0, bsterr.Err(bsterr.CodeReadingFailed, "failed to read bool value")
	}
	bv, err := ParseBool(b, desc)
	return bv, 1, err
}

// ParseBool parses a boolean binary value from the given byte.
// The desc flag indicates if the value is encoded in descending order.
func ParseBool(b byte, desc bool) (bool, error) {
	// 1. If the value is encoded in descending order, ReverseBytes the bytes.
	var v bool
	switch {
	case (!desc && b == BoolTrue) || (desc && b == BoolTrueDesc):
		v = true
	case (!desc && b == BoolFalse) || (desc && b == BoolFalseDesc):
		v = false
	default:
		return false, bsterr.Err(bsterr.CodeDecodingBinaryValue, "invalid bool binary value").
			WithDetails(
				bsterr.D("value", b),
				bsterr.D("encoding", desc),
			)
	}
	return v, nil
}

// SkipBool skips the boolean binary value from the given reader.
func SkipBool(s io.ReadSeeker) (int64, error) {
	n, err := s.Seek(1, io.SeekCurrent)
	if err != nil {
		return n, bsterr.Err(bsterr.CodeDecodingBinaryValue, "failed to skip bool value")
	}
	return 1, nil
}

// Constant boolean values used in the binary encoding.
const (
	BoolTrue      = byte(0x1)
	BoolFalse     = byte(0x0)
	BoolTrueDesc  = BoolFalse
	BoolFalseDesc = BoolTrue
)
