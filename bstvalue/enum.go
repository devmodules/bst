package bstvalue

import (
	"bytes"
	"fmt"
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// Compile-time check to ensure that EnumValue implements the Value interface.
var _ Value = (*EnumValue)(nil)

// EnumValue is the Value implementation for the enum.
// Implements the Value interface.
type EnumValue struct {
	EnumType *bsttype.Enum
	Index    int
}

// NewEnumStringValue creates a new EnumValue with the given string value.
func NewEnumStringValue(et *bsttype.Enum, v string) (*EnumValue, error) {
	uv, found := et.StringIndex(v)
	if !found {
		return nil, bsterr.Err(bsterr.CodeTypeConstraintViolation, "value not found in enum type definition")
	}
	return &EnumValue{
		EnumType: et,
		Index:    int(uv),
	}, nil
}

// MustNewEnumValue creates a new EnumValue with the given string value.
// Panics on error.
func MustNewEnumValue(et *bsttype.Enum, index uint) *EnumValue {
	uv, err := NewEnumValue(et, index)
	if err != nil {
		panic(err)
	}
	return uv
}

// NewEnumValue creates a new EnumValue with the given value.
func NewEnumValue(et *bsttype.Enum, v uint) (*EnumValue, error) {
	_, found := et.IndexString(v)
	if !found {
		return nil, bsterr.Err(bsterr.CodeTypeConstraintViolation, "value not found in enum type definition")
	}
	return &EnumValue{
		EnumType: et,
		Index:    int(v),
	}, nil
}

// EmptyEnumValue creates a new EnumValue with the given value.
func EmptyEnumValue(et *bsttype.Enum) *EnumValue {
	return &EnumValue{
		EnumType: et,
		Index:    -1,
	}
}

func emptyEnumValue(t bsttype.Type) Value {
	return &EnumValue{EnumType: t.(*bsttype.Enum)}
}

// String returns a human-readable description of the EnumValue.
func (x *EnumValue) String() string {
	idxName, ok := x.EnumType.IndexString(uint(x.Index))
	if !ok {
		return "Enum(Undefined)"
	}
	return fmt.Sprintf("Enum(%d: %s)", x.Index, idxName)
}

// IndexString returns the string value of the enum.
func (x *EnumValue) IndexString() (string, bool) {
	s, ok := x.EnumType.IndexString(uint(x.Index))
	return s, ok
}

// Kind returns the basic kind of the value.
// Implements the Value interface.
func (x *EnumValue) Kind() bsttype.Kind {
	return bsttype.KindEnum
}

// Skip skips the bytes in the reader to the next value.
// Implements the Value interface.
func (x *EnumValue) Skip(rs io.ReadSeeker, o bstio.ValueOptions) (int64, error) {
	return bstio.SkipEnumIndex(rs, x.EnumType.ValueBytes, o.Descending)
}

// MarshalValue marshals the value to the byte slice.
// Implements the Value interface.
func (x *EnumValue) MarshalValue(options bstio.ValueOptions) ([]byte, error) {
	return bstio.MarshalEnumIndex(x.Index, x.EnumType.ValueBytes, options.Descending)
}

// UnmarshalValue unmarshals the value from the byte slice.
func (x *EnumValue) UnmarshalValue(in []byte, options bstio.ValueOptions) error {
	_, err := x.ReadValue(bytes.NewReader(in), options)
	return err
}

// ReadValue reads the value from the byte slice.
// Implements the Value interface.
func (x *EnumValue) ReadValue(r io.Reader, options bstio.ValueOptions) (int, error) {
	index, n, err := bstio.ReadEnumIndex(r, x.EnumType.ValueBytes, options.Descending)
	if err != nil {
		return n, err
	}
	x.Index = index

	return n, nil
}

// WriteValue writes the value to the byte slice.
func (x *EnumValue) WriteValue(w io.Writer, options bstio.ValueOptions) (int, error) {
	bin, err := x.MarshalValue(options)
	if err != nil {
		return 0, err
	}

	n, err := w.Write(bin)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "error writing enum value")
	}
	return n, nil
}

// Type returns the type of the value.
func (x *EnumValue) Type() bsttype.Type {
	return x.EnumType
}
