package bstvalue

import (
	"fmt"
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// Compile-time check to ensure that BoolValue implements the Value interface.
var _ Value = (*BoolValue)(nil)

// BoolValue returns the value as a bool.
// Implements the Value interface.
type BoolValue struct {
	Value bool
}

// NewBoolValue creates a new bool value.
func NewBoolValue(v bool) *BoolValue {
	return &BoolValue{Value: v}
}

func emptyBoolValue(_ bsttype.Type) Value {
	return &BoolValue{}
}

// Type returns the type of the value.
// Implements the Value interface.
func (b *BoolValue) Type() bsttype.Type {
	return bsttype.Boolean()
}

// String returns a human-readable representation of the BoolValue.
func (b BoolValue) String() string {
	return fmt.Sprintf("Bool(%v)", b.Value)
}

// Kind returns the basic kind of the value.
// Implements the Value interface.
func (b BoolValue) Kind() bsttype.Kind {
	return bsttype.KindBoolean
}

// Skip the bytes in the reader to the next value.
// Implements the Value interface.
func (b BoolValue) Skip(br io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipBool(br)
}

// MarshalValue writes the value to the byte slice.
// Implements the Value interface.
func (b *BoolValue) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	bt := b.binaryValue(o)
	return []byte{bt}, nil
}

// UnmarshalValue reads the value from the byte slice.
// Implements the Value interface.
func (b *BoolValue) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	if len(in) != 1 {
		return bsterr.Err(bsterr.CodeDecodingBinaryValue, "invalid bool binary value length").
			WithDetails(
				bsterr.D("length", len(in)),
				bsterr.D("expected", 1),
			)
	}

	bv, err := bstio.ParseBool(in[0], o.Descending)
	if err != nil {
		return err
	}
	b.Value = bv

	return nil
}

// ReadValue reads the value from the reader.
// Implements the Value interface.
func (b *BoolValue) ReadValue(br io.Reader, o bstio.ValueOptions) (int, error) {
	bv, n, err := bstio.ReadBool(br, o.Descending)
	if err != nil {
		return n, err
	}
	b.Value = bv
	return n, nil
}

// WriteValue writes the value to the writer.
// Implements the Value interface.
func (b *BoolValue) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	bt := b.binaryValue(o)
	n, err := w.Write([]byte{bt})
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write bool value")
	}
	return n, err
}

func (b *BoolValue) binaryValue(o bstio.ValueOptions) byte {
	var bt byte
	switch {
	case b.Value && o.Descending:
		bt = bstio.BoolTrueDesc
	case b.Value && !o.Descending:
		bt = bstio.BoolTrue
	case !b.Value && o.Descending:
		bt = bstio.BoolFalseDesc
	case !b.Value && !o.Descending:
		bt = bstio.BoolFalse
	}
	return bt
}
