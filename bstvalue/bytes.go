package bstvalue

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
	"github.com/devmodules/bst/internal/iopool"
)

// Compile-time check to ensure that Bytes implements the Value interface.
var _ Value = (*Bytes)(nil)

// Bytes is the value descriptor for the []byte.
type Bytes struct {
	BytesType *bsttype.Bytes
	Value     []byte
}

// MustNewBytes creates a new Bytes.
func MustNewBytes(b []byte, bt *bsttype.Bytes) *Bytes {
	if bt.FixedSize > 0 && len(b) != bt.FixedSize {
		panic(fmt.Sprintf("invalid bytes value, expected %d bytes, got %d", bt.FixedSize, len(b)))
	}
	return &Bytes{
		BytesType: bt,
		Value:     b,
	}
}

// NewBytes creates a new Bytes.
func NewBytes(b []byte, bt *bsttype.Bytes) (*Bytes, error) {
	if bt.FixedSize > 0 && len(b) != bt.FixedSize {
		return nil, bsterr.Err(bsterr.CodeTypeConstraintViolation, "fixed size bytes value constraint violated").
			WithDetails(
				bsterr.D("expected_size", bt.FixedSize),
				bsterr.D("actual_size", len(b)),
			)
	}
	return &Bytes{
		BytesType: bt,
		Value:     b,
	}, nil
}

// EmptyBytes creates a new empty Bytes for input BytesType.
func EmptyBytes(bt *bsttype.Bytes) *Bytes {
	bv := &Bytes{BytesType: bt}
	if bt.FixedSize > 0 {
		bv.Value = make([]byte, bt.FixedSize)
	}
	return bv
}

func emptyBytesValue(t bsttype.Type) Value {
	bt := t.(*bsttype.Bytes)
	return EmptyBytes(bt)
}

// Type returns the type of the value.
// Implements the Value interface.
func (x *Bytes) Type() bsttype.Type {
	return x.BytesType
}

// String returns a human-readable representation of the Bytes.
// Implements the Value interface.
func (x Bytes) String() string {
	var sb strings.Builder
	sb.WriteString(x.BytesType.String())
	sb.WriteString("{")
	for i, b := range x.Value {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%02x", b))
	}
	sb.WriteString("}")
	return sb.String()
}

// Kind returns the basic kind of the value.
// Implements the Value interface.
func (x *Bytes) Kind() bsttype.Kind {
	return bsttype.KindBytes
}

// UnmarshalValue reads the value from the byte slice.
// Implements the Value interface.
func (x *Bytes) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	v, _, err := bstio.ReadBytes(bytes.NewReader(in), x.BytesType.FixedSize, o.Descending, o.Comparable)
	if err != nil {
		return err
	}

	x.Value = v
	return nil
}

// ReadValue reads the value from the byte slice.
// Implements the Value interface.
func (x *Bytes) ReadValue(r io.Reader, o bstio.ValueOptions) (int, error) {
	bt, n, err := bstio.ReadBytes(r, x.BytesType.FixedSize, o.Descending, o.Comparable)
	if err != nil {
		return n, err
	}
	x.Value = bt

	return n, nil
}

// WriteValue writes the value to the byte slice.
// Implements the Value interface.
func (x *Bytes) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	return bstio.WriteBytes(w, x.BytesType.FixedSize, x.Value, o.Descending, o.Comparable)
}

// MarshalValue writes the value to the byte slice.
// Implements the Value interface.
func (x *Bytes) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	buf := iopool.GetBuffer(nil)
	defer iopool.ReleaseBuffer(buf)

	_, err := bstio.WriteBytes(buf, x.BytesType.FixedSize, x.Value, o.Descending, o.Comparable)
	if err != nil {
		return nil, err
	}
	cp := buf.BytesCopy()
	return cp, nil
}

// Skip the bytes in the reader to the next value.
// Implements the Value interface.
func (x *Bytes) Skip(rs io.ReadSeeker, o bstio.ValueOptions) (int64, error) {
	return bstio.SkipBytes(rs, x.BytesType.FixedSize, o.Descending, o.Comparable)
}
