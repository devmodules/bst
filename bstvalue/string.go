package bstvalue

import (
	"bytes"
	"fmt"
	"io"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
	"github.com/devmodules/bst/internal/iopool"
)

// Compile-time check to ensure that StringValue implements the Value interface.
var _ Value = (*StringValue)(nil)

// StringValue is the value descriptor for the string.
// A string could be either stored in comparable or non-comparable binary.
// A comparable binary guarantees that the string is stored in an ordered way, however it
// is slower to skip over that binary, as there is no way to know the length of the string
// in advance - and the value needs to be read until the escape characters.
type StringValue struct {
	Value string
}

// NewStringValue returns a new StringValue with the given string.
func NewStringValue(v string) *StringValue {
	return &StringValue{Value: v}
}

// EmptyStringValue returns a new StringValue with an empty string.
func EmptyStringValue() *StringValue {
	return &StringValue{}
}

func emptyStringValue(_ bsttype.Type) Value {
	return &StringValue{}
}

// String returns the human-readable string representation of the StringValue.
func (x StringValue) String() string {
	return fmt.Sprintf("String(%q)", x.Value)
}

// Type returns the type of the value.
// Implements the Value interface.
func (x *StringValue) Type() bsttype.Type {
	return bsttype.String()
}

// Kind returns the basic kind of the value.
// Implements the Value interface.
func (x *StringValue) Kind() bsttype.Kind {
	return bsttype.KindString
}

// UnmarshalValue reads the value from the byte slice.
// Implements the Value interface.
func (x *StringValue) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	v, _, err := bstio.ReadString(bytes.NewReader(in), o.Descending, o.Comparable)
	if err != nil {
		return err
	}

	x.Value = v
	return nil
}

// ReadValue reads the value from the byte slice.
// Implements the Value interface.
func (x *StringValue) ReadValue(r io.Reader, o bstio.ValueOptions) (int, error) {
	v, n, err := bstio.ReadString(r, o.Descending, o.Comparable)
	if err != nil {
		return n, err
	}

	x.Value = v
	return n, nil
}

// WriteValue writes the value to the writer.
// Implements the Value interface.
func (x *StringValue) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	return bstio.WriteString(w, x.Value, o.Descending, o.Comparable)
}

// Skip the bytes in the reader to the next value.
// Implements the Value interface.
func (x *StringValue) Skip(rs io.ReadSeeker, o bstio.ValueOptions) (int64, error) {
	return bstio.SkipString(rs, o.Descending, o.Comparable)
}

// MarshalValue writes the value to the byte slice.
// Implements the Value interface.
func (x *StringValue) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	buf := iopool.GetBuffer(nil)
	defer iopool.ReleaseBuffer(buf)
	_, err := bstio.WriteString(buf, x.Value, o.Descending, o.Comparable)
	if err != nil {
		return nil, err
	}
	return buf.BytesCopy(), nil
}
