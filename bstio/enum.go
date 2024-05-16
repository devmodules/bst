package bstio

import (
	"io"

	"github.com/devmodules/bst/bsterr"
)

// MarshalEnumIndex encodes the enum index with selected number of bits.
// The number of bits is determined by the valueBytes parameter.
// The descending parameter determines if the value is encoded in descending order.
// If the number of bytes is not defined it writes the index as size variable Uint.
func MarshalEnumIndex(index int, valueBytes uint8, descending bool) ([]byte, error) {
	var bin []byte
	switch mv := valueBytes; mv {
	case BinarySizeZero:
		bin = MarshalUint(uint(index), descending)
	case BinarySizeUint8:
		bin = MarshalUint8(uint8(index), descending)
	case BinarySizeUint16:
		bin = MarshalUint16(uint16(index), descending)
	case BinarySizeUint32:
		bin = MarshalUint32(uint32(index), descending)
	case BinarySizeUint64:
		bin = MarshalUint64(uint64(index), descending)
	default:
		return nil, bsterr.Err(bsterr.CodeWritingFailed, "unsupported enum value bits").
			WithDetails(
				bsterr.D("valueBits", mv),
			)
	}
	return bin, nil
}

// WriteEnumIndex writes the enum index with selected number of bits.
// The number of bits is determined by the valueBytes parameter.
// The descending parameter determines if the value is encoded in descending order.
// If the number of bytes is not defined it writes the index as size variable Uint.
func WriteEnumIndex(w io.Writer, index int, valueBytes uint8, descending bool) (int, error) {
	var (
		n   int
		err error
	)
	switch mv := valueBytes; mv {
	case BinarySizeZero:
		n, err = WriteUint(w, uint(index), descending)
	case BinarySizeUint8:
		n, err = WriteUint8(w, uint8(index), descending)
	case BinarySizeUint16:
		n, err = WriteUint16(w, uint16(index), descending)
	case BinarySizeUint32:
		n, err = WriteUint32(w, uint32(index), descending)
	case BinarySizeUint64:
		n, err = WriteUint64(w, uint64(index), descending)
	default:
		return 0, bsterr.Err(bsterr.CodeWritingFailed, "unsupported enum value bits").
			WithDetails(
				bsterr.D("valueBits", mv),
			)
	}
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write enum value")
	}
	return n, err
}

// ReadEnumIndex reads the enum index with selected number of bits.
// The number of bits is determined by the valueBytes parameter.
// The descending parameter determines if the value is encoded in descending order.
// If the number of bytes is not defined it reads the index as size variable Uint.
func ReadEnumIndex(r io.Reader, valueBytes uint8, descending bool) (int, int, error) {
	var (
		n     int
		err   error
		index int
	)
	switch mv := valueBytes; mv {
	case BinarySizeZero:
		var v uint
		v, n, err = ReadUint(r, descending)
		index = int(v)
	case BinarySizeUint8:
		var v uint8
		v, n, err = ReadUint8(r, descending)
		index = int(v)
	case BinarySizeUint16:
		var v uint16
		v, n, err = ReadUint16(r, descending)
		index = int(v)
	case BinarySizeUint32:
		var v uint32
		v, n, err = ReadUint32(r, descending)
		index = int(v)
	case BinarySizeUint64:
		var v uint64
		v, n, err = ReadUint64(r, descending)
		index = int(v)
	default:
		return 0, 0, bsterr.Err(bsterr.CodeInvalidIntegerBytesValue, "unsupported enum value bits").
			WithDetails(
				bsterr.D("valueBits", mv),
			)
	}
	if err != nil {
		return 0, n, err
	}
	return index, n, nil
}

// SkipEnumIndex skips the enum index with selected number of bits.
// The number of bits is determined by the valueBytes parameter.
// The descending parameter determines if the value is encoded in descending order.
// If the number of bytes is not defined it skips the index as size variable Uint.
func SkipEnumIndex(rs io.ReadSeeker, valueBytes uint8, desc bool) (int64, error) {
	switch valueBytes {
	case BinarySizeZero:
		return SkipUint(rs, desc)
	case BinarySizeUint8:
		return SkipUint8Value(rs)
	case BinarySizeUint16:
		return SkipUint16(rs)
	case BinarySizeUint32:
		return SkipUint32(rs)
	case BinarySizeUint64:
		return SkipUint64(rs)
	default:
		return 0, bsterr.Err(bsterr.CodeInvalidIntegerBytesValue, "invalid enum size value").
			WithDetail("size", valueBytes)
	}
}
