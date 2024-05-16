package bstvalue

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// Compile-time check that NullableValue implements the Value interface.
var _ Value = (*NullableValue)(nil)

// NullableValue is value type could be nullable.
type NullableValue struct {
	NullableType *bsttype.Nullable
	Value        Value
	IsNull       bool
}

// NullValueOf returns a null value of the given type.
func NullValueOf(t *bsttype.Nullable) *NullableValue {
	ev := EmptyValueOf(t.Type)
	if ev == nil {
		panic(fmt.Sprintf("empty value not found for type %v", t))
	}
	return &NullableValue{IsNull: true, Value: ev, NullableType: t}
}

// EmptyNullableValue returns an empty value of the given type.
func EmptyNullableValue(t bsttype.Type) *NullableValue {
	return &NullableValue{IsNull: true, Value: EmptyValueOf(t), NullableType: bsttype.NullableOf(t)}
}

// MustNullableValue wraps the given value in a NullableValue.
// Panics if the value is nil.
func MustNullableValue(v Value, isNull bool) *NullableValue {
	if v == nil {
		panic("undefined nullable value")
	}
	nt := bsttype.NullableOf(v.Type())
	return &NullableValue{NullableType: nt, Value: v, IsNull: isNull}
}

// NullableValueOf wraps the given value in a NullableValue.
func NullableValueOf(v Value, isNull bool) (*NullableValue, error) {
	if v == nil {
		return nil, bsterr.Err(bsterr.CodeInvalidValue, "undefined nullable value")
	}
	nt := bsttype.NullableOf(v.Type())
	return &NullableValue{NullableType: nt, Value: v, IsNull: isNull}, nil
}

func emptyNullableValue(t bsttype.Type) Value {
	nt := t.(*bsttype.Nullable)

	ev := EmptyValueOf(nt.Elem())
	if ev == nil {
		panic(fmt.Sprintf("empty value not found for type %v", nt.Elem()))
	}
	return &NullableValue{IsNull: true, Value: ev}
}

// Kind returns the kind of the value.
// Implements the Value interface.
func (*NullableValue) Kind() bsttype.Kind {
	return bsttype.KindNullable
}

// String returns a human-readable string representation of the NullableValue.
func (x *NullableValue) String() string {
	if !x.IsNull && x.Value == nil {
		return "UndefinedNullableValue"
	}
	var sb strings.Builder
	sb.WriteString("Nullable(")
	sb.WriteString(x.NullableType.Elem().String())
	sb.WriteString(")")
	if x.IsNull {
		sb.WriteString("{null}")
	} else {
		sb.WriteRune('{')
		sb.WriteString(x.Value.String())
		sb.WriteRune('}')
	}
	return sb.String()
}

// Skip the byte reader and return the number of bytes skipped.
func (x *NullableValue) Skip(rs io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
	// 2. Read the null flag.
	nf, err := bstio.ReadByte(rs)
	if err != nil {
		return 0, bsterr.Err(bsterr.CodeDecodingBinaryValue, "failed to read null flag")
	}

	if options.Descending {
		nf = ^nf
	}

	// 3. If the value is null, skip the value.
	if nf == 0x0 {
		return 1, nil
	}

	// 4. Otherwise, skip the value.
	skipped, err := x.Elem().Skip(rs, options)
	if err != nil {
		return skipped + 1, bsterr.Err(bsterr.CodeDecodingBinaryValue, "failed to skip value")
	}
	return skipped + 1, nil
}

// MarshalValue marshals the value to the database.
// Implements the Value interface.
func (x *NullableValue) MarshalValue(o bstio.ValueOptions) ([]byte, error) {
	var buf bytes.Buffer
	_, err := x.WriteValue(&buf, o)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalValue unmarshals the value from the database.
func (x *NullableValue) UnmarshalValue(in []byte, options bstio.ValueOptions) error {
	_, err := x.ReadValue(bytes.NewReader(in), options)
	if err != nil {
		return err
	}
	return nil
}

// ReadValue reads the value from the byte reader.
func (x *NullableValue) ReadValue(r io.Reader, o bstio.ValueOptions) (int, error) {
	nf, err := bstio.ReadNullableFlag(r, o.Descending)
	if err != nil {
		return 0, err
	}

	switch nf {
	case 0x1:
		// A 1-bit indicates that the value is not-null.
		x.IsNull = false
		var n int
		n, err = x.Value.ReadValue(r, o)
		if err != nil {
			return n + 1, err
		}
		return n + 1, nil
	case 0x0:
		x.IsNull = true
		// A 0-bit indicates that the value is null.
		return 1, nil
	default:
		return 0, bsterr.Err(bsterr.CodeDecodingBinaryValue, "invalid nullable flag byte").
			WithDetail("value", fmt.Sprintf("%#x", nf))
	}
}

// WriteValue writes the value to the byte writer.
// Implements the Value interface.
func (x *NullableValue) WriteValue(w io.Writer, o bstio.ValueOptions) (int, error) {
	if x.IsNull {
		bt := bstio.NullableIsNull
		if o.Descending {
			bt = bstio.NullableIsNullDesc
		}
		n, err := w.Write([]byte{bt})
		if err != nil {
			return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write nullable flag byte")
		}
		return n, nil
	}

	bt := bstio.NullableIsNotNull
	if o.Descending {
		bt = bstio.NullableIsNotNullDesc
	}
	total, err := w.Write([]byte{bt})
	if err != nil {
		return total, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write nullable flag byte")
	}

	n, err := x.Value.WriteValue(w, o)
	if err != nil {
		return total + n, err
	}
	return total + n, nil
}

// Type returns the type of the value.
func (x *NullableValue) Type() bsttype.Type {
	return &bsttype.Nullable{Type: x.Elem().Type()}
}

// Elem dereferences the pointer wrapped Value.
func (x *NullableValue) Elem() Value {
	return x.Value
}
