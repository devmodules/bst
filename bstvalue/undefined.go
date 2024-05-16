package bstvalue

import (
	"io"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

var _ Value = (*UndefinedValue)(nil)

// UndefinedValue is the value of the undefined type.
// It is used as a placeholder for the undefined type value.
// Implements the Value interface.
type UndefinedValue struct{}

func emptyUndefinedValue(_ bsttype.Type) Value {
	return UndefinedValue{}
}

// Type returns the type of the value.
// Implements the Value interface.
func (UndefinedValue) Type() bsttype.Type {
	return &bsttype.Basic{}
}

// String returns a human-readable description of the UndefinedValue.
func (u UndefinedValue) String() string {
	return "Undefined"
}

// Kind returns the basic kind of the value.
// Implements the Value interface.
func (u UndefinedValue) Kind() bsttype.Kind {
	return bsttype.KindUndefined
}

// Skip panics because the undefined type is not supported.
// Implements the Value interface.
func (u UndefinedValue) Skip(_ io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return 0, nil
}

// MarshalValue panics because the undefined type is not supported.
// Implements the Marshaler interface.
func (u UndefinedValue) MarshalValue(_ bstio.ValueOptions) ([]byte, error) {
	return []byte{}, nil
}

// UnmarshalValue panics because the undefined type is not supported.
// Implements the Unmarshaler interface.
func (u UndefinedValue) UnmarshalValue(_ []byte, _ bstio.ValueOptions) error {
	return nil
}

// ReadValue panics because the undefined type is not supported.
// Implements the ValueReader interface.
func (u UndefinedValue) ReadValue(_ io.Reader, _ bstio.ValueOptions) (int, error) {
	return 0, nil
}

// WriteValue panics because the undefined type is not supported.
// Implements the ValueWriter interface.
func (u UndefinedValue) WriteValue(_ io.Writer, _ bstio.ValueOptions) (int, error) {
	return 0, nil
}
