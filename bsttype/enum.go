package bsttype

import (
	"fmt"
	"io"
	"strings"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
)

// Compile-time check to ensure that Enum implements the Type  interface.
var (
	_ Type         = (*Enum)(nil)
	_ TypeReader   = (*Enum)(nil)
	_ TypeWriter   = (*Enum)(nil)
	_ TypeSkipper  = (*Enum)(nil)
	_ TypeComparer = (*Enum)(nil)
)

// Compile-time checks for internal interfaces
var (
	_ copier = (*Enum)(nil)
)

type (
	// Enum is the Type implementation for the enum.
	// Binary representation:
	// Size(bits)   | Name                | Description
	// -------------+---------------------+------------
	//    8		    | ValueBytes Size     | Header with the size of the value bytes length.
	//    8		    | Elements Count Size | This is the size of the elements number.
	//    8-64      | Elements Count      | The number of elements in the enum.
	//    N * Count | Element             | Binary representation of the elements.
	Enum struct {
		// ValueBytes determines the number of bytes used to store the enum buffIndex value.
		ValueBytes uint8
		// Elements is the list of enum elements.
		Elements []EnumElement

		isShared bool
	}

	// EnumElement is the element of the Enum.
	// Binary representation:
	// Size(bits)   | Name                | Description
	// -------------+---------------------+------------
	//    8 		| String Length Size  | Header with the size of the string length.
	//    0-64      | String Length       | The length of the String (0 if the length size was marked as 0x00).
	//    0-N       | String              | String representing the string of the element.
	//    N         | Index               | The buffIndex of the element. The value size depends on the ValueBytes of the Enum.
	EnumElement struct {
		// String is the string value of the element.
		String string
		// Index is the buffIndex of the element.
		Index uint
	}
)

// String returns a human-readable representation of the Enum.
// Implements Type interface.
// Example: Enum(Elements: [{String: "A", Index: 0}, {String: "B", Index: 1}], ValueBits: 8)
func (x *Enum) String() string {
	var sb strings.Builder
	sb.WriteString("Enum(Elements: [")
	for i, e := range x.Elements {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("{String: %q, Index: %d}", e.String, e.Index))
	}
	sb.WriteString("], ValueBytes: ")
	sb.WriteString(fmt.Sprintf("%d", x.ValueBytes))
	sb.WriteString(")")
	return sb.String()
}

// Kind returns the basic kind of the value.
func (*Enum) Kind() Kind {
	return KindEnum
}

// StringIndex returns the buffIndex of the string value.
func (x *Enum) StringIndex(v string) (uint, bool) {
	for _, ev := range x.Elements {
		if ev.String == v {
			return ev.Index, true
		}
	}
	return 0, false
}

// IndexString returns the buffIndex of the string value.
func (x *Enum) IndexString(i uint) (string, bool) {
	for _, ev := range x.Elements {
		if ev.Index == i {
			return ev.String, true
		}
	}
	return "", false
}

// CompareType returns true if the types are equal.
// Implements the TypeComparer interface.
func (x *Enum) CompareType(to TypeComparer) bool {
	xto, ok := to.(*Enum)
	if !ok {
		return false
	}

	if len(x.Elements) != len(xto.Elements) {
		return false
	}

	for i, v := range x.Elements {
		if v != xto.Elements[i] {
			return false
		}
	}
	return true
}

// SkipType skips the bytes in the reader to the next value.
// Implements the TypeSkipper interface.
func (x *Enum) SkipType(rs io.ReadSeeker) (int64, error) {
	// 1. Read the ValueBytes of the enum.
	bt, err := bstio.ReadByte(rs)
	if err != nil {
		return 0, err
	}

	bytesSkipped := int64(1)

	// 3. Read the length of the enum elements.
	length, n, err := bstio.ReadUint(rs, false)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to read enum value bytes length")
	}
	bytesSkipped += int64(n)

	// 4. Skip all elements of the enum.
	var n64 int64
	for i := uint(0); i < length; i++ {
		// 4.1. Skip the String field.
		n64, err = bstio.SkipNonComparableString(rs, false)
		if err != nil {
			return bytesSkipped, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to skip enum element String field")
		}
		bytesSkipped += n64

		// 4.2. Skip the Index field.
		switch bt {
		case bstio.BinarySizeUint8:
			n64, err = bstio.SkipUint8Value(rs)
		case bstio.BinarySizeUint16:
			n64, err = bstio.SkipUint16(rs)
		case bstio.BinarySizeUint32:
			n64, err = bstio.SkipUint32(rs)
		case bstio.BinarySizeUint64:
			n64, err = bstio.SkipUint64(rs)
		case bstio.BinarySizeZero:
			n64, err = bstio.SkipUint(rs, false)
		default:
			return 0, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to skip enum element Index field")
		}
		if err != nil {
			return bytesSkipped, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to skip enum element Index field")
		}
		bytesSkipped += n64
	}

	return bytesSkipped, nil
}

// ReadType reads the value from the byte slice.
// Implements the TypeReader interface.
func (x *Enum) ReadType(r io.Reader) (int, error) {
	// 1. Read the ValueBytes of the enum which also is array length size header.
	bt, err := bstio.ReadByte(r)
	if err != nil {
		return 0, err
	}
	bytesRead := 1

	// 2. Set the ValueBytes.
	x.ValueBytes = bt

	// 3. Read the elements array length.
	length, n, err := bstio.ReadUint(r, false)
	if err != nil {
		return 0, err
	}
	bytesRead += n

	// 4. Read the elements array.
	var read int
	x.Elements = make([]EnumElement, length)
	for i := uint(0); i < length; i++ {
		var (
			str string
			ui  uint
		)

		// 4.1. Read the String field.
		str, read, err = bstio.ReadStringNonComparable(r, false)
		if err != nil {
			return 0, err
		}
		bytesRead += read

		// 4.2. Read the Index field.
		switch bt {
		case bstio.BinarySizeZero:
			ui, read, err = bstio.ReadUint(r, false)
			if err != nil {
				return bytesRead + read, err
			}
		case bstio.BinarySizeUint8:
			var v uint8
			v, read, err = bstio.ReadUint8(r, false)
			if err != nil {
				return bytesRead + read, err
			}
			ui = uint(v)
		case bstio.BinarySizeUint16:
			var v uint16
			v, read, err = bstio.ReadUint16(r, false)
			if err != nil {
				return bytesRead + read, err
			}
			ui = uint(v)
		case bstio.BinarySizeUint32:
			var v uint32
			v, read, err = bstio.ReadUint32(r, false)
			if err != nil {
				return bytesRead + read, err
			}
			ui = uint(v)
		case bstio.BinarySizeUint64:
			var v uint64
			v, read, err = bstio.ReadUint64(r, false)
			if err != nil {
				return bytesRead + read, err
			}
			ui = uint(v)
		default:
			return bytesRead, bsterr.Err(bsterr.CodeTypeConstraintViolation, "invalid enum size value").
				WithDetail("size", bt)
		}
		bytesRead += read

		// 4.3. Set the element with the read values.
		x.Elements[i] = EnumElement{
			String: str,
			Index:  ui,
		}
	}

	return bytesRead, nil
}

// WriteType writes the value to the byte slice.
// Implements the TypeWriter interface.
func (x *Enum) WriteType(w io.Writer) (int, error) {
	// 1. Write byte the ValueBytes of the enum which also is array length size header.
	n, err := bstio.WriteUint8(w, x.ValueBytes, false)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write enum value bytes")
	}
	written := n

	// 2. Write array size.
	n, err = bstio.WriteUint(w, uint(len(x.Elements)), false)
	if err != nil {
		return written + n, err
	}
	written += n

	// 3. Write array elements one by one.
	for _, v := range x.Elements {
		// 3.1. Write element String value.
		n, err = bstio.WriteString(w, v.String, false, false)
		if err != nil {
			return written + n, err
		}
		written += n

		// 3.2. Write element Index value.
		switch x.ValueBytes {
		case bstio.BinarySizeZero:
			n, err = bstio.WriteUint(w, v.Index, false)
		case bstio.BinarySizeUint8:
			n, err = bstio.WriteUint8(w, uint8(v.Index), false)
		case bstio.BinarySizeUint16:
			n, err = bstio.WriteUint16(w, uint16(v.Index), false)
		case bstio.BinarySizeUint32:
			n, err = bstio.WriteUint32(w, uint32(v.Index), false)
		case bstio.BinarySizeUint64:
			n, err = bstio.WriteUint64(w, uint64(v.Index), false)
		default:
			return written, bsterr.Err(bsterr.CodeTypeConstraintViolation, "invalid enum size value")
		}
		if err != nil {
			return written + n, err
		}
		written += n
	}
	return written, nil
}

func (x *Enum) copy(shared bool) Type {
	var cp *Enum
	if shared {
		cp = getSharedEnum()
	} else {
		cp = &Enum{}
	}
	cp.ValueBytes = x.ValueBytes
	if cap(cp.Elements) < len(x.Elements) {
		cp.Elements = make([]EnumElement, len(x.Elements))
	} else {
		cp.Elements = cp.Elements[:len(x.Elements)]
	}
	copy(cp.Elements, x.Elements)
	return cp
}

//
// Shared Pool
//

var _sharedEnumPool = &sharedPool{defaultSize: 10}

func getSharedEnum() *Enum {
	v := _sharedEnumPool.pool.Get()
	st, ok := v.(*Enum)
	if ok {
		return st
	}
	return &Enum{
		Elements: make([]EnumElement, 0, _sharedEnumPool.defaultSize),
		isShared: true,
	}
}

func putSharedEnum(x *Enum) {
	if !x.isShared {
		return
	}
	length := cap(x.Elements)
	*x = Enum{isShared: true, Elements: x.Elements[:0]}
	_sharedEnumPool.put(x, length)
}
