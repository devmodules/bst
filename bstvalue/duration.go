package bstvalue

import (
	"fmt"
	"io"
	"time"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// Compile-time check to ensure that DurationValue implements the Value interface.
var _ Value = (*DurationValue)(nil)

// DurationValue is the value descriptor for the time.Duration.
type DurationValue struct {
	Value time.Duration
}

// NewDurationValue returns a new DurationValue.
func NewDurationValue(v time.Duration) *DurationValue {
	return &DurationValue{Value: v}
}

func emptyDurationValue(_ bsttype.Type) Value {
	return &DurationValue{}
}

// String returns a human-readable description of the DurationValue.
func (x DurationValue) String() string {
	return fmt.Sprintf("Duration(%s)", x.Value)
}

// Type returns the type of the value.
// Implements the Value interface.
func (*DurationValue) Type() bsttype.Type {
	return bsttype.Duration()
}

// Kind returns the basic kind of the value.
// Implements the Value interface.
func (*DurationValue) Kind() bsttype.Kind {
	return bsttype.KindDuration
}

// Skip the bytes in the reader to the next value.
// Implements the Value interface.
func (*DurationValue) Skip(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipInt64(rs)
}

// MarshalValue writes the value to the byte slice.
// Implements the Value interface.
func (x *DurationValue) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	return bstio.MarshalInt64(x.Value.Nanoseconds(), o.Descending), nil
}

// UnmarshalValue reads the value from the byte slice.
// Implements the Value interface.
func (x *DurationValue) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	v, err := bstio.ParseInt64(in, o.Descending)
	if err != nil {
		return err
	}

	x.Value = time.Duration(v)
	return nil
}

// ReadValue reads the value from the byte slice.
// Implements the Value interface.
func (x *DurationValue) ReadValue(r io.Reader, o bstio.ValueOptions) (int, error) {
	v, n, err := bstio.ReadInt64(r, o.Descending)
	if err != nil {
		return n, err
	}

	x.Value = time.Duration(v)
	return n, nil
}

// WriteValue writes the value to the byte slice.
// Implements the Value interface.
func (x *DurationValue) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	bt := bstio.MarshalInt64(x.Value.Nanoseconds(), o.Descending)
	n, err := w.Write(bt)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write duration value")
	}

	return n, nil
}
