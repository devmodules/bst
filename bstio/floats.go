package bstio

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/devmodules/bst/bsterr"
)

// MarshalFloat32 returns the binary representation of the float32 value.
// The first bit of the first byte is set to 1 for positive values, whereas for negative it takes a value of 0.
// This ensures comparability of the binary representation on the bytes level.
// The desc flag determines the order of the bytes.
func MarshalFloat32(value float32, desc bool) []byte {
	ui := math.Float32bits(value)
	fb := byte(ui >> 24)
	if value < 0 {
		// Positive values by standard IEEE-754 are encoded with the first bit set to 0.
		// In this encoding the first bit of the first byte is set to 0 for positive values.
		fb &= NegativeBit8Mask
	} else {
		// Negative values by standard IEEE-754 are encoded with the first bit set to 1.
		// In this encoding the first bit of the first byte is set to 1 for negative values.
		fb |= 1 << 7
	}
	res := []byte{
		fb,
		byte(ui >> 16),
		byte(ui >> 8),
		byte(ui),
	}
	if desc {
		ReverseBytes(res)
	}
	return res
}

// WriteFloat32 writes the float32 value to the writer.
// The desc flag determines the order of the bytes.
func WriteFloat32(w io.Writer, v float32, desc bool) (int, error) {
	if bw, ok := w.(io.ByteWriter); ok {
		return writeFloat32ByteWriter(bw, v, desc)
	}

	n, err := w.Write(MarshalFloat32(v, desc))
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write float value")
	}

	return n, nil
}

func writeFloat32ByteWriter(bw io.ByteWriter, v float32, desc bool) (n int, err error) {
	ui := math.Float32bits(v)
	fb := byte(ui >> 24)
	if v < 0 {
		// Positive values by standard IEEE-754 are encoded with the first bit set to 0.
		// In this encoding the first bit of the first byte is set to 0 for positive values.
		fb &= NegativeBit8Mask
	}

	writeByteFn := func(b byte) {
		if err != nil {
			return
		}
		if desc {
			b = ^b
		}
		err = bw.WriteByte(b)
		if err != nil {
			err = bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write float value")
			return
		}
		n++
	}

	writeByteFn(fb)
	writeByteFn(byte(ui >> 16))
	writeByteFn(byte(ui >> 8))
	writeByteFn(byte(ui))

	return n, err
}

// ReadFloat32 reads a float32 value from the reader.
// The desc flag determines the order of the bytes.
// Returns the float32 value and the number of read bytes.
func ReadFloat32(r io.Reader, desc bool) (float32, int, error) {
	if br, ok := r.(io.ByteReader); ok {
		return readFloat32ByteReader(br, desc)
	}
	bl := make([]byte, 4)
	n, err := r.Read(bl)
	if err != nil {
		return 0, n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read float value")
	}

	fv, err := ParseFloat32(bl, desc)
	if err != nil {
		return 0, n, err
	}
	return fv, n, nil
}

func readFloat32ByteReader(br io.ByteReader, desc bool) (float32, int, error) {
	var (
		n   int
		err error
	)
	readByteFn := func(msb bool) uint32 {
		if err != nil {
			return 0
		}
		bt, er := br.ReadByte()
		if er != nil {
			err = er
			return 0
		}
		if desc {
			bt = ^bt
		}
		if msb {
			bt = ParseSignedValueMSB(bt)
		}
		n++
		return uint32(bt)
	}

	u32 := readByteFn(true)<<24 | readByteFn(false)<<16 | readByteFn(false)<<8 | readByteFn(false)
	return math.Float32frombits(u32), n, err
}

// ParseFloat32 parses the binary representation of a float32 value.
func ParseFloat32(bl []byte, desc bool) (float32, error) {
	// 1. If the value is encoded in ascending order flip the first bit of the first byte.
	//    This ensures that the binary representation is comparable on the bytes level.
	if desc {
		ReverseBytes(bl)
	}
	bl[0] = ParseSignedValueMSB(bl[0])
	return math.Float32frombits(binary.BigEndian.Uint32(bl)), nil
}

// SkipFloat32 skips the bytes of a float32 value.
// Return number of bytes skipped.
func SkipFloat32(s io.ReadSeeker) (int64, error) {
	n, err := s.Seek(4, io.SeekCurrent)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip float value")
	}
	return 4, nil
}

// MarshalFloat64 returns the binary representation of the float64 value.
// The first bit of the first byte is set to 1 for positive values, whereas for negative it takes a value of 0.
// This ensures comparability of the binary representation on the bytes level.
// The desc flag determines the order of the bytes.
func MarshalFloat64(v float64, desc bool) []byte {
	ui := math.Float64bits(v)
	fb := byte(ui >> 56)
	if v < 0 {
		fb &= NegativeBit8Mask
	} else {
		fb |= PositiveBit8Mask
	}
	res := []byte{
		fb,
		byte(ui >> 48),
		byte(ui >> 40),
		byte(ui >> 32),
		byte(ui >> 24),
		byte(ui >> 16),
		byte(ui >> 8),
		byte(ui),
	}
	if desc {
		ReverseBytes(res)
	}
	return res
}

// WriteFloat64 writes the float64 value to the writer.
// The desc flag determines the order of the bytes.
func WriteFloat64(w io.Writer, v float64, desc bool) (int, error) {
	if bw, ok := w.(io.ByteWriter); ok {
		return writeFloat64ByteWriter(bw, v, desc)
	}
	return writeFloat64Writer(w, v, desc)
}

func writeFloat64Writer(w io.Writer, v float64, desc bool) (int, error) {
	n, err := w.Write(MarshalFloat64(v, desc))
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write float value")
	}
	return n, nil
}

func writeFloat64ByteWriter(bw io.ByteWriter, v float64, desc bool) (int, error) {
	ui := math.Float64bits(v)
	fb := byte(ui >> 56)
	if v < 0 {
		fb &= NegativeBit8Mask
	} else {
		fb |= PositiveBit8Mask
	}
	var (
		n   int
		err error
	)
	writeByteFn := func(b byte) {
		if err != nil {
			return
		}
		if desc {
			b = ^b
		}
		if err = bw.WriteByte(b); err != nil {
			err = bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write float value")
			return
		}
		n++
	}

	writeByteFn(fb)
	writeByteFn(byte(ui >> 48))
	writeByteFn(byte(ui >> 40))
	writeByteFn(byte(ui >> 32))
	writeByteFn(byte(ui >> 24))
	writeByteFn(byte(ui >> 16))
	writeByteFn(byte(ui >> 8))
	writeByteFn(byte(ui))
	return n, err
}

// ReadFloat64 reads a float64 value from the reader.
// The desc flag determines the order of the bytes.
func ReadFloat64(r io.Reader, desc bool) (float64, int, error) {
	if br, ok := r.(io.ByteReader); ok {
		return readFloat64ByteReader(br, desc)
	}
	bl := make([]byte, 8)
	n, err := r.Read(bl)
	if err != nil {
		return 0, 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read float value")
	}
	fv, err := ParseFloat64(bl, desc)
	if err != nil {
		return 0, 0, err
	}
	return fv, n, nil
}

func readFloat64ByteReader(br io.ByteReader, desc bool) (float64, int, error) {
	var (
		n   int
		err error
	)
	readByteFn := func(msb bool) uint64 {
		if err != nil {
			return 0
		}
		bt, er := br.ReadByte()
		if er != nil {
			err = er
			return 0
		}
		if desc {
			bt = ^bt
		}
		if msb {
			bt = ParseSignedValueMSB(bt)
		}
		n++
		return uint64(bt)
	}

	u64 := readByteFn(true)<<56 | readByteFn(false)<<48 | readByteFn(false)<<40 | readByteFn(false)<<32 |
		readByteFn(false)<<24 | readByteFn(false)<<16 | readByteFn(false)<<8 | readByteFn(false)
	if err != nil {
		return 0, n, err
	}
	return math.Float64frombits(u64), n, nil
}

// ParseFloat64 parses the binary representation of a float64 value.
// The first bit of the first byte is set to 1 for positive values, whereas for negative it takes a value of 0.
// This ensures comparability of the binary representation on the bytes level.
// The desc flag determines the order of the bytes.
func ParseFloat64(bl []byte, desc bool) (float64, error) {
	// 1. If the value is encoded in descending order, ReverseBytes the bytes.
	if desc {
		ReverseBytes(bl)
	}

	// 2. Flip the sign bit.
	bl[0] = ParseSignedValueMSB(bl[0])

	return math.Float64frombits(binary.BigEndian.Uint64(bl)), nil
}

// SkipFloat64 skips a float64 value from the reader.
func SkipFloat64(s io.ReadSeeker) (int64, error) {
	n, err := s.Seek(8, io.SeekCurrent)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip float64 value")
	}
	return 8, nil
}
