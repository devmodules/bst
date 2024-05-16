package bstio

import (
	"io"

	"github.com/devmodules/bst/bsterr"
)

// WriteOneOfIndex writes the index value to the writer in the specified number of bytes.
// If the number of bytes is not defined it writes the index as size variable Uint.
// If the descending flag is set the index is written in descending order.
// Returns the number of bytes written and any error.
func WriteOneOfIndex(w io.Writer, index uint, indexBytes uint8, descending bool) (int, error) {
	var (
		n   int
		err error
	)
	switch indexBytes {
	case BinarySizeZero:
		n, err = WriteUint(w, index, descending)
	case BinarySizeUint8:
		n, err = WriteUint8(w, uint8(index), descending)
	case BinarySizeUint16:
		n, err = WriteUint16(w, uint16(index), descending)
	case BinarySizeUint32:
		n, err = WriteUint32(w, uint32(index), descending)
	case BinarySizeUint64:
		n, err = WriteUint64(w, uint64(index), descending)
	default:
		return 0, bsterr.Err(bsterr.CodeInvalidIntegerBytesValue, "invalid oneOfType buffIndex bytes number").
			WithDetails(
				bsterr.D("indexBytes", indexBytes),
			)
	}
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write oneOfType buffIndex")
	}
	return n, nil
}

// ReadOneOfIndex reads the oneOf index value from the reader in the specified number of bytes.
// If the number of bytes is not defined it reads the index as size variable Uint.
// If the descending flag is set the index is read in descending order.
// Returns the index and number of read bytes.
func ReadOneOfIndex(r io.Reader, indexBytes uint8, desc bool) (uint, int, error) {
	// 1. Read the buffIndex value.
	var (
		bytesRead int
		err       error
		idx       uint
	)
	switch indexBytes {
	case BinarySizeZero:
		// 0 is a special value, meaning that the value is not set.
		idx, bytesRead, err = ReadUint(r, desc)
		if err != nil {
			return 0, bytesRead, err
		}
	case BinarySizeUint8:
		var b uint8
		b, bytesRead, err = ReadUint8(r, desc)
		if err != nil {
			return 0, bytesRead, err
		}
		idx = uint(b)
	case BinarySizeUint16:
		var b uint16
		b, bytesRead, err = ReadUint16(r, desc)
		if err != nil {
			return 0, bytesRead, err
		}
		idx = uint(b)
	case BinarySizeUint32:
		var b uint32
		b, bytesRead, err = ReadUint32(r, desc)
		if err != nil {
			return 0, bytesRead, err
		}
		idx = uint(b)
	case BinarySizeUint64:
		var b uint64
		b, bytesRead, err = ReadUint64(r, desc)
		if err != nil {
			return 0, bytesRead, err
		}
		idx = uint(b)
	default:
		return 0, 0, bsterr.Err(bsterr.CodeDecodingBinaryValue, "invalid oneOf buffIndex size").
			WithDetails(
				bsterr.D("indexBytes", indexBytes),
			)
	}
	return idx, bytesRead, nil
}
