package bstvalue

import (
	"fmt"
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// Compile-time check to ensure that Float32Value implements the Value interface.
var _ Value = (*Float32Value)(nil)

// Float32Value is the value descriptor for the float32.
// The IEEE-754 standard specifies single-precision 32-bit floating point numbers
// as follows:
//
//			sign bit: 1 bit
//	     exponent bits: 8 bits
//	     significand bits: 23 bits
//	     total: 32 bits
//	 However this binary representation does not provide raw-bytes comparison
//	 for negative and positive values.
//	 To overcome this limitation, the first bit of the first byte
//	 is marked in opposite way for negative and positive values.
//	 In example in the IEEE-754 the first bit of the first byte is set to 1 for negative values
//	 and 0 for positive values. In this encoding the first bit of the first byte is set to 1 for positive
//	 values and 0 for negative values.
//	 Zero is represented as positive value.
type Float32Value struct {
	Value float32
}

// NewFloat32Value creates a new Float32Value.
func NewFloat32Value(in float32) *Float32Value {
	return &Float32Value{Value: in}
}

func emptyFloat32Value(_ bsttype.Type) Value {
	return &Float32Value{}
}

// Type returns the type of the value.
// Implements the Value interface.
func (*Float32Value) Type() bsttype.Type {
	return bsttype.Float32()
}

// String returns a human-readable description of the Float32Value.
func (x Float32Value) String() string {
	return fmt.Sprintf("Float32(%v)", x.Value)
}

// Kind returns the basic kind of the value.
// Implements the Value interface.
func (x *Float32Value) Kind() bsttype.Kind {
	return bsttype.KindFloat32
}

// Skip the bytes in the reader to the next value.
// Implements the Value interface.
func (x *Float32Value) Skip(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipFloat32(rs)
}

// MarshalValue writes the value to the byte slice.
// Implements the Value interface.
func (x *Float32Value) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	return bstio.MarshalFloat32(x.Value, o.Descending), nil
}

// UnmarshalValue reads the value from the byte slice.
// Implements the Value interface.
func (x *Float32Value) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	if len(in) != 4 {
		return bsterr.Err(bsterr.CodeDecodingBinaryValue, "failed to unmarshal float value: invalid length").
			WithDetails(
				bsterr.D("length", len(in)),
				bsterr.D("expected", 4),
			)
	}

	fv, err := bstio.ParseFloat32(in, o.Descending)
	if err != nil {
		return err
	}
	x.Value = fv
	return nil
}

// ReadValue reads the value from the byte slice.
// Implements the Value interface.
func (x *Float32Value) ReadValue(r io.Reader, o bstio.ValueOptions) (int, error) {
	v, n, err := bstio.ReadFloat32(r, o.Descending)
	if err != nil {
		return 0, err
	}
	x.Value = v
	return n, nil
}

// WriteValue writes the value to the byte slice.
// Implements the Value interface.
func (x *Float32Value) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	v := bstio.MarshalFloat32(x.Value, o.Descending)
	n, err := w.Write(v)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write float value")
	}
	return n, nil
}

// Compile-time check to ensure that Float64Value implements the Value interface.
var _ Value = (*Float64Value)(nil)

// Float64Value is the value descriptor for the float64.
type Float64Value struct {
	Value float64
}

// NewFloat64Value creates a new Float64Value.
func NewFloat64Value(in float64) *Float64Value {
	return &Float64Value{Value: in}
}

func emptyFloat64Value(_ bsttype.Type) Value {
	return &Float64Value{}
}

// Type returns the type of the value.
// Implements the Value interface.
func (*Float64Value) Type() bsttype.Type {
	return bsttype.Float64()
}

// Kind returns the basic kind of the value.
// Implements the Value interface.
func (*Float64Value) Kind() bsttype.Kind {
	return bsttype.KindFloat64
}

// String returns a human-readable description of the Float64Value.
func (x Float64Value) String() string {
	return fmt.Sprintf("Float64(%v)", x.Value)
}

// Skip the bytes in the reader to the next value.
// Implements the Value interface.
func (x *Float64Value) Skip(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipFloat64(rs)
}

// MarshalValue writes the value to the byte slice.
// Implements the Value interface.
func (x *Float64Value) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	return bstio.MarshalFloat64(x.Value, o.Descending), nil
}

// UnmarshalValue reads the value from the byte slice.
// Implements the Value interface.
func (x *Float64Value) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	if len(in) != 8 {
		return bsterr.Err(bsterr.CodeDecodingBinaryValue, "failed to unmarshal float value: invalid length").
			WithDetails(
				bsterr.D("length", len(in)),
				bsterr.D("expected", 8),
			)
	}

	fv, err := bstio.ParseFloat64(in, o.Descending)
	if err != nil {
		return err
	}
	x.Value = fv
	return nil
}

// ReadValue reads the value from the byte slice.
// Implements the Value interface.
func (x *Float64Value) ReadValue(r io.Reader, o bstio.ValueOptions) (int, error) {
	v, n, err := bstio.ReadFloat64(r, o.Descending)
	if err != nil {
		return n, err
	}
	x.Value = v
	return n, nil
}

// WriteValue writes the value to the byte slice.
// Implements the Value interface.
func (x *Float64Value) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	v := bstio.MarshalFloat64(x.Value, o.Descending)
	n, err := w.Write(v)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write float value")
	}
	return n, nil
}
