package bstvalue

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

var _ Value = (*DateTime)(nil)

// DateTime is the value for the DateTime.
// Implements the Value interface.
type DateTime struct {
	DateTimeType *bsttype.DateTime
	Value        time.Time
}

// NewDateTimeValue creates a new DateTime.
func NewDateTimeValue(dt *bsttype.DateTime, v time.Time) *DateTime {
	return &DateTime{
		DateTimeType: dt,
		Value:        v,
	}
}

// EmptyDateTimeValue returns an empty DateTime.
func EmptyDateTimeValue(t *bsttype.DateTime) *DateTime {
	return &DateTime{DateTimeType: t}
}

func emptyDateTimeValue(t bsttype.Type) Value {
	return &DateTime{DateTimeType: t.(*bsttype.DateTime)}
}

// String returns a human-readable description of the DateTime.
func (x *DateTime) String() string {
	return fmt.Sprintf("%s (%s)", x.DateTimeType.String(), x.Value.Format(time.RFC3339))
}

// Type returns the type of the value.
// Implements the Value interface.
func (x *DateTime) Type() bsttype.Type {
	return x.DateTimeType
}

// Kind returns the basic kind of the value.
// Implements the Value interface.
func (x *DateTime) Kind() bsttype.Kind {
	return bsttype.KindDateTime
}

// Skip the value in the reader.
// Implements the Value interface.
func (x *DateTime) Skip(rs io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
	return bstio.SkipDateTime(rs, options.Descending)
}

// MarshalValue marshals the value to the writer.
// Implements the Value interface.
func (x *DateTime) MarshalValue(options bstio.ValueOptions) ([]byte, error) {
	var buf bytes.Buffer
	_, err := x.WriteValue(&buf, options)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalValue unmarshals the value from the reader.
func (x *DateTime) UnmarshalValue(in []byte, options bstio.ValueOptions) error {
	br := bytes.NewReader(in)
	_, err := x.ReadValue(br, options)
	return err
}

// ReadValue reads the value from the byte reader.
// Implements the Value interface.
func (x *DateTime) ReadValue(r io.Reader, options bstio.ValueOptions) (int, error) {
	dt := x.DateTimeType
	tm, bytesRead, err := bstio.ReadDateTime(r, options.Descending, dt.Location())
	if err != nil {
		return bytesRead, err
	}
	x.Value = tm
	return bytesRead, nil
}

// WriteValue writes the value to the writer.
// Implements the Value interface.
func (x *DateTime) WriteValue(w io.Writer, options bstio.ValueOptions) (int, error) {
	return bstio.WriteDateTime(w, x.Value, options.Descending, x.DateTimeType.Location())
}
