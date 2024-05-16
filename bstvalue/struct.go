package bstvalue

import (
	"bytes"
	"io"
	"strings"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bstskip"
	"github.com/devmodules/bst/bsttype"
)

// Compile-time check if StructValue implements Value interface.
var _ Value = (*StructValue)(nil)

// StructValue is a struct value.
// Before each field a header is stored which contains following information:
//   - Field number (buffIndex of the field in the struct).
//   - Binary size of the value (in bytes).
//   - Binary value of the field.
//     I.e.:
//     Field 0 - Int64 (8 bytes)
//     1 | 8 | 0x01 0x02 0x03 0x04 0x05 0x06 0x07 0x08
type StructValue struct {
	StructType *bsttype.Struct
	Fields     []Value
}

// MustNewStructValue creates a new struct value.
func MustNewStructValue(st *bsttype.Struct, fields []Value) *StructValue {
	v, err := NewStructValue(st, fields)
	if err != nil {
		panic(err)
	}
	return v
}

// NewStructValue creates a new struct value.
func NewStructValue(t *bsttype.Struct, fields []Value) (*StructValue, error) {
	if len(fields) != len(t.Fields) {
		return nil, bsterr.Err(bsterr.CodeMissingFixedSizeValues, "struct value has wrong number of fields")
	}
	for i, f := range fields {
		if !bsttype.TypesEqual(f.Type(), t.Fields[i].Type) {
			return nil, bsterr.Err(bsterr.CodeMismatchingValueType, "struct value has wrong type for field").
				WithDetails(
					bsterr.D("struct", t),
					bsterr.D("typeField", t.Fields[i]),
					bsterr.D("valueField", f.Type()),
				)
		}
	}

	return &StructValue{
		StructType: t,
		Fields:     fields,
	}, nil
}

// EmptyStructValueOf returns an empty struct value of the given type.
func EmptyStructValueOf(t *bsttype.Struct) *StructValue {
	v := &StructValue{StructType: t}
	v.Fields = make([]Value, len(t.Fields))
	for i := range t.Fields {
		v.Fields[i] = EmptyValueOf(t.Fields[i].Type)
	}
	return v
}

// String provides a human-readable representation of the struct value.
// Implements fmt.Stringer and the Value interface.
func (x StructValue) String() string {
	sb := strings.Builder{}
	sb.WriteString("struct ")
	sb.WriteString("{")
	for i, f := range x.Fields {
		if i > len(x.StructType.Fields)-1 {
			break
		}
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(x.StructType.Fields[i].Name)
		sb.WriteString(": ")
		sb.WriteString(f.String())
	}
	sb.WriteRune('}')
	return sb.String()
}

// Type returns the type of the value.
// Implements the Value interface.
func (x *StructValue) Type() bsttype.Type {
	return x.StructType
}

// Kind returns the basic kind of the value.
// Implements the Value interface.
func (x *StructValue) Kind() bsttype.Kind {
	return bsttype.KindStruct
}

// Skip skips the value in the reader.
// Implements the Value interface.
func (x *StructValue) Skip(rs io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
	return bstskip.SkipStruct(rs, x.StructType, options)
}

// MarshalValue marshals the value to the writer.
// Implements the Value interface.
func (x *StructValue) MarshalValue(options bstio.ValueOptions) ([]byte, error) {
	var buf bytes.Buffer
	_, err := x.WriteValue(&buf, options)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalValue unmarshals the value from the reader.
// Implements the Value interface.
func (x *StructValue) UnmarshalValue(in []byte, options bstio.ValueOptions) error {
	_, err := x.ReadValue(bytes.NewReader(in), options)
	return err
}

// ReadValue reads the value from the byte slice.
func (x *StructValue) ReadValue(r io.Reader, options bstio.ValueOptions) (int, error) {
	var (
		bytesRead, n     int
		boolBuf, boolPos byte
		err              error
	)
	desc := options.Descending
	for fi, f := range x.Fields {
		fDesc := desc
		if x.StructType.Fields[fi].Descending {
			fDesc = !desc
		}

		if f.Kind() == bsttype.KindBoolean {
			prev, ok := x.StructType.PreviewPrevElemType(fi)
			if !ok || boolPos == 0 || (ok && prev.Kind() == bsttype.KindBoolean) {
				boolBuf, err = bstio.ReadByte(r)
				if err != nil {
					return bytesRead, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to read bool value")
				}
				bytesRead++
			}
			v := boolBuf&(1<<boolPos) != 0
			if fDesc {
				v = !v
			}
			x.Fields[fi] = NewBoolValue(v)
			boolPos++

			if boolPos == 8 {
				boolPos = 0
				boolBuf = 0x00
			}
			continue
		}

		n, err = f.ReadValue(r, bstio.ValueOptions{Descending: fDesc})
		if err != nil {
			return bytesRead, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read struct field").
				WithDetail("field", x.StructType.Fields[fi].Name)
		}
		bytesRead += n
	}
	return bytesRead, nil
}

// WriteValue writes the value to the byte slice.
func (x *StructValue) WriteValue(w io.Writer, options bstio.ValueOptions) (int, error) {
	var (
		bytesWritten int
		boolBuf      byte
		boolPos      int
	)
	for fi, f := range x.Fields {
		if bv, ok := f.(*BoolValue); ok {
			var bin byte
			if bv.Value {
				bin = 0x01
			}
			boolBuf |= bin << boolPos
			boolPos++
			if boolPos == 8 || !x.isNextBool(fi) {
				if options.Descending {
					boolBuf = ^boolBuf
				}
				_, err := w.Write([]byte{boolBuf})
				if err != nil {
					return bytesWritten, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write struct field").
						WithDetail("field", x.StructType.Fields[fi].Name)
				}
				bytesWritten++
			}
			continue
		}
		n, err := f.WriteValue(w, options)
		if err != nil {
			return bytesWritten, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write struct field").
				WithDetails(bsterr.D("field", x.StructType.Fields[fi].Name))
		}
		bytesWritten += n
	}
	return bytesWritten, nil
}

func (x *StructValue) isNextBool(i int) bool {
	if i+1 < len(x.Fields)-1 {
		return x.Fields[i+1].Kind() == bsttype.KindBoolean
	}
	return false
}
