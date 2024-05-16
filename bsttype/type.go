package bsttype

import (
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
)

// Type defines a descriptor of the value type.
type Type interface {
	// Kind returns the basic kind of the value.
	Kind() Kind
	// String gets a string representation of the type.
	String() string
}

type copier interface {
	copy(shared bool) Type
}

// TypeSkipper is the interface used to skip the binary data for specific type.
// It is used only if the Type is complex and have some specific data.
type TypeSkipper interface {
	SkipType(io.ReadSeeker) (int64, error)
}

// TypeReader is the interface for reading types.
type TypeReader interface {
	ReadType(r io.Reader) (int, error)
}

// TypeWriter is the interface for writing types.
type TypeWriter interface {
	WriteType(w io.Writer) (int, error)
}

// TypeComparer is the interface used to compare different types.
// It is used only for complex types.
type TypeComparer interface {
	CompareType(to TypeComparer) bool
}

// TypesEqual compares two types.
func TypesEqual(t1, t2 Type) bool {
	if t1.Kind() != t2.Kind() {
		return false
	}

	tc1, ok1 := t1.(TypeComparer)
	tc2, ok2 := t2.(TypeComparer)
	if !ok1 && !ok2 {
		return true
	}
	if ok1 && !ok2 {
		return false
	}
	if !ok1 && ok2 {
		return false
	}

	return tc1.CompareType(tc2)
}

// ReadType reads the binary representation of the Type from the reader.
func ReadType(r io.Reader, sharedDefs bool) (Type, int, error) {
	// 1. Read the type kind.
	bh, err := bstio.ReadByte(r)
	if err != nil {
		return nil, 0, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to read type header")
	}
	total := 1

	// 2. Create the type for given kind.
	et := emptyKindType(Kind(bh), sharedDefs)
	if et.Kind() == KindUndefined {
		return nil, total, bsterr.Err(bsterr.CodeEncodingBinaryValue, "undefined Kind for value type")
	}

	// 3. Check if the type ha a ReadType function.
	tr, ok := et.(TypeReader)
	if !ok {
		return et, total, nil
	}

	// 4. Read the type content.
	var n int
	n, err = tr.ReadType(r)
	if err != nil {
		return nil, total + n, err
	}
	return et, total + n, nil
}

// WriteType writes the type in binary representation to the writer.
// Returns the number of bytes written.
func WriteType(w io.Writer, vt Type) (int, error) {
	err := bstio.WriteByte(w, byte(vt.Kind()))
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write type").
			WithDetail("type", vt.Kind())
	}
	total := 1

	// 1.2. If the type implements TypeContent interface, write the content.
	tc, ok := vt.(TypeWriter)
	if !ok {
		return total, nil
	}

	var bw int
	bw, err = tc.WriteType(w)
	return total + bw, err
}

// SkipType skips the type in binary representation.
func SkipType(rs io.ReadSeeker) (int64, error) {
	// 1. Parse the type header.
	bk, err := bstio.ReadByte(rs)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read type header")
	}
	bytesSkipped := int64(1)

	// 2. From the type header get its Kind.
	et := emptyKindType(Kind(bk), true)
	defer PutSharedType(et)

	// 3. Check if the elem type implements the TypeContent interface.
	var n int64
	tc, ok := et.(TypeSkipper)
	if !ok {
		return bytesSkipped, nil
	}
	// 4. Skip the type content and return the number of bytes skipped.
	n, err = tc.SkipType(rs)
	return bytesSkipped + n, err
}
