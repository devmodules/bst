package bstvalue

import (
	"bytes"
	"fmt"
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bstskip"
	"github.com/devmodules/bst/bsttype"
)

// Compile-time check for interface implementation.
var _ Value = (*AnyValue)(nil)

// AnyValue is the value wrapper that could take any value.
// This could be used for schemaless values.
// The binary format of the AnyValue looks is a combination of the
// embedded Type and then the value itself.
// If embedded value is composed of multiple values (i.e. an array of nullable structs).
// then all subsequent values are also encoded along with their types.
type AnyValue struct {
	Value Value
}

// EmptyAnyValue returns an empty value that could take any value.
func EmptyAnyValue() *AnyValue {
	return &AnyValue{}
}

// AnyValueOf returns a value that could take any value.
func AnyValueOf(v Value) *AnyValue {
	return &AnyValue{v}
}

func emptyAnyValue(_ bsttype.Type) Value {
	return &AnyValue{}
}

// Type implements the Value interface.
// It returns the type of the value.
func (x *AnyValue) Type() bsttype.Type {
	return bsttype.Any()
}

// String returns a human-readable representation of the AnyValue.
func (x *AnyValue) String() string {
	return fmt.Sprintf("Any(%s)", x.Value.String())
}

// Elem implements the ValueElement interface.
func (x *AnyValue) Elem() Value {
	return x.Value
}

// Kind implements the Value interface.
// It returns the basic kind of the value.
func (x *AnyValue) Kind() bsttype.Kind {
	return x.Value.Kind()
}

// Skip implements the Value interface.
// It skips the bytes in the reader to the next value.
func (x *AnyValue) Skip(rs io.ReadSeeker, o bstio.ValueOptions) (int64, error) {
	return bstskip.SkipAny(rs, o)
}

// MarshalValue implements the Value interface.
// It marshals the value to a binary database format.
func (x *AnyValue) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	var buf bytes.Buffer
	_, err := x.WriteValue(&buf, o)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalValue implements the Value interface.
// It unmarshals the value from a binary database format.
func (x *AnyValue) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	_, err := x.ReadValue(bytes.NewReader(in), o)
	return err
}

// ReadValue implements the Value interface.
// It reads the value from a reader.
func (x *AnyValue) ReadValue(r io.Reader, o bstio.ValueOptions) (int, error) {
	// 1. Read the type of the value.
	// 1.1. Read the header of the type.
	et, i, err := bsttype.ReadType(r, false)
	if err != nil {
		return i, err
	}
	bytesRead := i
	v := EmptyValueOf(et)

	var n int
	n, err = v.ReadValue(r, o)
	if err != nil {
		return bytesRead + n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to read value").
			WithDetail("type", et.Kind())
	}
	bytesRead += n
	x.Value = v
	return bytesRead, nil
}

// WriteValue implements the Value interface.
// It writes the value to a writer.
func (x *AnyValue) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	// 1. Write the type of the value.
	// 1.1. Write the header of the type.
	vt := x.Value.Type()
	total, err := bsttype.WriteType(w, vt)
	if err != nil {
		return total, err
	}

	// 2. Write the value.
	var bw int
	bw, err = x.Value.WriteValue(w, o)
	if err != nil {
		return total + bw, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write value").
			WithDetail("type", vt.Kind())
	}
	total += bw
	return total, nil
}
