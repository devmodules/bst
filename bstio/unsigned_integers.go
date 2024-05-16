package bstio

import (
	"encoding/binary"
	"io"
	"math/bits"

	"github.com/devmodules/bst/bsterr"
)

// ReadUint8 reads an unsigned 8-bit integer from the given reader.
func ReadUint8(r io.Reader, desc bool) (uint8, int, error) {
	bt, err := ReadByte(r)
	if err != nil {
		return 0, 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read uint8 value")
	}
	bt, err = ParseUint8Value(bt, desc)
	if err != nil {
		return 0, 1, err
	}
	return bt, 1, nil
}

// WriteUint8 writes an unsigned 8-bit integer to the given writer.
func WriteUint8(w io.Writer, v uint8, desc bool) (int, error) {
	if bw, ok := w.(io.ByteWriter); ok {
		if err := writeUint8ByteWriter(bw, v, desc); err != nil {
			return 0, err
		}
		return 1, nil
	}

	if err := writeUint8Writer(w, v, desc); err != nil {
		return 0, err
	}
	return 1, nil
}

func writeUint8Writer(w io.Writer, v uint8, desc bool) error {
	if desc {
		v = ^v
	}
	_, err := w.Write([]byte{v})
	if err != nil {
		return bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write uint8 value")
	}
	return nil
}

func writeUint8ByteWriter(w io.ByteWriter, v uint8, desc bool) error {
	if desc {
		v = ^v
	}
	if err := w.WriteByte(v); err != nil {
		return bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write uint8 value")
	}

	return nil
}

// MarshalUint8 encodes an unsigned 8-bit integer to the given writer.
func MarshalUint8(v uint8, desc bool) []byte {
	if desc {
		v = ^v
	}
	return []byte{v}
}

// ParseUint8Value parses an unsigned 8-bit integer from the given byte.
func ParseUint8Value(bt byte, desc bool) (uint8, error) {
	// 1. If the value is encoded in descending order, reverse the bytes.
	if desc {
		bt = ^bt
	}
	return bt, nil
}

// SkipUint8Value skips an unsigned 8-bit integer from the given reader.
func SkipUint8Value(s io.ReadSeeker) (int64, error) {
	_, err := s.Seek(1, io.SeekCurrent)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip uint8 value")
	}
	return 1, nil
}

// WriteUint16 writes an unsigned 16-bit integer to the given writer.
func WriteUint16(w io.Writer, v uint16, desc bool) (int, error) {
	if bw, ok := w.(io.ByteWriter); ok {
		if err := writeUint16ByteWriter(bw, v, desc); err != nil {
			return 0, err
		}
		return 2, nil
	}

	if err := writeUint16Writer(w, v, desc); err != nil {
		return 0, err
	}
	return 2, nil
}

func writeUint16Writer(w io.Writer, v uint16, desc bool) error {
	if desc {
		v = ^v
	}
	_, err := w.Write([]byte{byte(v >> 8), byte(v)})
	if err != nil {
		return bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write uint16 value")
	}
	return nil
}

func writeUint16ByteWriter(w io.ByteWriter, v uint16, desc bool) error {
	if desc {
		v = ^v
	}
	if err := w.WriteByte(byte(v >> 8)); err != nil {
		return bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write uint16 value")
	}
	if err := w.WriteByte(byte(v)); err != nil {
		return bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write uint16 value")
	}
	return nil
}

// WriteUint32 writes an unsigned 32-bit integer to the given writer.
func WriteUint32(w io.Writer, v uint32, desc bool) (int, error) {
	if bw, ok := w.(io.ByteWriter); ok {
		return writeUint32ByteWriter(bw, v, desc)
	}
	return writeUint32Writer(w, v, desc)
}

func writeUint32Writer(w io.Writer, v uint32, desc bool) (int, error) {
	if desc {
		v = ^v
	}
	n, err := w.Write([]byte{
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	})
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write uint32 value")
	}
	return n, nil
}

func writeUint32ByteWriter(bw io.ByteWriter, v uint32, desc bool) (int, error) {
	if desc {
		v = ^v
	}

	var (
		n   int
		err error
	)
	writeByte := func(b byte) {
		if err != nil {
			return
		}
		err = bw.WriteByte(b)
		if err != nil {
			err = bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write uint32 value")
			return
		}
		n++
	}
	writeByte(byte(v >> 24))
	writeByte(byte(v >> 16))
	writeByte(byte(v >> 8))
	writeByte(byte(v))
	return n, err
}

// WriteUint64 writes an unsigned 64-bit integer to the given writer.
func WriteUint64(w io.Writer, v uint64, desc bool) (int, error) {
	if bw, ok := w.(io.ByteWriter); ok {
		return writeUint64ByteWriter(bw, v, desc)
	}
	return writeUint64Writer(w, v, desc)
}

func writeUint64ByteWriter(bw io.ByteWriter, v uint64, desc bool) (int, error) {
	if desc {
		v = ^v
	}

	var (
		n   int
		err error
	)

	writeByte := func(b byte) {
		if err != nil {
			return
		}
		err = bw.WriteByte(b)
		if err != nil {
			err = bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write uint64 value")
			return
		}
		n++
	}
	writeByte(byte(v >> 56))
	writeByte(byte(v >> 48))
	writeByte(byte(v >> 40))
	writeByte(byte(v >> 32))
	writeByte(byte(v >> 24))
	writeByte(byte(v >> 16))
	writeByte(byte(v >> 8))
	writeByte(byte(v))
	return n, err
}

func writeUint64Writer(w io.Writer, v uint64, desc bool) (int, error) {
	if desc {
		v = ^v
	}

	n, err := w.Write([]byte{
		byte(v >> 56),
		byte(v >> 48),
		byte(v >> 40),
		byte(v >> 32),
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	})
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write uint64 value")
	}
	return n, nil
}

// MarshalUint16 marshals unsigned integer 16-bit value to a binary format.
func MarshalUint16(v uint16, desc bool) []byte {
	res := []byte{
		byte(v >> 8),
		byte(v),
	}
	if desc {
		ReverseBytes(res)
	}
	return res
}

// ReadUint16 reads an unsigned 16-bit integer from the reader.
func ReadUint16(r io.Reader, desc bool) (uint16, int, error) {
	if br, ok := r.(io.ByteReader); ok {
		return readUint16ByteReader(br, desc)
	}
	return readUint16Reader(r, desc)
}

func readUint16ByteReader(br io.ByteReader, desc bool) (uint16, int, error) {
	bt, err := br.ReadByte()
	if err != nil {
		return 0, 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read uint16 value")
	}
	b, err := br.ReadByte()
	if err != nil {
		return 0, 1, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read uint16 value")
	}

	v := uint16(bt)<<8 | uint16(b)
	if desc {
		v = ^v
	}
	return v, 2, nil
}

func readUint16Reader(br io.Reader, desc bool) (uint16, int, error) {
	bl := make([]byte, 2)
	n, err := br.Read(bl)
	if err != nil {
		return 0, n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read uint16 value")
	}

	uv := uint16(bl[0])<<8 | uint16(bl[1])
	if desc {
		uv = ^uv
	}
	return uv, n, nil
}

// ParseUint16 parses a binary representation into an unsigned 16-bit integer.
// If desc is true, the value is expected to be in descending order.
func ParseUint16(bl []byte, desc bool) (uint16, error) {
	if len(bl) != 2 {
		return 0, bsterr.Errf(bsterr.CodeDecodingBinaryValue, "invalid uint16 binary length").
			WithDetails(
				bsterr.D("length", len(bl)),
				bsterr.D("expected", 2),
			)
	}
	if desc {
		ReverseBytes(bl)
	}
	return binary.BigEndian.Uint16(bl), nil
}

// SkipUint16 skips a binary representation of the 16-bit unsigned integer.
func SkipUint16(s io.ReadSeeker) (int64, error) {
	_, err := s.Seek(2, io.SeekCurrent)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip uint16 value")
	}
	return 2, nil
}

// MarshalUint32 encodes an unsigned 32-bit integer to a binary format.
// If desc is true, the value is expected to be in descending order.
func MarshalUint32(v uint32, desc bool) []byte {
	res := []byte{
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	}
	if desc {
		ReverseBytes(res)
	}
	return res
}

// ReadUint32 reads binary formatted, unsigned 32-bit integer from the reader.
// If desc is true, the value is expected to be in descending order.
func ReadUint32(r io.Reader, desc bool) (uint32, int, error) {
	if br, ok := r.(io.ByteReader); ok {
		return readUint32ByteReader(br, desc)
	}
	return readUint32Reader(r, desc)
}

func readUint32Reader(br io.Reader, desc bool) (uint32, int, error) {
	bl := make([]byte, 4)
	n, err := br.Read(bl)
	if err != nil {
		return 0, n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read uint32 value")
	}

	uv := uint32(bl[0])<<24 | uint32(bl[1])<<16 | uint32(bl[2])<<8 | uint32(bl[3])
	if desc {
		uv = ^uv
	}
	return uv, n, nil
}

func readUint32ByteReader(br io.ByteReader, desc bool) (uint32, int, error) {
	var (
		n   int
		err error
	)
	next := func() uint32 {
		if err != nil {
			return 0
		}
		b, er := br.ReadByte()
		if er != nil {
			err = bsterr.ErrWrap(er, bsterr.CodeDecodingBinaryValue, "failed to read uint32 value")
			return 0
		}
		n++

		return uint32(b)
	}

	uv := next()<<24 | next()<<16 | next()<<8 | next()

	if desc {
		uv = ^uv
	}

	return uv, n, err
}

// ParseUint32 parses a binary representation into an unsigned 32-bit integer.
// If desc is true, the value is expected to be in descending order.
func ParseUint32(bl []byte, desc bool) (uint32, error) {
	if len(bl) != 4 {
		return 0, bsterr.Errf(bsterr.CodeDecodingBinaryValue, "invalid uint32 binary length").
			WithDetails(
				bsterr.D("length", len(bl)),
				bsterr.D("expected", 4),
			)
	}

	if desc {
		ReverseBytes(bl)
	}

	return binary.BigEndian.Uint32(bl), nil
}

// SkipUint32 skips a binary representation of the 32-bit unsigned integer.
func SkipUint32(s io.ReadSeeker) (int64, error) {
	_, err := s.Seek(4, io.SeekCurrent)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip uint32 value")
	}
	return 4, nil
}

// ReadUint64 reads binary formatted, unsigned 64-bit integer from the reader.
// If desc is true, the value is expected to be in descending order.
func ReadUint64(r io.Reader, desc bool) (uint64, int, error) {
	if br, ok := r.(io.ByteReader); ok {
		return readUint64ByteReader(br, desc)
	}
	return readUint64Reader(r, desc)
}

func readUint64Reader(br io.Reader, desc bool) (uint64, int, error) {
	bl := make([]byte, 8)
	n, err := br.Read(bl)
	if err != nil {
		return 0, n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read uint64 value")
	}

	uv := uint64(bl[0])<<56 | uint64(bl[1])<<48 | uint64(bl[2])<<40 | uint64(bl[3])<<32 |
		uint64(bl[4])<<24 | uint64(bl[5])<<16 | uint64(bl[6])<<8 | uint64(bl[7])

	if desc {
		uv = ^uv
	}
	return uv, n, nil
}

func readUint64ByteReader(br io.ByteReader, desc bool) (uint64, int, error) {
	var (
		n   int
		err error
	)

	next := func() uint64 {
		if err != nil {
			return 0
		}
		b, er := br.ReadByte()
		if er != nil {
			err = bsterr.ErrWrap(er, bsterr.CodeDecodingBinaryValue, "failed to read uint64 value")
			return 0
		}
		n++
		return uint64(b)
	}

	uv := next()<<56 | next()<<48 | next()<<40 | next()<<32 |
		next()<<24 | next()<<16 | next()<<8 | next()

	if desc {
		uv = ^uv
	}

	return uv, n, err
}

// ParseUint64 parses a binary representation into an unsigned 64-bit integer.
// If desc is true, the value is expected to be in descending order.
func ParseUint64(bl []byte, desc bool) (uint64, error) {
	if len(bl) != 8 {
		return 0, bsterr.Errf(bsterr.CodeDecodingBinaryValue, "invalid uint64 binary length").
			WithDetails(
				bsterr.D("length", len(bl)),
				bsterr.D("expected", 8),
			)
	}
	if desc {
		ReverseBytes(bl)
	}
	return binary.BigEndian.Uint64(bl), nil
}

// SkipUint64 skips a binary representation of the 64-bit unsigned integer.
func SkipUint64(s io.ReadSeeker) (int64, error) {
	_, err := s.Seek(8, io.SeekCurrent)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip uint64 value")
	}
	return 8, nil
}

// MarshalUint64 encodes 64-bit unsigned integer into a binary format.
func MarshalUint64(v uint64, desc bool) []byte {
	res := []byte{
		byte(v >> 56),
		byte(v >> 48),
		byte(v >> 40),
		byte(v >> 32),
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	}
	if desc {
		ReverseBytes(res)
	}
	return res
}

// MarshalUint encodes an unsigned integer with varying number of bytes into a binary format.
// If desc is true, the value is expected to be in descending order.
// The binary size depends on the value and:
//   - if the value is 0 - size = 1
//   - if the value is [1,256) - size = 2
//   - if the value is [256,65536) - size = 3
//   - if the value is [65536,16777216) - size = 4
//   - if the value is [16777216,4294967296) - size = 5
//   - if the value is [4294967296,1099511627776) - size = 6
//   - if the value is [1099511627776,281474976710656) - size = 7
//   - if the value is [281474976710656,72057594037927936) - size = 8
//   - if the value is [72057594037927936,18446744073709551616) - size = 9
func MarshalUint(uv uint, desc bool) []byte {
	bytesNo := findUintBytes(uv)

	res := make([]byte, bytesNo+1)
	header := byte(bytesNo)
	if desc {
		header = ^header
	}
	res[0] = header

	var bt byte

	for i := bytesNo; i >= 1; i-- {
		bt = byte(uv >> uint(8*(i-1)))
		if desc {
			bt = ^bt
		}
		res[i] = bt
	}
	return res
}

// MarshalUintValue encodes an unsigned integer with varying number of bytes into a binary format.
// If desc is true, the value is expected to be in descending order.
// The binary size depends on the value and:
//   - if the value is 0 - size = 1
//   - if the value is [1,256) - size = 2
//   - if the value is [256,65536) - size = 3
//   - if the value is [65536,16777216) - size = 4
//   - if the value is [16777216,4294967296) - size = 5
//   - if the value is [4294967296,1099511627776) - size = 6
//   - if the value is [1099511627776,281474976710656) - size = 7
//   - if the value is [281474976710656,72057594037927936) - size = 8
//   - if the value is [72057594037927936,18446744073709551616) - size = 9
func MarshalUintValue(uv uint, size byte, desc bool) []byte {
	res := make([]byte, size)
	var bt byte
	for i := size; i >= 1; i-- {
		bt = byte(uv >> uint(8*(i-1)))
		if desc {
			bt = ^bt
		}
		res[i] = bt
	}
	return res
}

func findUintBytes(v uint) int {
	lz := bits.LeadingZeros64(uint64(v))
	return ((64 - lz) + 7) >> 3
}

// UintBinarySize returns the number of bytes required to encode an unsigned integer.
func UintBinarySize(value uint) int {
	return findUintBytes(value) + 1
}

// UintSizeHeader returns the number of bytes required to encode an unsigned integer with a header.
func UintSizeHeader(value uint, desc bool) byte {
	header := byte(findUintBytes(value))
	if desc {
		header = ^header
	}
	return header
}

// ReadUint encodes an unsigned integer with varying number of bytes from the reader.
// If desc is true, the value is expected to be in descending order.
func ReadUint(r io.Reader, desc bool) (uint, int, error) {
	// 1. Read the header byte.
	size, err := ReadByte(r)
	if err != nil {
		return 0, 0, err
	}
	n := 1

	// 2. If the value is encoded in descending order, ReverseBytes the bytes.
	if desc {
		size = ^size
	}

	if size > 8 {
		return 0, n, bsterr.Errf(bsterr.CodeDecodingBinaryValue, "invalid uint binary size").
			WithDetails(
				bsterr.D("size", size),
				bsterr.D("expectedMax", "8"),
			)
	}

	readByteFn := func() uint {
		if err != nil {
			return 0
		}
		b, er := ReadByte(r)
		if er != nil {
			err = er
			return 0
		}
		if desc {
			b = ^b
		}
		n++
		return uint(b)
	}

	var res uint
	for i := size; i >= 1; i-- {
		res |= readByteFn() << uint((i-1)*8)
	}
	if err != nil {
		return 0, n, err
	}
	return res, n, nil
}

// DecodeUintBinarySize reads an uint binary size header from the reader.
func DecodeUintBinarySize(br io.ReadSeeker, desc bool) (int, error) {
	// 1. Read the header byte.
	fs, err := ReadByte(br)
	if err != nil {
		return 0, bsterr.Errf(bsterr.CodeDecodingBinaryValue, "reading byte header failed").
			WithDetail("err", err.Error())
	}

	if desc {
		fs = ^fs
	}

	if fs > 8 {
		return 0, bsterr.Errf(bsterr.CodeDecodingBinaryValue, "invalid uint binary size").
			WithDetails(
				bsterr.D("size", fs),
				bsterr.D("expectedMax", "8"),
			)
	}
	return int(fs), nil
}

// SkipUint skips a binary representation of the varying size unsigned integer.
// If desc is true, the value is expected to be in descending order.
func SkipUint(s io.ReadSeeker, desc bool) (int64, error) {
	// 1. Read the header byte - 1 byte.
	size, err := DecodeUintBinarySize(s, desc)
	if err != nil {
		return 0, err
	}
	bytesSkipped := int64(1)
	if size == 0 {
		return bytesSkipped, nil
	}
	if size > 8 {
		return 0, bsterr.Errf(bsterr.CodeDecodingBinaryValue, "invalid uint binary size").
			WithDetails(
				bsterr.D("size", size),
				bsterr.D("expectedMax", "8"),
			)
	}

	_, err = s.Seek(int64(size), io.SeekCurrent)
	if err != nil {
		return bytesSkipped, bsterr.Err(bsterr.CodeDecodingBinaryValue, "failed to skip uint varying size value")
	}
	bytesSkipped += int64(size)
	return bytesSkipped, nil
}

// WriteUint writes an unsigned integer to the given writer.
// If desc is true, the value is expected to be in descending order.
func WriteUint(w io.Writer, uv uint, desc bool) (int, error) {
	if bw, ok := w.(io.ByteWriter); ok {
		return WriteUintByteWriter(bw, uv, desc)
	}
	n, err := w.Write(MarshalUint(uv, desc))
	if err != nil {
		return n, bsterr.Err(bsterr.CodeEncodingBinaryValue, "failed to write uint value")
	}
	return n, nil
}

// WriteUintByteWriter writes an unsigned integer to the given byte writer.
// If desc is true, the value is expected to be in descending order.
// The number of bytes written is returned.
func WriteUintByteWriter(bw io.ByteWriter, uv uint, desc bool) (int, error) {
	// 1. Determine the number of bytes used to encode the value.
	size := findUintBytes(uv)

	// 2. Check if the size byte needs to be reversed due to descending order.
	sizeByte := byte(size)
	if desc {
		sizeByte = ^sizeByte
	}

	// 3. Write the size header.
	if err := bw.WriteByte(sizeByte); err != nil {
		return 1, bsterr.Err(bsterr.CodeEncodingBinaryValue, "failed to write uint value")
	}

	// 4. Get slice by slice a byte of the value and write it to the writer.
	var bt byte
	for i := size; i >= 1; i-- {
		bt = byte(uv >> uint(8*(i-1)))
		if desc {
			bt = ^bt
		}
		if err := bw.WriteByte(bt); err != nil {
			return i + 1, bsterr.Err(bsterr.CodeEncodingBinaryValue, "failed to write uint value")
		}
	}
	return size + 1, nil
}

// ReadUintValue reads an unsigned integer value from the reader, where the size header is provided.
func ReadUintValue(r io.Reader, size byte, desc bool) (uint, int, error) {
	if size > 8 {
		return 0, 0, bsterr.Errf(bsterr.CodeDecodingBinaryValue, "invalid uint binary size").
			WithDetails(
				bsterr.D("size", size),
				bsterr.D("expectedMax", "8"),
			)
	}

	var (
		err error
		n   int
	)

	readByteFn := func() uint {
		if err != nil {
			return 0
		}
		b, er := ReadByte(r)
		if er != nil {
			err = er
			return 0
		}
		if desc {
			b = ^b
		}
		n++
		return uint(b)
	}

	var res uint
	for i := size; i >= 1; i-- {
		res |= readByteFn() << uint((i-1)*8)
	}
	if err != nil {
		return 0, n, err
	}
	return res, n, nil
}

// WriteUintValue writes an unsigned integer value (without the size header) to the given writer.
func WriteUintValue(w io.Writer, v uint, size byte, desc bool) (int, error) {
	if bw, ok := w.(io.ByteWriter); ok {
		return writeUintValueByteWriter(bw, v, size, desc)
	}
	n, err := w.Write(MarshalUintValue(v, size, desc))
	if err != nil {
		return n, bsterr.Err(bsterr.CodeEncodingBinaryValue, "failed to write uint value")
	}
	return n, nil
}

func writeUintValueByteWriter(bw io.ByteWriter, uv uint, size byte, desc bool) (int, error) {
	// 4. Get slice by slice a byte of the value and write it to the writer.
	var bt byte
	for i := int(size); i >= 1; i-- {
		bt = byte(uv >> uint(8*(i-1)))
		if desc {
			bt = ^bt
		}
		if err := bw.WriteByte(bt); err != nil {
			return i, bsterr.Err(bsterr.CodeEncodingBinaryValue, "failed to write uint value")
		}
	}
	return int(size), nil
}
