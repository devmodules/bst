package bstvalue

import (
	"bytes"
	"fmt"
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// Uint8Value is a valuer that returns a uint8.
type Uint8Value struct {
	Value uint8
}

// NewUint8Value returns a new Uint8Value with the given value.
func NewUint8Value(v uint8) *Uint8Value {
	return &Uint8Value{Value: v}
}

func emptyUint8Value(_ bsttype.Type) Value {
	return &Uint8Value{}
}

// String returns human-readable representation of the value.
func (x Uint8Value) String() string {
	return fmt.Sprintf("Uint8(%d)", x.Value)
}

// Type returns the type of the value.
// Implements Value interface.
func (x *Uint8Value) Type() bsttype.Type {
	return bsttype.Uint8()
}

// Kind returns the kind of the value.
// Implements the Value interface.
func (x *Uint8Value) Kind() bsttype.Kind {
	return bsttype.KindUint8
}

// Skip seeks through the input reader to skip the value.
// Implements the Value interface.
func (x *Uint8Value) Skip(s io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipUint8Value(s)
}

// WriteValue writes the value to the writer.
// Implements the Value interface.
func (x *Uint8Value) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	return bstio.WriteUint8(w, x.Value, o.Descending)
}

// UnmarshalValue decodes the value from a binary format.
// Implements the Unmarshaler interface.
func (x *Uint8Value) UnmarshalValue(in []byte, options bstio.ValueOptions) error {
	if len(in) != 1 {
		return bsterr.Err(bsterr.CodeDecodingBinaryValue, "invalid uint8 value length").
			WithDetails(
				bsterr.D("length", len(in)),
				bsterr.D("expected", 1),
			)
	}

	v, err := bstio.ParseUint8Value(in[0], options.Descending)
	if err != nil {
		return err
	}
	x.Value = v
	return nil
}

// ReadValue reads the value from the reader.
// Implements the Value interface.
func (x *Uint8Value) ReadValue(r io.Reader, options bstio.ValueOptions) (int, error) {
	v, n, err := bstio.ReadUint8(r, options.Descending)
	if err != nil {
		return n, err
	}
	x.Value = v
	return n, nil
}

// MarshalValue returns the value in a binary format.
// Implements the ValueMarshaler interface.
func (x Uint8Value) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	return bstio.MarshalUint8(x.Value, o.Descending), nil
}

// GoValue returns the value in a go format.
func (x Uint8Value) GoValue() interface{} {
	return x.Value
}

// Compile-time check for Uint16Value implements the Value interface.
var _ Value = (*Uint16Value)(nil)

// Uint16Value is a valuer that returns a uint16.
type Uint16Value struct {
	Value uint16
}

// NewUint16Value returns a new Uint16Value with the given value.
func NewUint16Value(v uint16) *Uint16Value {
	return &Uint16Value{Value: v}
}

func emptyUint16Value(_ bsttype.Type) Value {
	return &Uint16Value{}
}

// String returns human-readable representation of the value.
func (x Uint16Value) String() string {
	return fmt.Sprintf("Uint16(%d)", x.Value)
}

// Type returns the type of the value.
// Implements Value interface.
func (x *Uint16Value) Type() bsttype.Type {
	return bsttype.Uint16()
}

// Kind returns the kind of the value.
// Implements the Value interface.
func (x *Uint16Value) Kind() bsttype.Kind {
	return bsttype.KindUint16
}

// Skip seeks through the input reader to skip the value.
// Implements the Value interface.
func (x *Uint16Value) Skip(s io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipUint16(s)
}

// WriteValue writes the value to the writer.
func (x *Uint16Value) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	return bstio.WriteUint16(w, x.Value, o.Descending)
}

// ReadValue reads the value from the reader.
func (x *Uint16Value) ReadValue(r io.Reader, options bstio.ValueOptions) (int, error) {
	v, n, err := bstio.ReadUint16(r, options.Descending)
	if err != nil {
		return n, err
	}
	x.Value = v
	return n, nil
}

// UnmarshalValue decodes the value from a binary format.
// Implements the encoding.BinaryUnmarshaler interface.
func (x *Uint16Value) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	v, err := bstio.ParseUint16(in, o.Descending)
	if err != nil {
		return err
	}
	x.Value = v
	return nil
}

// MarshalValue returns the value in a binary format.
// The value is encoded in little endian encoding.
// Implements the ValueMarshaler interface.
func (x *Uint16Value) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	return bstio.MarshalUint16(x.Value, o.Descending), nil
}

// BinarySize returns the size of the binary value.
func (x *Uint16Value) BinarySize() int {
	return 2
}

// Compile-time check for Uint32Value implements the Value interface.
var _ Value = (*Uint32Value)(nil)

// Uint32Value is a valuer that returns a uint32.
type Uint32Value struct {
	Value uint32
}

// NewUint32Value returns a new Uint32Value with the given value.
func NewUint32Value(v uint32) *Uint32Value {
	return &Uint32Value{Value: v}
}

func emptyUint32Value(_ bsttype.Type) Value {
	return &Uint32Value{}
}

// String returns human-readable representation of the value.
func (x Uint32Value) String() string {
	return fmt.Sprintf("Uint32(%d)", x.Value)
}

// Type returns the type of the value.
// Implements Value interface.
func (x *Uint32Value) Type() bsttype.Type {
	return bsttype.Uint32()
}

// Kind returns the kind of the value.
// Implements the Value interface.
func (x *Uint32Value) Kind() bsttype.Kind {
	return bsttype.KindUint32
}

// Skip seeks through the input reader to skip the value.
// Implements the Value interface.
func (x *Uint32Value) Skip(s io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipUint32(s)
}

// WriteValue writes the value to the writer.
// Implements the Value interface.
func (x *Uint32Value) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	return bstio.WriteUint32(w, x.Value, o.Descending)
}

// ReadValue reads the value from the reader.
// Implements the Value interface.
func (x *Uint32Value) ReadValue(r io.Reader, options bstio.ValueOptions) (int, error) {
	v, n, err := bstio.ReadUint32(r, options.Descending)
	if err != nil {
		return n, err
	}

	x.Value = v
	return n, nil
}

// UnmarshalValue decodes the value from a binary format.
// Implements the encoding.BinaryUnmarshaler interface.
func (x *Uint32Value) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	v, err := bstio.ParseUint32(in, o.Descending)
	if err != nil {
		return err
	}
	x.Value = v
	return nil
}

// MarshalValue returns the value in a binary format.
// The value is encoded in big endian encoding.
// Implements the ValueMarshaler interface.
func (x *Uint32Value) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	v := x.Value
	desc := o.Descending
	res := bstio.MarshalUint32(v, desc)
	return res, nil
}

// BinarySize returns the size of the binary value.
func (x Uint32Value) BinarySize() int {
	return 4
}

// Uint64Value is a valuer that returns a uint64.
type Uint64Value struct {
	Value uint64
}

// NewUint64Value returns a new Uint64Value with the given value.
func NewUint64Value(v uint64) *Uint64Value {
	return &Uint64Value{Value: v}
}

func emptyUint64Value(_ bsttype.Type) Value {
	return &Uint64Value{}
}

// String returns human-readable representation of the value.
func (x Uint64Value) String() string {
	return fmt.Sprintf("Uint64(%d)", x.Value)
}

// Type returns the type of the value.
// Implements Value interface.
func (x *Uint64Value) Type() bsttype.Type {
	return bsttype.Uint64()
}

// Kind returns the kind of the value.
// Implements the Value interface.
func (x *Uint64Value) Kind() bsttype.Kind {
	return bsttype.KindUint64
}

// Skip seeks through the input reader to skip the value.
// Implements the Value interface.
func (x *Uint64Value) Skip(s io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipUint64(s)
}

// WriteValue writes the value to the writer.
// Implements the Value interface.
func (x *Uint64Value) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	v, err := x.MarshalValue(o)
	if err != nil {
		return 0, err
	}

	n, err := w.Write(v)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write uint64 value")
	}
	return n, nil
}

// ReadValue reads the value from the reader.
// Implements the Value interface.
func (x *Uint64Value) ReadValue(r io.Reader, options bstio.ValueOptions) (int, error) {
	v, n, err := bstio.ReadUint64(r, options.Descending)
	if err != nil {
		return n, err
	}

	x.Value = v
	return n, nil
}

// UnmarshalValue decodes the value from a binary format.
// The value is expected to be encoded in big endian.
// Implements the Value interface.
func (x *Uint64Value) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	v, err := bstio.ParseUint64(in, o.Descending)
	if err != nil {
		return err
	}
	x.Value = v
	return nil
}

// MarshalValue returns the value in a binary format.
// The value is encoded in big endian encoding.
// Implements the Value interface.
func (x Uint64Value) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	return bstio.MarshalUint64(x.Value, o.Descending), nil
}

// BinarySize returns the size of the binary value.
// Implements the ValueMarshaler interface.
func (x Uint64Value) BinarySize() int {
	return 8
}

// GoValue returns the value in a go format.
func (x Uint64Value) GoValue() interface{} {
	return x.Value
}

// Compile-time check if UintValue implements Value interface.
var _ Value = (*UintValue)(nil)

// UintValue is the type that wraps `uint` value, and provides
// its binary form with various length.
// Implements the ValueMarshaler interface.
type UintValue struct {
	Value uint
}

// NewUintValue returns a new UintValue with the given value.
func NewUintValue(v uint) *UintValue {
	return &UintValue{Value: v}
}

func emptyUintValue(_ bsttype.Type) Value {
	return &UintValue{}
}

// String returns human-readable representation of the value.
func (x UintValue) String() string {
	return fmt.Sprintf("Uint(%d)", x.Value)
}

// Type returns the type of the value.
// Implements Value interface.
func (x *UintValue) Type() bsttype.Type {
	return bsttype.Uint()
}

// Kind returns the kind of the value.
// Implements the Value interface.
func (x *UintValue) Kind() bsttype.Kind {
	return bsttype.KindUint
}

// Skip seeks through the input reader to skip the value.
// Implements the Value interface.
func (x *UintValue) Skip(s io.ReadSeeker, o bstio.ValueOptions) (int64, error) {
	return bstio.SkipUint(s, o.Descending)
}

// WriteValue writes the value to the writer.
// Implements the Value interface.
func (x *UintValue) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	return bstio.WriteUint(w, x.Value, o.Descending)
}

// ReadValue reads the value from the reader.
// Implements the Value interface.
func (x *UintValue) ReadValue(r io.Reader, o bstio.ValueOptions) (int, error) {
	v, n, err := bstio.ReadUint(r, o.Descending)
	if err != nil {
		return n, err
	}

	x.Value = v
	return n, nil
}

// UnmarshalValue decodes the value from a binary format.
// Implements the Value interface.
func (x *UintValue) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	v, _, err := bstio.ReadUint(bytes.NewReader(in), o.Descending)
	if err != nil {
		return err
	}
	x.Value = v
	return nil
}

// MarshalValue returns the value in a binary format.
// The value is encoded in big endian encoding with a header that indicates
// binary length of the value.
// Implements the Value interface.
func (x UintValue) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	return bstio.MarshalUint(x.Value, o.Descending), nil
}

// BinarySize returns the size of the binary value.
// Implements the ValueMarshaler interface.
func (x UintValue) BinarySize() int {
	return bstio.UintBinarySize(x.Value)
}

// GoValue returns the value in a go format.
func (x UintValue) GoValue() interface{} {
	return x.Value
}
