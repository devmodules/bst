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

// Compile time check if OneOfValue implements Value interface.
var _ Value = (*OneOfValue)(nil)

// OneOfValue is a value that can be marshaled into selected list of types.
// Implements Value interface.
type OneOfValue struct {
	Value     Value
	OneOfType *bsttype.OneOf
	Index     uint
}

func emptyOneOfValue(t bsttype.Type) Value {
	return &OneOfValue{OneOfType: t.(*bsttype.OneOf)}
}

// Elem returns the element of the value.
// Implements ElementValue interface.
func (x *OneOfValue) Elem() Value {
	return x.Value
}

// String returns a human-readable representation of the OneOfValue.
// Example: "OneOf.Name(Index: 1, Value: Int(8))"
func (x *OneOfValue) String() string {
	return fmt.Sprintf("OneOf(Index: %d, Value: %s)", x.Index, x.Value)
}

// IndexName returns the name of the element type represented by the buffIndex value.
func (x *OneOfValue) IndexName() (string, bool) {
	for _, e := range x.OneOfType.Elements {
		if e.Index == x.Index {
			return e.Name, true
		}
	}
	return "", false
}

// IndexType returns the type of the element type represented by the buffIndex value.
func (x *OneOfValue) IndexType() (bsttype.Type, bool) {
	for _, e := range x.OneOfType.Elements {
		if e.Index == x.Index {
			return e.Type, true
		}
	}
	return nil, false
}

// MustNewOneOfValue creates a new OneOfValue. If an error occurs, it panics.
func MustNewOneOfValue(oneOfType *bsttype.OneOf, v Value, index uint) *OneOfValue {
	ov, err := NewOneOfValue(oneOfType, v, index)
	if err != nil {
		panic(err)
	}
	return ov
}

// NewOneOfValue creates a new OneOfValue.
func NewOneOfValue(oneOfType *bsttype.OneOf, v Value, index uint) (*OneOfValue, error) {
	// Find the element of the input type.
	idx := -1
	for i := range oneOfType.Elements {
		if oneOfType.Elements[i].Index == index {
			idx = i
			break
		}
	}

	// If the buffIndex is not valid, return an error.
	if idx == -1 {
		return nil, bsterr.Err(bsterr.CodeTypeConstraintViolation, "oneOfType buffIndex is not found")
	}

	if !bsttype.TypesEqual(v.Type(), oneOfType.Elements[idx].Type) {
		return nil, bsterr.Err(bsterr.CodeTypeConstraintViolation, "the value type does not match the oneOfType type").
			WithDetails(
				bsterr.D("value", v.Type()),
				bsterr.D("expected", oneOfType.Elements[idx].Type),
			)
	}

	return &OneOfValue{
		Value:     v,
		OneOfType: oneOfType,
		Index:     index,
	}, nil
}

// Type returns the type of the value.
func (x *OneOfValue) Type() bsttype.Type {
	return x.OneOfType
}

// Kind returns the kind of the value.
// Implements Value interface.
func (x *OneOfValue) Kind() bsttype.Kind {
	return bsttype.KindOneOf
}

// Skip passes through the reader to the end of OneOfValue.
// It first needs to decode corresponding value buffIndex, to know which type to skip.
// Then it initializes the value for given element type and skips it.
// Implements Value interface.
func (x *OneOfValue) Skip(rs io.ReadSeeker, o bstio.ValueOptions) (int64, error) {
	return bstskip.SkipOneOf(rs, x.OneOfType, o)
}

// MarshalValue marshals the value into the writer.
func (x *OneOfValue) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	var buf bytes.Buffer
	_, err := x.WriteValue(&buf, o)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalValue unmarshals the value from the reader.
// Implements Value interface.
func (x *OneOfValue) UnmarshalValue(in []byte, o bstio.ValueOptions) error {
	_, err := x.ReadValue(bytes.NewReader(in), o)
	return err
}

// ReadValue reads the value from the reader.
// Implements Value interface.
func (x *OneOfValue) ReadValue(br io.Reader, o bstio.ValueOptions) (int, error) {
	// 1. Read the buffIndex.
	idx, bytesRead, err := bstio.ReadOneOfIndex(br, x.OneOfType.IndexBytes, o.Descending)
	if err != nil {
		return bytesRead, err
	}

	// 2. Match the buffIndex to the element.
	var elem bsttype.Type
	for i := range x.OneOfType.Elements {
		if x.OneOfType.Elements[i].Index == idx {
			elem = x.OneOfType.Elements[i].Type
			break
		}
	}

	// 3. If the buffIndex did not match, return an error.
	if elem == nil {
		return bytesRead, bsterr.Err(bsterr.CodeTypeConstraintViolation, "oneOfType buffIndex doesn't match to the elements")
	}
	x.Index = idx

	// 4. Initialize the value for given type.
	x.Value = EmptyValueOf(elem)

	// 5. Read the value.
	bytesRead, err = x.Value.ReadValue(br, o)
	if err != nil {
		return bytesRead, err
	}

	return bytesRead, nil
}

// WriteValue writes the value to the writer.
func (x *OneOfValue) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	// 1. Write the value buffIndex.
	ot := x.OneOfType
	index := x.Index
	descending := o.Descending
	n, err := bstio.WriteOneOfIndex(w, index, ot.IndexBytes, descending)
	if err != nil {
		return n, err
	}
	bytesWritten := n

	// 2. Write the value.
	n, err = x.Value.WriteValue(w, o)
	if err != nil {
		return bytesWritten + n, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write oneOfType value")
	}
	return bytesWritten + n, nil
}
