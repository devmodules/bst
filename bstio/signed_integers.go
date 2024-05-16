package bstio

import (
	"encoding/binary"
	"io"

	"github.com/devmodules/bst/bsterr"
)

// WriteInt8 writes an int8 value to the writer.
// If desc is true, the value is encoded in descending order.
// Positive values has the highest bit set to 1, whereas negative values have the highest bit set to 0.
// This ensures comparability of the values on the bytes level.
func WriteInt8(w io.Writer, iv int8, desc bool) (int, error) {
	if bw, ok := w.(io.ByteWriter); ok {
		return writeInt8ByteWriter(bw, iv, desc)
	}
	return writeInt8Writer(w, iv, desc)
}

func writeInt8ByteWriter(w io.ByteWriter, iv int8, desc bool) (int, error) {
	value := uint8(iv)
	if iv < 0 {
		value &= NegativeBit8Mask
	} else {
		value |= PositiveBit8Mask
	}
	if desc {
		value = ^value
	}
	if err := w.WriteByte(value); err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write int8 value")
	}
	return 1, nil
}

func writeInt8Writer(w io.Writer, v int8, desc bool) (int, error) {
	var res byte
	if v < 0 {
		res = uint8(v) & NegativeBit8Mask
	} else {
		res = byte(v) | PositiveBit8Mask
	}
	if desc {
		res = ^res
	}
	var err error
	if bw, ok := w.(io.ByteWriter); ok {
		err = bw.WriteByte(res)
	} else {
		_, err = w.Write([]byte{res})
	}
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write int8 value")
	}
	return 1, nil
}

// ReadInt8 reads an int8 value from the reader.
// If desc is true, the value is encoded in descending order.
// Positive values has the highest bit set to 1, whereas negative values have the highest bit set to 0.
// This ensures comparability of the values on the bytes level.
func ReadInt8(r io.Reader, desc bool) (int8, int, error) {
	bt, err := ReadByte(r)
	if err != nil {
		return 0, 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read int8 value")
	}
	v, err := ParseInt8(bt, desc)
	if err != nil {
		return 0, 1, err
	}
	return v, 1, nil
}

// ParseInt8 parses a byte value into an int8 value.
func ParseInt8(bt byte, desc bool) (int8, error) {
	// 1. If the value is encoded in descending order, ReverseBytes the bytes.
	if desc {
		bt = ^bt
	}

	// 2. Flip the sign bit.
	bt = ParseSignedValueMSB(bt)
	return int8(bt), nil
}

// SkipInt8 skips an int8 value within the reader.
func SkipInt8(s io.ReadSeeker) (int64, error) {
	n, err := s.Seek(1, io.SeekCurrent)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip int8 value")
	}
	return 1, nil
}

// ParseSignedValueMSB parses the most significant bit of the most significant byte of signed integers.
// It literally flips the most significant bit to the opposite value.
func ParseSignedValueMSB(bt byte) byte {
	isPositive := (bt>>7)&0x1 == 1
	if !isPositive {
		bt |= 1 << 7
	} else {
		bt &= ^uint8(1 << 7)
	}
	return bt
}

// WriteInt16 writes an int16 value to the writer.
// If desc is true, the value is encoded in descending order.
// Positive values has the highest bit set to 1, whereas negative values have the highest bit set to 0.
// This ensures comparability of the values on the bytes level.
func WriteInt16(w io.Writer, v int16, desc bool) (int, error) {
	if bw, ok := w.(io.ByteWriter); ok {
		return writeInt16ByteWriter(bw, v, desc)
	}
	return writeInt16Writer(w, v, desc)
}

func writeInt16Writer(w io.Writer, iv int16, desc bool) (int, error) {
	value := uint16(iv)
	fb := byte(value >> 8)
	if iv < 0 {
		fb &= NegativeBit8Mask
	} else {
		fb |= PositiveBit8Mask
	}
	res := []byte{
		fb,
		byte(value),
	}

	if desc {
		ReverseBytes(res)
	}
	n, err := w.Write(res)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write int16 value")
	}
	return n, nil
}

func writeInt16ByteWriter(w io.ByteWriter, iv int16, desc bool) (int, error) {
	var (
		err error
		n   int
	)
	writeByte := func(b byte) {
		if err != nil {
			return
		}
		if desc {
			b = ^b
		}
		err = w.WriteByte(b)
		if err != nil {
			err = bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write int64 value")
			return
		}
		n++
	}

	value := uint16(iv)
	bt := byte(value >> 8)
	if iv < 0 {
		bt &= NegativeBit8Mask
	} else {
		bt |= PositiveBit8Mask
	}

	writeByte(bt)
	writeByte(byte(value))
	return n, err
}

// MarshalInt16 encodes an int16 value into a byte slice.
// If desc is true, the value is encoded in descending order.
// Positive values has the highest bit set to 1, whereas negative values have the highest bit set to 0.
// This ensures comparability of the values on the bytes level.
func MarshalInt16(v int16, desc bool) []byte {
	uv := uint16(v)
	fb := byte(uv >> 8)
	if v < 0 {
		fb &= NegativeBit8Mask
	} else {
		fb |= PositiveBit8Mask
	}
	res := []byte{
		fb,
		byte(uv),
	}
	if desc {
		ReverseBytes(res)
	}
	return res
}

// ReadInt16 reads an int16 value from the reader.
// If desc is true, the value is encoded in descending order.
// Positive values has the highest bit set to 1, whereas negative values have the highest bit set to 0.
// This ensures comparability of the values on the bytes level.
func ReadInt16(r io.Reader, desc bool) (int16, int, error) {
	if br, ok := r.(io.ByteReader); ok {
		return readInt16ByteReader(br, desc)
	}
	return readInt16Reader(r, desc)
}

func readInt16Reader(r io.Reader, desc bool) (int16, int, error) {
	bl := make([]byte, 2)
	n, err := r.Read(bl)
	if err != nil {
		return 0, n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read int16 value")
	}

	if desc {
		bl[0] = ^bl[0]
		bl[1] = ^bl[1]
	}

	bl[0] = ParseSignedValueMSB(bl[0])

	return int16(uint16(bl[0])<<8 | uint16(bl[1])), n, nil
}

func readInt16ByteReader(r io.ByteReader, desc bool) (int16, int, error) {
	var (
		err error
		n   int
	)

	next := func() uint16 {
		if err != nil {
			return 0
		}

		b, er := r.ReadByte()
		if er != nil {
			err = bsterr.ErrWrap(er, bsterr.CodeDecodingBinaryValue, "failed to read int16 value")
			return 0
		}
		n++
		return uint16(b)
	}

	uv := next()<<8 | next()
	if desc {
		uv = ^uv
	}

	isPositive := uv>>15 != 0
	if !isPositive {
		uv |= 1 << 15
	} else {
		uv &= ^uint16(1 << 15)
	}
	return int16(uv), n, err
}

// ParseInt16 parses binary encoded int16 value from the byte slice.
// If desc is true, the value is encoded in descending order.
// Positive values has the highest bit set to 1, whereas negative values have the highest bit set to 0.
// This ensures comparability of the values on the bytes level.
func ParseInt16(bl []byte, desc bool) (int16, error) {
	if len(bl) != 2 {
		return 0, bsterr.Err(bsterr.CodeDecodingBinaryValue, "failed to parse int16 value. not enough bytes to parse")
	}
	// 1. If the value is encoded in descending order, ReverseBytes the bytes.
	if desc {
		ReverseBytes(bl)
	}

	// 2. Flip the most significant bit.
	bl[0] = ParseSignedValueMSB(bl[0])
	return int16(binary.BigEndian.Uint16(bl)), nil
}

// SkipInt16 skips a binary representation of the int16 value.
func SkipInt16(s io.ReadSeeker) (int64, error) {
	_, err := s.Seek(2, io.SeekCurrent)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip int16 value")
	}
	return 2, nil
}

// WriteInt32 writes an int32 value to the writer.
// If desc is true, the value is encoded in descending order.
// Positive values has the highest bit set to 1, whereas negative values have the highest bit set to 0.
// This ensures comparability of the values on the bytes level.
func WriteInt32(w io.Writer, iv int32, desc bool) (int, error) {
	if bw, ok := w.(io.ByteWriter); ok {
		return writeInt32ByteWriter(bw, iv, desc)
	}
	return writeInt32Writer(w, iv, desc)
}

func writeInt32Writer(w io.Writer, iv int32, desc bool) (int, error) {
	value := uint32(iv)
	fb := byte(value >> 24)
	if iv < 0 {
		fb &= NegativeBit8Mask
	} else {
		fb |= PositiveBit8Mask
	}
	res := []byte{
		fb,
		byte(value >> 16),
		byte(value >> 8),
		byte(value),
	}
	if desc {
		ReverseBytes(res)
	}

	n, err := w.Write(res)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write int32 value")
	}
	return n, nil
}

func writeInt32ByteWriter(w io.ByteWriter, iv int32, desc bool) (int, error) {
	var (
		err error
		n   int
	)
	writeByte := func(b byte) {
		if err != nil {
			return
		}
		if desc {
			b = ^b
		}
		err = w.WriteByte(b)
		if err != nil {
			err = bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write int64 value")
			return
		}
		n++
	}
	value := uint32(iv)
	bt := byte(value >> 24)
	if iv < 0 {
		bt &= NegativeBit8Mask
	} else {
		bt |= PositiveBit8Mask
	}

	writeByte(bt)
	writeByte(byte(value >> 16))
	writeByte(byte(value >> 8))
	writeByte(byte(value))
	return n, err
}

// ReadInt32 reads an int32 value from the reader encoded in the binary representation.
// If desc is true, the value is encoded in descending order.
// Positive values has the highest bit set to 1, whereas negative values have the highest bit set to 0.
// This ensures comparability of the values on the bytes level.
func ReadInt32(r io.Reader, desc bool) (int32, int, error) {
	if br, ok := r.(io.ByteReader); ok {
		return readInt32ByteReader(br, desc)
	}
	return readInt32Reader(r, desc)
}

func readInt32ByteReader(br io.ByteReader, desc bool) (int32, int, error) {
	var (
		err error
		n   int
	)
	next := func() uint32 {
		if err != nil {
			return 0
		}
		b, er := br.ReadByte()
		if er != nil {
			err = bsterr.ErrWrap(er, bsterr.CodeReadingFailed, "failed to read int32 value")
			return 0
		}
		n++
		return uint32(b)
	}

	uv := next()<<24 | next()<<16 | next()<<8 | next()
	if desc {
		uv = ^uv
	}

	isPositive := (uv>>31)&0x1 == 1
	if !isPositive {
		uv |= 1 << 31
	} else {
		uv &= ^uint32(1 << 31)
	}
	return int32(uv), n, nil
}

func readInt32Reader(r io.Reader, desc bool) (int32, int, error) {
	bl := make([]byte, 4)
	n, err := r.Read(bl)
	if err != nil {
		return 0, n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read int32 value")
	}

	uv := uint32(bl[0])<<24 | uint32(bl[1])<<16 | uint32(bl[2])<<8 | uint32(bl[3])
	if desc {
		uv = ^uv
	}

	isPositive := (uv>>31)&0x1 == 1
	if !isPositive {
		uv |= 1 << 31
	} else {
		uv &= ^uint32(1 << 31)
	}
	return int32(uv), n, nil
}

// ParseInt32 decodes an int32 value from a binary representation.
// If desc is true, the value is encoded in descending order.
// Positive values has the highest bit set to 1, whereas negative values have the highest bit set to 0.
// This ensures comparability of the values on the bytes level.
func ParseInt32(bl []byte, desc bool) (int32, error) {
	if len(bl) != 4 {
		return 0, bsterr.Err(bsterr.CodeDecodingBinaryValue, "failed to parse int32 value. not enough bytes to parse")
	}

	// 1. If the value is encoded in descending order, ReverseBytes the bytes.
	if desc {
		ReverseBytes(bl)
	}

	// 2. Flip the most significant bit.
	bl[0] = ParseSignedValueMSB(bl[0])
	return int32(binary.BigEndian.Uint32(bl)), nil
}

// SkipInt32 skips a binary representation of the int32 value.
func SkipInt32(s io.ReadSeeker) (int64, error) {
	_, err := s.Seek(4, io.SeekCurrent)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip int32 value")
	}
	return 4, nil
}

// MarshalInt32 encodes an int32 value in the binary representation.
// If desc is true, the value is encoded in descending order.
// Positive values has the highest bit set to 1, whereas negative values have the highest bit set to 0.
//
//	This ensures comparability of the values on the bytes level.
func MarshalInt32(i32 int32, desc bool) []byte {
	value := uint32(i32)
	fb := byte(value >> 24)
	if i32 < 0 {
		fb &= NegativeBit8Mask
	} else {
		fb |= PositiveBit8Mask
	}
	res := []byte{
		fb,
		byte(value >> 16),
		byte(value >> 8),
		byte(value),
	}
	if desc {
		ReverseBytes(res)
	}
	return res
}

// ReadInt64 reads an int64 value from the reader encoded in the binary representation.
// If desc is true, the value is encoded in descending order.
// Positive values has the highest bit set to 1, whereas negative values have the highest bit set to 0.
// This ensures comparability of the values on the bytes level.
func ReadInt64(r io.Reader, desc bool) (int64, int, error) {
	if br, ok := r.(io.ByteReader); ok {
		return readInt64ByteReader(br, desc)
	}
	return readInt64Reader(r, desc)
}

func readInt64Reader(r io.Reader, desc bool) (int64, int, error) {
	bl := make([]byte, 8)
	n, err := r.Read(bl)
	if err != nil {
		return 0, n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read int64 value")
	}

	uv := uint64(bl[0])<<56 | uint64(bl[1])<<48 | uint64(bl[2])<<40 | uint64(bl[3])<<32 |
		uint64(bl[4])<<24 | uint64(bl[5])<<16 | uint64(bl[6])<<8 | uint64(bl[7])

	// 2. If the value is encoded in descending order, ReverseBytes the bytes.
	if desc {
		uv = ^uv
	}

	// 2. Flip the most significant bit of the int64 value.
	isPositive := (uv>>63)&0x1 == 1
	if !isPositive {
		uv |= 1 << 63
	} else {
		uv &= ^uint64(1 << 63)
	}
	return int64(uv), n, nil
}

func readInt64ByteReader(r io.ByteReader, desc bool) (int64, int, error) {
	var (
		n   int
		err error
	)

	rb := func() uint64 {
		if err != nil {
			return 0
		}
		var b byte
		b, err = r.ReadByte()
		if err != nil {
			return 0
		}
		n++
		return uint64(b)
	}

	uv := rb()<<56 | rb()<<48 | rb()<<40 | rb()<<32 | rb()<<24 | rb()<<16 | rb()<<8 | rb()
	if desc {
		uv = ^uv
	}

	// 2. Flip the most significant bit of the int64 value.
	isPositive := (uv>>63)&0x1 == 1
	if !isPositive {
		uv |= 1 << 63
	} else {
		uv &= ^uint64(1 << 63)
	}
	return int64(uv), n, err
}

// WriteInt64 writes an int64 value to the writer encoded in the binary representation.
// If desc is true, the value is encoded in descending order.
// Positive values has the highest bit set to 1, whereas negative values have the highest bit set to 0.
// This ensures comparability of the values on the bytes level.
func WriteInt64(w io.Writer, iv int64, desc bool) (int, error) {
	if bw, ok := w.(io.ByteWriter); ok {
		return writeInt64ByteWriter(bw, iv, desc)
	}
	return writeInt64Writer(w, iv, desc)
}

func writeInt64ByteWriter(w io.ByteWriter, iv int64, desc bool) (int, error) {
	var (
		err error
		n   int
	)
	writeByteFn := func(b byte) {
		if err != nil {
			return
		}
		if desc {
			b = ^b
		}
		err = w.WriteByte(b)
		if err != nil {
			err = bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write int64 value")
			return
		}
		n++
	}
	value := uint64(iv)
	bt := byte(value >> 56)
	if iv < 0 {
		bt &= NegativeBit8Mask
	} else {
		bt |= PositiveBit8Mask
	}

	writeByteFn(bt)
	writeByteFn(byte(value >> 48))
	writeByteFn(byte(value >> 40))
	writeByteFn(byte(value >> 32))
	writeByteFn(byte(value >> 24))
	writeByteFn(byte(value >> 16))
	writeByteFn(byte(value >> 8))
	writeByteFn(byte(value))
	return n, err
}

func writeInt64Writer(w io.Writer, iv int64, desc bool) (int, error) {
	value := uint64(iv)
	fb := byte(value >> 56)
	if iv < 0 {
		fb &= NegativeBit8Mask
	} else {
		fb |= PositiveBit8Mask
	}
	res := []byte{
		fb,
		byte(value >> 48),
		byte(value >> 40),
		byte(value >> 32),
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value),
	}
	if desc {
		ReverseBytes(res)
	}

	n, err := w.Write(res)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write int64 value")
	}
	return n, nil
}

// ParseInt64 parses an int64 value from the string representation.
// If desc is true, the value is encoded in descending order.
// Positive values has the highest bit set to 1, whereas negative values have the highest bit set to 0.
// This ensures comparability of the values on the bytes level.
func ParseInt64(bl []byte, desc bool) (int64, error) {
	// 1. Check the length of the bytes.
	if len(bl) != 8 {
		return 0, bsterr.Err(bsterr.CodeDecodingBinaryValue, "failed to parse int64 value. not enough bytes to parse")
	}

	// 2. If the value is encoded in descending order, ReverseBytes the bytes.
	if desc {
		ReverseBytes(bl)
	}

	// 2. Flip the most significant bit of the int64 value.
	bl[0] = ParseSignedValueMSB(bl[0])
	return int64(binary.BigEndian.Uint64(bl)), nil
}

// SkipInt64 skips a binary representation of the int64 value.
func SkipInt64(s io.ReadSeeker) (int64, error) {
	_, err := s.Seek(8, io.SeekCurrent)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip int64 value")
	}
	return 8, nil
}

// MarshalInt64 encodes the int64 value into the binary representation.
// If desc is true, the value is encoded in descending order.
func MarshalInt64(iv int64, desc bool) []byte {
	value := uint64(iv)
	fb := byte(value >> 56)
	if iv < 0 {
		fb &= NegativeBit8Mask
	} else {
		fb |= PositiveBit8Mask
	}
	res := []byte{
		fb,
		byte(value >> 48),
		byte(value >> 40),
		byte(value >> 32),
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value),
	}
	if desc {
		ReverseBytes(res)
	}
	return res
}

// ReadInt reads a varying length int value from the reader.
// If desc is true, the value is encoded in descending order.
// If comparable is true, the value is encoded in comparable representation.
// In the comparable representation the value is encoded as 64-bit signed integer value.
// In non-comparable representation the value is encoded as 64-bit unsigned integer value,
// which is fast and efficient in terms of the number of bytes required to represent the value.
// Returns the integer value along with the number of bytes read.
func ReadInt(r io.Reader, desc, comparable bool) (int, int, error) {
	var (
		iv, n int
		err   error
	)
	if comparable {
		var v int64
		v, n, err = ReadInt64(r, desc)
		iv = int(v)
	} else {
		var uv uint
		uv, n, err = ReadUint(r, desc)
		iv = int(uv)
	}
	return iv, n, err
}

// SkipInt skips a varying length int value from the reader.
// If desc is true, the value is encoded in descending order.
func SkipInt(s io.ReadSeeker, desc, comparable bool) (int64, error) {
	if !comparable {
		return SkipUint(s, desc)
	}
	_, err := s.Seek(8, io.SeekCurrent)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip int value")
	}

	return 8, nil
}

// WriteInt writes a varying length int value to the writer.
// If desc is true, the value is encoded in descending order.
// If comparable is true, the value is encoded in comparable representation.
// In the comparable representation the value is encoded as 64-bit signed integer value.
// In non-comparable representation the value is encoded as 64-bit unsigned integer value,
// which is fast and efficient in terms of the number of bytes required to represent the value.
func WriteInt(w io.Writer, iv int, desc, comparable bool) (int, error) {
	if comparable {
		return WriteInt64(w, int64(iv), desc)
	}
	return WriteUint(w, uint(iv), desc)
}
