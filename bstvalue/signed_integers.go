package bstvalue

import (
	"bytes"
	"fmt"
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
	"github.com/devmodules/bst/internal/iopool"
)

// Compile-time interface check.
var (
	_ Value       = (*Int8Value)(nil)
	_ ValueReader = (*Int8Value)(nil)
	_ ValueWriter = (*Int8Value)(nil)
)

// NewInt8Value creates a new int8 value.
func NewInt8Value(v int8) *Int8Value {
	return &Int8Value{Value: v}
}

// Int8Value is a valuer that returns a int8.
type Int8Value struct {
	Value int8
}

func emptyInt8Value(_ bsttype.Type) Value {
	return &Int8Value{}
}

// Type returns the type of the value.
// Implements the ValueMarshaler interface.
func (*Int8Value) Type() bsttype.Type {
	return bsttype.Int8()
}

// String returns human-readable string representation of the Int8Value.
func (x Int8Value) String() string {
	return fmt.Sprintf("Int8(%d)", x.Value)
}

// Kind returns the kind of the value.
func (*Int8Value) Kind() bsttype.Kind {
	return bsttype.KindInt8
}

// Skip the bytes in the reader to the next value.
func (*Int8Value) Skip(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipInt8(rs)
}

// WriteValue writes the value to a binary format.
// Implements the ValueWriter interface.
func (x *Int8Value) WriteValue(w io.Writer, options bstio.ValueOptions) (int, error) {
	v, err := x.MarshalValue(options)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(v)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write int8 value")
	}
	return n, nil
}

// ReadValue reads the value from a binary format.
// Implements the ValueReader interface.
func (x *Int8Value) ReadValue(r io.Reader, options bstio.ValueOptions) (int, error) {
	v, n, err := bstio.ReadInt8(r, options.Descending)
	if err != nil {
		return n, err
	}

	x.Value = v
	return n, nil
}

// UnmarshalValue unmarshals the value from a binary format.
func (x *Int8Value) UnmarshalValue(in []byte, options bstio.ValueOptions) error {
	if len(in) != 1 {
		return bsterr.Err(bsterr.CodeDecodingBinaryValue, "invalid int8 value size")
	}
	iv, err := bstio.ParseInt8(in[0], options.Descending)
	if err != nil {
		return err
	}

	x.Value = iv
	return nil
}

// MarshalValue returns the value in a binary format.
// If the value is negative, the first bit is set to 0, and for positive values,
// the first bit is set to 1.
// Negative values are converted to uint8 so that with a first 0 bit
// the value is smaller than positive values.
// Implements the ValueMarshaler interface.
func (x *Int8Value) MarshalValue(options bstio.ValueOptions) ([]byte, error) {
	v := x.Value
	var res byte
	if v < 0 {
		res = uint8(v) & bstio.NegativeBit8Mask
	} else {
		res = byte(v) | bstio.PositiveBit8Mask
	}
	if options.Descending {
		res = ^res
	}
	return []byte{res}, nil
}

// Compile time check that Int16Value implements the Value interface.
var (
	_ Value       = (*Int16Value)(nil)
	_ Marshaler   = (*Int16Value)(nil)
	_ Unmarshaler = (*Int16Value)(nil)
	_ ValueWriter = (*Int16Value)(nil)
	_ ValueReader = (*Int16Value)(nil)
)

// Int16Value is a valuer that returns a int16.
type Int16Value struct {
	Value int16
}

// NewInt16Value creates a new int16 value.
func NewInt16Value(v int16) *Int16Value {
	return &Int16Value{Value: v}
}

func emptyInt16Value(_ bsttype.Type) Value {
	return &Int16Value{}
}

// Type returns the type of the value.
// Implements the Value interface.
func (x *Int16Value) Type() bsttype.Type {
	return bsttype.Int16()
}

// String returns human-readable string representation of the Int16Value.
func (x Int16Value) String() string {
	return fmt.Sprintf("Int16(%d)", x.Value)
}

// Kind returns the kind of the value.
// Implements the Value interface.
func (x *Int16Value) Kind() bsttype.Kind {
	return bsttype.KindInt16
}

// Skip seeks over the reader after the value.
// Implements the Value interface.
func (x *Int16Value) Skip(s io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipInt16(s)
}

// ReadValue reads the value from a binary format.
// Implements the ValueReader interface.
func (x *Int16Value) ReadValue(r io.Reader, o bstio.ValueOptions) (int, error) {
	iv, n, err := bstio.ReadInt16(r, o.Descending)
	if err != nil {
		return n, err
	}
	x.Value = iv
	return n, nil
}

// UnmarshalValue unmarshals the value from a binary format.
// Implements the ValueMarshaler interface.
func (x *Int16Value) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	iv, err := bstio.ParseInt16(in, o.Descending)
	if err != nil {
		return err
	}
	x.Value = iv
	return nil
}

// WriteValue writes the value to a binary format.
// Implements the ValueWriter interface.
func (x *Int16Value) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	return bstio.WriteInt16(w, x.Value, o.Descending)
}

// MarshalValue returns the value in a binary format.
// The value is encoded in big endian encoding.
// Implements the Value interface.
func (x *Int16Value) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	return bstio.MarshalInt16(x.Value, o.Descending), nil
}

var (
	_ Value       = (*Int32Value)(nil)
	_ Marshaler   = (*Int32Value)(nil)
	_ Unmarshaler = (*Int32Value)(nil)
	_ ValueWriter = (*Int32Value)(nil)
	_ ValueReader = (*Int32Value)(nil)
)

// Int32Value is a valuer that returns a int32.
type Int32Value struct {
	Value int32
}

// NewInt32Value returns a new Int32Value.
func NewInt32Value(i int32) *Int32Value {
	return &Int32Value{Value: i}
}

func emptyInt32Value(_ bsttype.Type) Value {
	return &Int32Value{}
}

// String returns human-readable string representation of the Int32Value.
func (x Int32Value) String() string {
	return fmt.Sprintf("Int32(%d)", x.Value)
}

// Type returns the type of the value.
// Implements the Value interface.
func (x *Int32Value) Type() bsttype.Type {
	return bsttype.Int32()
}

// Kind returns the kind of the value.
// Implements the Value interface.
func (x *Int32Value) Kind() bsttype.Kind {
	return bsttype.KindInt32
}

// Skip seeks the input byte reader to the next value.
func (x *Int32Value) Skip(s io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipInt32(s)
}

// WriteValue writes the value to the output stream.
func (x *Int32Value) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	return bstio.WriteInt32(w, x.Value, o.Descending)
}

// UnmarshalValue unmarshals the value from a binary format.
// Implements the Unmarshaler interface.
func (x *Int32Value) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	iv, err := bstio.ParseInt32(in, o.Descending)
	if err != nil {
		return err
	}
	x.Value = iv
	return nil
}

// ReadValue reads the value from the input byte reader.
// Implements the ValueReader interface.
func (x *Int32Value) ReadValue(r io.Reader, o bstio.ValueOptions) (int, error) {
	iv, n, err := bstio.ReadInt32(r, o.Descending)
	if err != nil {
		return n, err
	}
	x.Value = iv
	return n, nil
}

// MarshalValue returns the value in a binary format.
// The value is encoded in big endian encoding.
// Implements the Value interface.
func (x *Int32Value) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	i32 := x.Value
	desc := o.Descending
	res := bstio.MarshalInt32(i32, desc)
	return res, nil
}

// Compile time checks for the Int64Value.
var _ Value = (*Int64Value)(nil)

// Int64Value is a valuer that returns a int64.
type Int64Value struct {
	Value int64
}

// NewInt64Value returns a new Int64Value.
func NewInt64Value(i int64) *Int64Value {
	return &Int64Value{Value: i}
}

func emptyInt64Value(_ bsttype.Type) Value {
	return &Int64Value{}
}

// Type implements the ValueMarshaler interface.
func (x *Int64Value) Type() bsttype.Type {
	return bsttype.Int64()
}

// String returns human-readable string representation of the Int64Value.
func (x Int64Value) String() string {
	return fmt.Sprintf("Int64(%d)", x.Value)
}

// Kind implements the ValueMarshaler interface.
func (x *Int64Value) Kind() bsttype.Kind {
	return bsttype.KindInt64
}

// ReadValue reads the binary value from the input reader.
func (x *Int64Value) ReadValue(r io.Reader, o bstio.ValueOptions) (int, error) {
	v, n, err := bstio.ReadInt64(r, o.Descending)
	if err != nil {
		return n, err
	}

	x.Value = v
	return n, nil
}

// WriteValue writes the binary value to the output writer.
// Implements the ValueMarshaler interface.
func (x *Int64Value) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	return bstio.WriteInt64(w, x.Value, o.Descending)
}

// Skip implements the Value interface.
func (x *Int64Value) Skip(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipInt64(rs)
}

// MarshalValue marshals the value to a binary format.
// The value is encoded in big endian encoding.
// Implements the Marshaler interface.
func (x *Int64Value) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	return bstio.MarshalInt64(x.Value, o.Descending), nil
}

// UnmarshalValue unmarshals the value from a binary format.
// Implements the Unmarshaler interface.
func (x *Int64Value) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	iv, err := bstio.ParseInt64(in, o.Descending)
	if err != nil {
		return err
	}

	x.Value = iv
	return nil
}

// Compile time checks for the IntValue.
var _ Value = (*IntValue)(nil)

// IntValue is a valuer that returns a int.
type IntValue struct {
	Value int
}

// NewIntValue returns a new IntValue.
func NewIntValue(i int) *IntValue {
	return &IntValue{Value: i}
}

func emptyIntValue(_ bsttype.Type) Value {
	return &IntValue{}
}

// Type implements the ValueMarshaler interface.
func (x *IntValue) Type() bsttype.Type {
	return bsttype.Int()
}

// String returns human-readable string representation of the IntValue.
func (x IntValue) String() string {
	return fmt.Sprintf("Int(%d)", x.Value)
}

// Kind implements the ValueMarshaler interface.
func (x *IntValue) Kind() bsttype.Kind {
	return bsttype.KindInt
}

// ReadValue reads the binary value from the input reader.
// Implements the ValueMarshaler interface.
func (x *IntValue) ReadValue(r io.Reader, o bstio.ValueOptions) (int, error) {
	v, n, err := bstio.ReadInt(r, o.Descending, o.Comparable)
	if err != nil {
		return 0, err
	}
	x.Value = v
	return n, nil
}

// WriteValue writes the binary value to the output writer.
// Implements the ValueMarshaler interface.
func (x *IntValue) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	return bstio.WriteInt(w, x.Value, o.Descending, o.Comparable)
}

// Skip implements the Value interface.
func (x *IntValue) Skip(rs io.ReadSeeker, o bstio.ValueOptions) (int64, error) {
	return bstio.SkipInt(rs, o.Descending, o.Comparable)
}

// MarshalValue marshals the value to a binary format.
// The value is encoded in big endian encoding.
// Implements the Marshaler interface.
func (x *IntValue) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	buf := iopool.GetBuffer(nil)
	_, err := bstio.WriteInt(buf, x.Value, o.Descending, o.Comparable)
	if err != nil {
		iopool.ReleaseBuffer(buf)
		return nil, err
	}
	cp := buf.BytesCopy()
	iopool.ReleaseBuffer(buf)
	return cp, nil
}

// UnmarshalValue unmarshals the value from a binary format.
// Implements the Unmarshaler interface.
func (x *IntValue) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	v, _, err := bstio.ReadInt(bytes.NewReader(in), o.Descending, o.Comparable)
	if err != nil {
		return err
	}
	x.Value = v
	return nil
}
