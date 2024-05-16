package bstio

import (
	"bytes"
	"errors"
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/internal/iopool"
)

type escapes struct {
	escape      byte
	escapedTerm byte
	escaped00   byte
	escapedFF   byte
}

// Defined escapes used in the binary encoding for specific encodings types.
const (
	BytesEscape = byte(0x00)
	ArrayEscape = byte(0x02)
	MapEscape   = byte(0x03)
)

// Binary escapes for comparable types.
var (
	BytesEscapeAscending  = escapes{BytesEscape, 0x01, 0xFF, 0x00}
	BytesEscapeDescending = escapes{^BytesEscape, 0xFE, 0x00, 0xFF}
	ArrayEscapeAscending  = escapes{ArrayEscape, 0x01, 0xFF, 0x00}
	ArrayEscapeDescending = escapes{^ArrayEscape, 0xFE, 0x00, 0xFF}
	MapEscapeAscending    = escapes{MapEscape, 0x01, 0xFF, 0x00}
	MapEscapeDescending   = escapes{^MapEscape, 0xFE, 0x00, 0xFF}
)

// ReadBytes reads a slice of bytes encoded in the binary format.
// If the fixed size is different from 0, the binary data has a fixed number of bytes.
// The desc flag indicates if the bytes are encoded in descending order.
// The comparable flag indicates if the bytes are comparable - no length defined for varying byte slice.
// The comparable binary is escaped with the BytesEscape byte and uses BytesEscapeAscending or BytesEscapeDescending
// depending on the desc flag.
func ReadBytes(r io.Reader, fixedSize int, desc, comparable bool) ([]byte, int, error) {
	// 1. For fixed size bytes, the amount of bytes to read is directly determined by fixed size.
	//    No matter if the value is comparable or not it is always stored in the same way.
	if fixedSize > 0 {
		return ReadFixedSizeBytes(r, fixedSize, desc)
	}

	// 2. For varying size bytes, in non-comparable format we need to read the length
	//    and then defined amount of bytes to read.
	if !comparable {
		return ReadBytesNonComparable(r, fixedSize, desc)
	}

	// 3. For varying size bytes, in comparable format we need to provide value escapes and stop on terminator.
	escape := BytesEscapeAscending
	if desc {
		escape = BytesEscapeDescending
	}
	if rs, ok := r.(io.ReadSeeker); ok {
		// 3.1. If the reader is a read seeker, it could be much faster to decode the value.
		return ReadComparableBytesSeeker(rs, desc, 64, escape)
	}
	return ReadComparableBytesReader(r, desc, escape)
}

// ReadComparableBytesReader reads binary data from the reader and
func ReadComparableBytesReader(r io.Reader, desc bool, escape escapes) ([]byte, int, error) {
	// 1. Obtain shared buffer.
	buf := iopool.GetBuffer(nil)
	defer iopool.ReleaseBuffer(buf)

	var bytesRead int
	// 2. Iterate byte by byte over the reader until we reach the escape terminator.
	for {
		// 2.1. Read the next byte.
		b, err := ReadByte(r)
		if err != nil {
			return nil, bytesRead, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read string value")
		}
		bytesRead++

		if err = buf.WriteByte(b); err != nil {
			return nil, 0, err
		}

		// 2.2. If the byte is not the escape, continue.
		if b != escape.escape {
			continue
		}

		// 2.3. Read the next byte.
		b, err = ReadByte(r)
		if err != nil {
			return nil, bytesRead, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "malformed string binary value")
		}
		bytesRead++

		// 2.4. Check if the next byte is a terminator byte, so that we can stop iterating.
		if b == escape.escapedTerm {
			break
		}

		// 2.5. If the next byte is not the escape, check consistency.
		if b != escape.escaped00 {
			return nil, bytesRead, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "malformed string binary value")
		}

		// 2.6. Write escaped byte and continue iteration.
		if err = buf.WriteByte(escape.escapedFF); err != nil {
			return nil, bytesRead, err
		}
	}

	if desc {
		ReverseBytes(buf.Bytes)
	}
	return buf.Bytes, len(buf.Bytes), nil
}

// ReadFixedSizeBytes reads a fixed size slice of bytes encoded in the binary format.
// The desc flag indicates if the bytes are encoded in descending order.
func ReadFixedSizeBytes(r io.Reader, fixedSize int, desc bool) ([]byte, int, error) {
	// 1. Allocate bytes data.
	bl := make([]byte, fixedSize)

	// 2. Read the content from the reader.
	n, err := r.Read(bl)
	if err != nil {
		return nil, 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read fixed size bytes value")
	}

	// 3. For descending order, ReverseBytes the bytes.
	if desc {
		ReverseBytes(bl)
	}
	return bl, n, nil
}

// ReadBytesNonComparable reads a slice of bytes encoded in the binary format.
// If the fixed size is different from 0, the binary data has a fixed number of bytes.
// The desc flag indicates if the bytes are encoded in descending order.
func ReadBytesNonComparable(r io.Reader, fixedSize int, desc bool) ([]byte, int, error) {
	// 1. Read the length of the byte slice.
	var total int64
	length := fixedSize
	if length == 0 {
		luv, n, err := ReadUint(r, desc)
		if err != nil {
			return []byte{}, int(n), err
		}
		total += int64(n)
		length = int(luv)
	}
	if length == 0 {
		return []byte{}, int(total), nil
	}

	// 2. Read the byte slice.
	bl := make([]byte, length)
	n, err := r.Read(bl)
	if err != nil {
		return nil, int(total), bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "malformed bytes value binary input")
	}
	n += int(total)

	// 3. If the value is encoded in descending order, ReverseBytes the bytes.
	if desc {
		ReverseBytes(bl)
	}

	// 4. Return the byte slice.
	return bl, n, nil
}

// ReadComparableBytesSeeker reads binary data from the seeker and returns the decoded value.
// The desc flag indicates if the bytes are encoded in descending order.
// The minSize is the minimum size of the value.
// Escapes are used to escape the value.
func ReadComparableBytesSeeker(rs io.ReadSeeker, desc bool, minSize int, escape escapes) ([]byte, int, error) {
	// 1. Save current position of the read seeker so that we may know where we need to stop.
	curPos, err := rs.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "seeking through read seeker failed")
	}

	// 2. Prepare a buffer to read the string into.
	var (
		n, escapePosition int64
		nn                int
	)
	// 3. Start with the mi	nimum size allocation slice.
	nMax := int64(minSize)
	buf := make([]byte, nMax)

	var isEOF, foundTerminator bool
	var r []byte

	// 4. Fill the buffer until we reach the escape bytes.
	for !isEOF {
		// 4.1. If the buffer is full, double the size.
		if n == nMax {
			nMax *= 2
			bNew := make([]byte, nMax)
			copy(bNew, buf)
			buf = bNew
		}

		// 4.2. Read the next bytes in the buffer.
		nn, err = rs.Read(buf[n:])
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, int(n), bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "malformed binary value")
			}
			// 4.2.1. If we reached the end of the file, mark it so that no more bytes are read.
			isEOF = true
		}

		eIdx := int(n)
		// 5. Search for the escape byte in current buffer.
		for eIdx <= int(n)+nn {
			idx := bytes.IndexByte(buf[eIdx:], escape.escape)
			if idx == -1 {
				break
			}
			idx += eIdx

			// 6. If the escape character was found in the last byte of current index, we need to read more bytes to the buffer.
			if idx == len(buf)-1 {
				break
			}

			// 7. Check if the next byte in the buffer is the escaped term.
			if buf[idx+1] == escape.escapedTerm {
				if r == nil {
					r = buf[:idx]
				} else {
					r = append(r, buf[eIdx:idx]...)
				}
				foundTerminator = true
				break
			}

			// 8. If the next byte in the buffer is not the escaped term, the next byte should be a 0x00 byte.
			if buf[idx+1] != escape.escaped00 {
				return nil, int(n), bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "malformed bytes value")
			}

			// 9. If the escape00 was found, replace it with the escapeFF byte.
			r = append(r, buf[eIdx:idx]...)
			r = append(r, escape.escapedFF)
			eIdx = idx + 2
		}
		n += int64(nn)
		escapePosition += int64(nn)

		if foundTerminator {
			break
		}
	}

	// 10. Check if the escape term was found.
	if !foundTerminator {
		return nil, int(n), bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "malformed bytes value")
	}

	// 11. Set the position of the read seeker to the position of the escape term.
	nextPos := curPos + escapePosition + 1
	if _, err = rs.Seek(nextPos, io.SeekStart); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, int(n), bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read string value")
		}
	}

	// 12. If the value is encoded in descending order, ReverseBytes the bytes.
	if desc {
		ReverseBytes(r)
	}
	return r, int(n), nil
}

// BytesBinarySize returns the size of the bytes in binary format.
func BytesBinarySize(fixedSize int, bin []byte, desc, comparable bool) uint {
	if fixedSize > 0 {
		return uint(fixedSize)
	}

	if comparable {
		if desc {
			return BytesComparableBinarySize(bin, BytesEscapeDescending)
		}
		return BytesComparableBinarySize(bin, BytesEscapeAscending)
	}
	// The length header + the length of the byte slice.
	return uint(UintBinarySize(uint(len(bin))) + len(bin))
}

// BytesComparableBinarySize returns the size of the bytes in binary format.
// The escape needs to be provided for matching case i.e.:
// - If the bytes are encoded in ascending order, the escape is BytesEscapeAscending.
func BytesComparableBinarySize(bin []byte, es escapes) uint {
	var lastIndex, escapeCount int
	for {
		i := bytes.IndexByte(bin[lastIndex:], es.escape)
		if i == -1 {
			break
		}
		lastIndex = i + 1
		escapeCount++
	}
	return uint(len(bin) + escapeCount + 2)
}

// WriteBytes encodes and writes input bytes in a binary format to the writer.
// If the fixed size is provided, the bytes are encoded in fixed size.
// If the desc flag is set, the bytes are encoded in descending order.
// Comparable flag is used to determine if the bytes are encoded in comparable mode.
// NOTE: Input data could be malformed during encoding.
func WriteBytes(w io.Writer, fixedSize int, v []byte, desc, comparable bool) (int, error) {
	if fixedSize > 0 || !comparable {
		return writeBytesNonComparable(w, fixedSize, v, desc)
	}
	return writeBytesInternalComparable(w, v, BytesEscape, desc)
}

func writeBytesInternalComparable(w io.Writer, data []byte, eb byte, desc bool) (int, error) {
	// 1. Check if the value is empty.
	if len(data) == 0 {
		return WriteEmptyComparableBytes(w, desc)
	}

	var b []byte

	// 2. Iterate over the byte slice and check if there is anything to escape.
	for {
		i := bytes.IndexByte(data, eb)
		if i == -1 {
			break
		}

		b = append(b, data[:i]...)
		b = append(b, eb, 0xff)
		data = data[i+1:]
	}

	data = append(b, data...)
	data = append(data, eb, 0x01)

	// 4. If the value is encoded in descending order, ReverseBytes the bytes.
	if desc {
		ReverseBytes(data)
	}

	// 5. Write the binary data.
	n, err := w.Write(data)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write bytes value")
	}

	return n, nil
}

// WriteBufferedBytesInternalComparable writes the bytes in a binary format to the input writer.
// The bytes are encoded in comparable mode, taken out of the shared buffer.
func WriteBufferedBytesInternalComparable(w io.Writer, sb *iopool.SharedBuffer, eb byte, desc bool) (int, error) {
	// 1. Check if the value is empty.
	if len(sb.Bytes) == 0 {
		return WriteEmptyComparableBytes(w, desc)
	}

	var b []byte
	data := sb.Bytes

	// 2. Iterate over the byte slice and check if there is anything to escape.
	for {
		i := bytes.IndexByte(data, eb)
		if i == -1 {
			break
		}

		b = append(b, data[:i]...)
		b = append(b, eb, 0xff)
		data = data[i+1:]
	}

	data = append(b, data...)
	data = append(data, eb, 0x01)

	// 4. If the value is encoded in descending order, ReverseBytes the bytes.
	if desc {
		ReverseBytes(data)
	}

	// 5. Write the binary data.
	n, err := w.Write(data)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write bytes value")
	}

	return n, nil
}

func writeBytesNonComparable(w io.Writer, fixedSize int, v []byte, desc bool) (int, error) {
	var bytesWritten int

	// 1. Non-fixed size bytes require to store the length of the data.
	if fixedSize == 0 {
		n, err := WriteUint(w, uint(len(v)), desc)
		if err != nil {
			return 0, err
		}
		bytesWritten += n
	}

	// 2. Write the binary data.
	if v != nil {
		// 3. For descending order, ReverseBytes the bytes.
		if desc {
			cp := make([]byte, len(v))
			copy(cp, v)
			v = cp
			ReverseBytes(v)
		}

		// 4. Write the binary data.
		n, err := w.Write(v)
		if err != nil {
			return bytesWritten + n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write bytes value")
		}
		bytesWritten += n
	}
	return bytesWritten, nil
}

// SkipBytes skips the binary encoded bytes from the input read seeker.
// If the fixed size is provided, the bytes are encoded in fixed size.
// If the desc flag is set, the bytes are encoded in descending order.
// Comparable flag is used to determine if the bytes are encoded in comparable mode.
func SkipBytes(rs io.ReadSeeker, fixedSize int, descending, comparable bool) (int64, error) {
	// 1. For fixed size bytes, the amount of bytes to skip is the fixed size.
	if fixedSize > 0 {
		_, err := rs.Seek(int64(fixedSize), io.SeekCurrent)
		if err != nil {
			return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip bytes value")
		}
		return int64(fixedSize), nil
	}

	// 2. The values encoded in comparable format are slightly more difficult to skip.
	if comparable {
		escape := BytesEscapeAscending
		if descending {
			escape = BytesEscapeDescending
		}
		return SkipComparableBytes(rs, 64, escape)
	}

	// 3. Read the length of the value.
	length, skipped, err := ReadUint(rs, descending)
	if err != nil {
		return int64(skipped), err
	}

	// 4. Skip the number of bytes specified in the length.
	_, err = rs.Seek(int64(length), io.SeekCurrent)
	if err != nil {
		return int64(skipped), bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip bytes value")
	}
	return int64(skipped) + int64(length), nil
}

// SkipComparableBytes skips the binary encoded bytes from the input read seeker in comparable mode.
// The min size determines a size of a buffer to read the data, and the escape byte is used to determine if the data is escaped.
func SkipComparableBytes(rs io.ReadSeeker, minSize int, escape escapes) (int64, error) {
	// 1. Save current position of the read seeker so that we may know where we need to stop.
	curPos, err := rs.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "seeking through read seeker failed")
	}

	// 2. Prepare a buffer to read the string into.
	var (
		n, escapePosition int64
		nn                int
	)
	// 3. Start with the minimum size allocation slice.
	nMax := int64(minSize)
	buf := make([]byte, nMax)
	var isEOF, foundTerminator bool

	// 4. Fill the buffer until we reach the escape bytes.
	for !isEOF {
		// 4.1. If the buffer is full, double the size.
		if n == nMax {
			nMax *= 2
			bNew := make([]byte, nMax)
			copy(bNew, buf)
			buf = bNew
		}

		// 4.2. Read the next bytes in the buffer.
		nn, err = rs.Read(buf[n:])
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "malformed binary value")
			}
			// 4.2.1. If we reached the end of the file, mark it so that no more bytes are read.
			isEOF = true
		}

		eIdx := int(n)
		// 5. Search for the escape byte in current buffer.
		for eIdx <= int(n)+nn {
			idx := bytes.IndexByte(buf[eIdx:], escape.escape)
			if idx == -1 {
				break
			}
			idx += eIdx

			// 6. If the escape character was found in the last byte of current index, we need to read more bytes to the buffer.
			if idx == len(buf)-1 {
				break
			}

			// 7. Check if the next byte in the buffer is the escaped term.
			if buf[idx+1] == escape.escapedTerm {
				foundTerminator = true
				break
			}

			// 8. If the next byte in the buffer is not the escaped term, the next byte should be a 0x00 byte.
			if buf[idx+1] != escape.escaped00 {
				return n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "malformed bytes value")
			}

			// 9. If the escape00 was found, replace it with the escapeFF byte.
			eIdx = idx + 2
		}
		n += int64(nn)
		escapePosition += int64(nn)

		if foundTerminator {
			break
		}
	}

	// 10. Check if the escape term was found.
	if !foundTerminator {
		return n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "malformed string value")
	}

	// 11. Set the position of the read seeker to the position of the escape term.
	nextPos := curPos + escapePosition + 1
	if _, err = rs.Seek(nextPos, io.SeekStart); err != nil {
		if !errors.Is(err, io.EOF) {
			return n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read string value")
		}
	}
	return n, nil
}
