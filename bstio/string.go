package bstio

import (
	"bytes"
	"io"
	"reflect"
	"unsafe"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/internal/iopool"
)

// WriteString encodes and writes an input string to the writer in the binary representation.
// If the desc flag is set to true, the string is encoded in descending order.
// The string could be encoded either in comparable or non-comparable mode.
// A comparable-mode is slightly slower, and is based on the specific bytes escapes, however guarantees that the
// raw binary representation could be compared to other binary representations.
// In non-comparable mode, the string's length is firstly encoded, and then the string's bytes are written.
// A non-comparable mode is faster and more efficient in memory allocations, but it is not guaranteed to be comparable.
func WriteString(w io.Writer, s string, desc, comparable bool) (int, error) {
	if comparable {
		return WriteStringComparable(w, s, desc)
	}
	return WriteStringNonComparable(w, s, desc)
}

// WriteStringNonComparable encodes and writes an input string to the writer in the binary representation.
// If the desc flag is set to true, the string is encoded in descending order.
// A non-comparable string at first encodes the length of the string and then it's bytes.
func WriteStringNonComparable(w io.Writer, v string, desc bool) (int, error) {
	// 1. Write the length of the string.
	bytesWritten, err := WriteUint(w, uint(len(v)), desc)
	if err != nil {
		return bytesWritten, err
	}

	// 2. If the length is 0, return without any bytes written.
	if v == "" {
		return bytesWritten, nil
	}

	// 3. Treat the input differently for ascending and descending order.
	var bts []byte
	if desc {
		// 3.1. For the descending bytes we need to modify the input string and ReverseBytes its bytes.
		bts = []byte(v)
		ReverseBytes(bts)
	} else {
		// 3.2. Ascending order does not require any modifications, and we can unsafely cast the string to bytes.
		bts = UnsafeStringToBytes(v)
	}

	// 4. Write the string bytes to the input writer.
	n, err := w.Write(bts)
	if err != nil {
		return bytesWritten + n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write string value")
	}
	return bytesWritten + n, nil
}

// UnsafeStringToBytes converts the input string to a byte slice without any memory allocations.
func UnsafeStringToBytes(v string) []byte {
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&v))
	bts := (*[0x7fffffff]byte)(unsafe.Pointer(hdr.Data))[:len(v):len(v)]
	return bts
}

// WriteStringComparable encodes and writes an input string to the writer in the binary representation.
// If the desc flag is set to true, the string is encoded in descending order.
// A comparable-mode is slightly slower, and is based on the specific bytes escapes, however guarantees that the
// raw binary representation could be compared to other binary representations.
func WriteStringComparable(w io.Writer, s string, desc bool) (int, error) {
	// 1. Convert the string to a byte slice.
	if s == "" {
		return WriteEmptyComparableBytes(w, desc)
	}

	// 2. Convert the string to a byte slice.
	data := UnsafeStringToBytes(s)

	var (
		b        []byte
		modified bool
	)
	temp := data

	// 3. Iterate over the byte slice and check if there is anything to escape.
	for {
		i := bytes.IndexByte(temp, 0x00)
		if i == -1 {
			break
		}

		b = append(b, temp[:i]...)
		b = append(b, 0x00, 0xff)
		temp = temp[i+1:]
		modified = true
	}

	// 4. If the value was not modified, and the value is stored in ascending order
	//    we can directly write the value to the writer.
	if !modified && !desc {
		// 4.1 Write the value.
		n, err := w.Write(data)
		if err != nil {
			return 0, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write string value")
		}

		// 4.2. Write the escape and terminator.
		if err = WriteByte(w, 0x00); err != nil {
			return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write string value")
		}
		if err = WriteByte(w, 0x01); err != nil {
			return n + 1, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write string value")
		}
		return n + 2, nil
	}

	// 5. If the value was not modified (data is still unsafe bytes), we need to convert it to a safe byte slice.
	if !modified && desc {
		temp = []byte(s)
	}

	// 6. Write the first buffer part to the writer.
	if desc {
		// 6.1. Reverse it for descending order.
		ReverseBytes(b)
	}
	n, err := w.Write(b)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write string value")
	}
	bytesWritten := n

	// 7. Write the second buffer part to the writer.
	if desc {
		// 7.1. Reverse it for descending order.
		ReverseBytes(temp)
	}
	n, err = w.Write(temp)
	if err != nil {
		return bytesWritten, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write string value")
	}
	bytesWritten += n

	// 8. Finish up with the escape and terminator.
	if desc {
		if err = WriteByte(w, 0xff); err != nil {
			return bytesWritten, err
		}
		if err = WriteByte(w, 0x00); err != nil {
			return bytesWritten + 1, err
		}
	} else {
		if err = WriteByte(w, 0x00); err != nil {
			return bytesWritten, err
		}
		if err = WriteByte(w, 0xff); err != nil {
			return bytesWritten + 1, err
		}
	}
	return bytesWritten + 2, nil
}

// WriteEmptyComparableBytes writes up empty comparable bytes to the writer.
func WriteEmptyComparableBytes(w io.Writer, desc bool) (int, error) {
	if desc {
		if err := WriteByte(w, 0xff); err != nil {
			return 0, err
		}
		if err := WriteByte(w, 0x00); err != nil {
			return 1, err
		}
	} else {
		if err := WriteByte(w, 0x00); err != nil {
			return 0, err
		}
		if err := WriteByte(w, 0xff); err != nil {
			return 1, err
		}
	}
	return 2, nil
}

// UnsafeBytesToString converts the input byte slice to a string without any addition memory allocations.
// This is a faster way to convert a byte slice to a string, but it is unsafe.
// Any change in the input byte slice will result with a panic.
func UnsafeBytesToString(in []byte) string {
	return *(*string)(unsafe.Pointer(&in))
}

// SkipString skips the binary representation of the string.
func SkipString(s io.ReadSeeker, desc, comparable bool) (int64, error) {
	if !comparable {
		return SkipNonComparableString(s, desc)
	}

	if desc {
		return SkipComparableBytes(s, 64, BytesEscapeDescending)
	}
	return SkipComparableBytes(s, 64, BytesEscapeAscending)
}

// SkipNonComparableString skips the binary representation of the non-comparable string.
func SkipNonComparableString(s io.ReadSeeker, desc bool) (int64, error) {
	length, skipped, err := ReadUint(s, desc)
	if err != nil {
		return int64(skipped), err
	}

	_, err = s.Seek(int64(length), io.SeekCurrent)
	if err != nil {
		return int64(skipped), bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip string value")
	}
	return int64(skipped) + int64(length), nil
}

// StringBinarySize returns the binary size of the string.
// The size is the length of the string plus the escape and terminator bytes.
func StringBinarySize(v string, comparable bool) uint {
	if comparable {
		bin := UnsafeStringToBytes(v)
		var (
			lastIndex int
			escapes   int
		)
		for {
			i := bytes.IndexByte(bin[lastIndex:], 0x00)
			if i == -1 {
				break
			}
			lastIndex = i + 1
			escapes++
		}
		return uint(len(bin) + escapes + 2)
	}
	// 1 byte for the length size + N bytes for the length + N bytes for the string.
	return uint(UintBinarySize(uint(len(v))) + len(v))
}

// EncodeStringNonComparable encodes the string in the binary format and writes it to the writer.
func EncodeStringNonComparable(v string, desc bool) []byte {
	bl := MarshalUint(uint(len(v)), desc)
	bv := make([]byte, len(bl)+len(v))
	copy(bv, bl)
	copy(bv[len(bl):], v)
	if desc {
		ReverseBytes(bv[len(bl):])
	}
	return bv
}

// ReadString read the binary representation of the string from the reader.
// The reader must be positioned at the start of the string.
// If the desc flag is true, the string is expected to be encoded in descending order.
// If the comparable flag is true, the string is expected to be comparable.
func ReadString(r io.Reader, desc, comparable bool) (string, int, error) {
	if comparable {
		return ReadStringComparable(r, desc)
	}
	return ReadStringNonComparable(r, desc)
}

// ReadStringNonComparable reads the binary representation of the non-comparable string from the reader.
// The reader must be positioned at the start of the string.
// If the desc flag is true, the string is expected to be encoded in descending order.
func ReadStringNonComparable(r io.Reader, desc bool) (string, int, error) {
	// 1. Read the length of the string.
	length, n, err := ReadUint(r, desc)
	if err != nil {
		return "", n, err
	}

	if length == 0 {
		return "", n, nil
	}

	// 2. Read the string.
	bl := make([]byte, length)
	var total int
	total, err = r.Read(bl)
	if err != nil {
		return "", n + total, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read string value")
	}

	// 3. If the value is encoded in descending order, ReverseBytes the bytes.
	if desc {
		ReverseBytes(bl)
	}

	// 4. Return the string.
	return UnsafeBytesToString(bl), n + total, nil
}

// ReadStringComparable reads the binary representation of the comparable string from the reader.
// The reader must be positioned at the start of the string.
// If the desc flag is true, the string is expected to be encoded in descending order.
func ReadStringComparable(r io.Reader, desc bool) (string, int, error) {
	escape := BytesEscapeAscending
	if desc {
		escape = BytesEscapeDescending
	}
	if rs, ok := r.(io.ReadSeeker); ok {
		return readStringValueComparableReadSeeker(rs, desc, escape)
	}
	return readStringValueComparableReader(r, desc, escape)
}

func readStringValueComparableReader(r io.Reader, desc bool, escape escapes) (string, int, error) {
	// 1. Initialize the result buffer.
	buf := iopool.GetBuffer(nil)
	defer iopool.ReleaseBuffer(buf)

	var n int
	// 2. Iterate byte by byte over the reader until we reach the escape terminator.
	for {
		// 2.1. Read the next byte.
		b, err := ReadByte(r)
		if err != nil {
			return "", n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read string value")
		}
		n++

		if err = buf.WriteByte(b); err != nil {
			return "", n, err
		}

		// 2.2. If the byte is not the escape, continue.
		if b != escape.escape {
			continue
		}

		// 2.3. Read the next byte.
		b, err = ReadByte(r)
		if err != nil {
			return "", n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "malformed string binary value")
		}
		n++

		// 2.4. Check if the next byte is a terminator byte, so that we can stop iterating.
		if b == escape.escapedTerm {
			break
		}

		// 2.5. If the next byte is not the escape, check consistency.
		if b != escape.escaped00 {
			return "", n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "malformed string binary value")
		}

		// 2.6. Write escaped byte and continue iteration.
		if err = buf.WriteByte(escape.escapedFF); err != nil {
			return "", n, err
		}
	}

	if desc {
		ReverseBytes(buf.Bytes)
	}

	// 3. In this position a string had to be decoded properly.
	//    Return the string and the number of bytes read.
	return UnsafeBytesToString(buf.Bytes), n, nil
}

func readStringValueComparableReadSeeker(rs io.ReadSeeker, desc bool, escape escapes) (string, int, error) {
	bt, n, err := ReadComparableBytesSeeker(rs, desc, 16, escape)
	if err != nil {
		return "", n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read string value")
	}
	if len(bt) == 0 {
		return "", n, nil
	}
	return UnsafeBytesToString(bt), n, nil
}
