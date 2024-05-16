package bstvalue

import (
	"io"
	"strings"
	"time"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// Compile-time check to ensure that TimestampValue implements the Value interface.
var _ Value = (*TimestampValue)(nil)

// TimestampValue is the value descriptor for the time.Time.
type TimestampValue struct {
	Value time.Time
}

// NewTimestampValue creates a new TimestampValue.
func NewTimestampValue(v time.Time) *TimestampValue {
	return &TimestampValue{Value: v}
}

func emptyTimestampValue(_ bsttype.Type) Value {
	return &TimestampValue{}
}

// Type returns the type of the value.
// Implements the Value interface.
func (*TimestampValue) Type() bsttype.Type {
	return bsttype.Timestamp()
}

// String returns a string representation of the value.
func (x *TimestampValue) String() string {
	var sb strings.Builder
	sb.WriteString("Timestamp(")
	sb.WriteString(x.Value.UTC().Format(time.RFC3339Nano))
	sb.WriteRune(')')
	return sb.String()
}

// Kind returns the basic kind of the value.
// Implements the Value interface.
func (x *TimestampValue) Kind() bsttype.Kind {
	return bsttype.KindTimestamp
}

// Skip the bytes in the reader to the next value.
// Implements the Value interface.
func (x *TimestampValue) Skip(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipInt64(rs)
}

// MarshalValue writes the value to the byte slice.
// Implements the Value interface.
func (x *TimestampValue) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	return bstio.MarshalInt64(x.Value.UTC().UnixNano(), o.Descending), nil
}

// UnmarshalValue reads the value from the byte slice.
// Implements the Value interface.
func (x *TimestampValue) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	v, err := bstio.ParseInt64(in, o.Descending)
	if err != nil {
		return err
	}

	x.Value = time.Unix(0, v).UTC()

	return nil
}

// ReadValue reads the value from the byte slice.
// Implements the Value interface.
func (x *TimestampValue) ReadValue(r io.Reader, o bstio.ValueOptions) (int, error) {
	v, n, err := bstio.ReadInt64(r, o.Descending)
	if err != nil {
		return n, err
	}

	x.Value = time.Unix(0, v).UTC()
	return n, nil
}

// WriteValue writes the value to the byte slice.
// Implements the Value interface.
func (x *TimestampValue) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	m := bstio.MarshalInt64(x.Value.UTC().UnixNano(), o.Descending)
	n, err := w.Write(m)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write date time value")
	}

	return n, nil
}
