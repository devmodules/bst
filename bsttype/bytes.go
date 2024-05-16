package bsttype

import (
	"io"
	"strconv"
	"strings"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
)

// Compile-time check to ensure that Bytes implements the Type interface.
var (
	_ Type         = (*Bytes)(nil)
	_ TypeReader   = (*Bytes)(nil)
	_ TypeWriter   = (*Bytes)(nil)
	_ TypeSkipper  = (*Bytes)(nil)
	_ TypeComparer = (*Bytes)(nil)
)

// Compile-time check to ensure that Bytes implements internal interfaces.
var (
	_ copier = (*Bytes)(nil)
)

// Bytes is the Type implementation for the []byte value.
// Binary encoding of the Bytes content is:
// Size(bits)   | Name                     | Description
// -------------+--------------------------+------------
//
//	2         | Header with the flags    | Is a Bytes header containing its flags.
//	6		    | FixedSize Length size    | Header with the binary size of the FixedSize integer.
//	0-64      | FixedSize Value          | The fixed size unsigned integer.
type Bytes struct {
	FixedSize int

	isShared bool
}

// String returns a human-readable representation of the Bytes.
func (x *Bytes) String() string {
	sb := strings.Builder{}
	sb.WriteString("Bytes")
	if x.FixedSize != 0 {
		sb.WriteRune('[')
		sb.WriteString(strconv.Itoa(x.FixedSize))
		sb.WriteRune(']')
	}
	return sb.String()
}

// Kind returns the basic kind of the value.
// Implements the Type interface.
func (*Bytes) Kind() Kind {
	return KindBytes
}

// HasFixedSize returns true if the bytes type has a fixed size.
func (x *Bytes) HasFixedSize() bool {
	return x.FixedSize != 0
}

// CompareType compares for equality between two types.
// Implements the TypeComparer interface.
func (x *Bytes) CompareType(to TypeComparer) bool {
	tx, ok := to.(*Bytes)
	if !ok {
		return false
	}

	return x.FixedSize == tx.FixedSize
}

// SkipType skips the bytes in the reader to the next value.
// Implements the TypeSkipper interface.
func (x *Bytes) SkipType(rs io.ReadSeeker) (int64, error) {
	bt, err := bstio.ReadByte(rs)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to skip bytes type")
	}

	header := bt >> 6
	bt = (bt << 2) >> 2

	// 1. The header byte should have the most significant bit set to 1
	// 	  to indicate that the value has fixed size.
	//    Otherwise, the bytes type content is a single 0x00 byte.
	if bt == 0x00 {
		return 1, nil
	}

	if header&0b00000010 == 0 {
		return 1, bsterr.Err(bsterr.CodeDecodingBinaryValue, "invalid bytes type header byte").
			WithDetails(
				bsterr.D("detail", "header byte should have the most significant bit set to 1"),
				bsterr.D("header_byte", bt),
			)
	}

	// 2. The header byte should have the least significant bit set to 1
	//    to indicate that the value has fixed size.
	//    The other bits of the header should determine binary size of the integer
	//    that describes the length of the bytes value.
	//
	//    Clear the most significant bit by shifting left and right by 1.
	size := (bt << 1) >> 1
	if size > 8 {
		return 1, bsterr.Errf(bsterr.CodeDecodingBinaryValue, "invalid bytes type header byte").
			WithDetails(
				bsterr.D("detail", "header byte should contain binary size of the fixed size integer"),
				bsterr.D("header_byte", bt),
			)
	}

	// 3. Skip the fixed size integer bytes.
	_, err = rs.Seek(int64(size), io.SeekCurrent)
	if err != nil {
		return 1, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip bytes type, fixed size integer")
	}
	return int64(size) + 1, nil
}

// ReadType reads the type from the byte slice.
// Implements the TypeReader interface.
func (x *Bytes) ReadType(r io.Reader) (int, error) {
	// 1. Read the first byte that contains both the size of the fixed size integer and the header flags.
	bt, err := bstio.ReadByte(r)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to read bytes type header")
	}
	totalBytes := 1

	// 2. Extract two most-significant bits of the header byte.
	header := bt >> 6

	// 3. Extract the least-significant bits of the header byte.
	bt = (bt << 2) >> 2

	// 4. The header byte should have the most significant bit set to 1
	// 	  to indicate that the value has fixed size.
	//    Otherwise, the bytes type content is a single 0x00 byte.
	if bt == 0x00 {
		x.FixedSize = 0
		return totalBytes, nil
	}

	// 5. The fixed size flag is expected to be written.
	isFixedSize := header&0b00000010 != 0
	if !isFixedSize {
		return totalBytes, bsterr.Err(bsterr.CodeDecodingBinaryValue, "invalid bytes type header byte").
			WithDetails(
				bsterr.D("detail", "header byte should have the most significant bit set to 1"),
				bsterr.D("header_byte", bt),
			)
	}

	// 6. Read the binary length of the fixed size integer.
	size, n, err := bstio.ReadUintValue(r, bt, false)
	if err != nil {
		return n + 1, err
	}
	totalBytes += n

	x.FixedSize = int(size)
	return totalBytes, nil
}

// WriteType writes the type to the writer.
// Implements the TypeWriter interface.
// The bytes type content is encoded as follows:
//  1. If the type has not defined fixed size, then the bytes type content is a single 0x00 byte.
//  2. If the fixed size is greater than 0, then the bytes type content is a
//     header byte followed by the fixed size integer bytes.
func (x *Bytes) WriteType(w io.Writer) (int, error) {
	header := byte(0)
	if x.FixedSize == 0 {
		err := bstio.WriteByte(w, header)
		if err != nil {
			return 0, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write bytes type")
		}
		return 1, nil
	}

	// 1. The header byte should have the most significant bit set to 1
	//    to indicate that the value has fixed size.
	// 	  10000000 - 0x80
	header |= 0b10000000

	// 2. The header byte should be shifted left by 1 and then ORed with the binary size of the fixed size integer.
	size := bstio.UintSizeHeader(uint(x.FixedSize), false)
	header |= size

	// 3. Write the header.
	if err := bstio.WriteByte(w, header); err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write bytes type")
	}
	bytesWritten := 1

	// 4. Write the fixed size integer bytes.
	n, err := bstio.WriteUintValue(w, uint(x.FixedSize), size, false)
	if err != nil {
		return bytesWritten + n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write bytes type")
	}
	return bytesWritten + n, nil
}

func (x *Bytes) copy(shared bool) Type {
	if !shared {
		return &Bytes{FixedSize: x.FixedSize}
	}
	sb := getSharedBytes()
	sb.FixedSize = x.FixedSize
	return sb
}

//
// Shared Pool
//

var _sharedBytesPool = &sharedPool{defaultSize: 10}

func getSharedBytes() *Bytes {
	v := _sharedBytesPool.pool.Get()
	st, ok := v.(*Bytes)
	if ok {
		return st
	}
	return &Bytes{isShared: true}
}

func putSharedBytes(x *Bytes) {
	if !x.isShared {
		return
	}
	*x = Bytes{isShared: true}
	_sharedBytesPool.pool.Put(x)
}
