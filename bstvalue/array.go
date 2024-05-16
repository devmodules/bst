package bstvalue

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bstskip"
	"github.com/devmodules/bst/bsttype"
)

// Compile-time check that ArrayValue implements the Value interface.
var _ Value = (*ArrayValue)(nil)

// ArrayValue is a value of the array type.
type ArrayValue struct {
	ArrayType *bsttype.Array
	Values    []Value
}

// MustArrayValueOf returns an array value of the given type and values.
func MustArrayValueOf(at *bsttype.Array, values []Value) *ArrayValue {
	av, err := ArrayValueOf(at, values)
	if err != nil {
		panic(err)
	}
	return av
}

// ArrayValueOf returns a new array value.
func ArrayValueOf(at *bsttype.Array, values []Value) (*ArrayValue, error) {
	av := &ArrayValue{ArrayType: at, Values: values}

	if at.HasFixedSize() {
		if len(values) != int(at.FixedSize) {
			return nil, bsterr.Errf(bsterr.CodeMissingFixedSizeValues, "array size mismatch")
		}
	}

	for _, v := range values {
		if !bsttype.TypesEqual(v.Type(), at.Type) {
			return nil, bsterr.Errf(bsterr.CodeMismatchingValueType, "array value type mismatch")
		}
	}
	return av, nil
}

// EmptyArrayValue returns a new empty array value.
func EmptyArrayValue(at *bsttype.Array) *ArrayValue {
	av := &ArrayValue{
		ArrayType: at,
	}
	if at.HasFixedSize() {
		av.Values = make([]Value, at.FixedSize)
	}
	return av
}

func emptyArrayValue(t bsttype.Type) Value {
	at := t.(*bsttype.Array)

	av := &ArrayValue{ArrayType: at}
	if at.HasFixedSize() {
		av.Values = make([]Value, at.FixedSize)
	}
	return av
}

// Kind returns the kind of the value.
func (x *ArrayValue) Kind() bsttype.Kind {
	return bsttype.KindArray
}

// Type returns the type of the value.
func (x *ArrayValue) Type() bsttype.Type {
	return x.ArrayType
}

// String returns a human-readable representation of the ArrayValue.
func (x *ArrayValue) String() string {
	var b strings.Builder
	b.WriteString(x.ArrayType.String())
	b.WriteString("[")
	for i, v := range x.Values {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(v.String())
	}
	b.WriteString("]")
	return b.String()
}

// Len returns the length of the array.
func (x *ArrayValue) Len() int {
	return len(x.Values)
}

// MarshalValue marshals the value to the database format.
// Implements the Value interface.
func (x *ArrayValue) MarshalValue(options bstio.ValueOptions) ([]byte, error) {
	buf := &bytes.Buffer{}
	_, err := x.WriteValue(buf, options)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Skip the bytes in the reader to the next value.
// Implements the Value interface.
func (x *ArrayValue) Skip(br io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
	return bstskip.SkipArray(br, x.ArrayType, options)
}

// UnmarshalValue unmarshals the value from the database format.
// Implements the Value interface.
func (x *ArrayValue) UnmarshalValue(data []byte, options bstio.ValueOptions) error {
	r := bytes.NewReader(data)
	if x.ArrayType.Type.Kind() == bsttype.KindBoolean {
		_, err := x.readBools(r, options)
		return err
	}
	_, err := x.readArray(r, options)
	return err
}

// ReadValue reads the value from the reader.
// Implements the Value interface.
func (x *ArrayValue) ReadValue(br io.Reader, options bstio.ValueOptions) (int, error) {
	if x.ArrayType.Type.Kind() == bsttype.KindBoolean {
		return x.readBools(br, options)
	}
	return x.readArray(br, options)
}

// WriteValue writes the value to the writer.
// Implements the Value interface.
func (x *ArrayValue) WriteValue(w io.Writer, options bstio.ValueOptions) (int, error) {
	if x.ArrayType.Type.Kind() == bsttype.KindBoolean {
		return x.writeBools(w, options)
	}
	return x.write(w, options)
}

// NthElem returns the nth element of the array.
// It panics if the buffIndex is out of range.
// Implements the Array interface.
func (x *ArrayValue) NthElem(n int) Value {
	if len(x.Values) < n-1 {
		panic("elem buffIndex out of range")
	}
	return x.Values[n]
}

// DropNthElem drops the nth element of the array.
// It panics if the buffIndex is out of range.
// If the array is fixed size, the nth element is changed to undefined.
// If the array is not fixed size, the nth element is removed, and whole slice is being shifted.
// Implements the Array interface.
func (x *ArrayValue) DropNthElem(n int) {
	if len(x.Values) < n-1 {
		panic("elem buffIndex out of range")
	}
	if x.ArrayType.HasFixedSize() {
		// If the array has a fixed size, the nth element needs to be cleared.
		// TODO(kucjac): implement undefined value.
		return
	}
	x.Values = append(x.Values[:n], x.Values[n+1:]...)
}

// Append appends the value to the array.
func (x *ArrayValue) Append(sv Value) error {
	// 1. Check if the array is fixed size.
	if x.ArrayType.HasFixedSize() {
		return bsterr.Err(bsterr.CodeTypeConstraintViolation, "cannot append to fixed array type")
	}

	// 2. Check if the value is of the same type.
	x.Values = append(x.Values, sv)
	return nil
}

func (x *ArrayValue) readArray(br io.Reader, options bstio.ValueOptions) (int, error) {
	// 1. Parse the length.
	length := int(x.ArrayType.FixedSize)
	var bytesRead int
	if !x.ArrayType.HasFixedSize() {
		// 2. Read the length for variable length array.
		luv, n, err := bstio.ReadUint(br, options.Descending)
		if err != nil {
			return n, err
		}
		bytesRead += n
		// 3. Allocate the array values.
		//	  If the array has a fixed size, the length is already allocated.
		x.Values = make([]Value, luv)
		length = int(luv)
	}

	// 2. If the length is 0, return empty slice.
	if length == 0 {
		return bytesRead, nil
	}

	// 3. Read the elements.
	for i := 0; i < length; i++ {
		ev := EmptyValueOf(x.ArrayType.Elem())
		if ev == nil {
			panic(fmt.Sprintf("unsupported array element type %v", x.ArrayType.Elem()))
		}
		nt, err := ev.ReadValue(br, options)
		if err != nil {
			return bytesRead + nt, err
		}
		x.Values[i] = ev
		bytesRead += nt
	}
	return bytesRead, nil
}

func (x *ArrayValue) readBools(br io.Reader, options bstio.ValueOptions) (int, error) {
	// 1. Parse the length.
	length := int(x.ArrayType.FixedSize)
	var bytesRead int
	if !x.ArrayType.HasFixedSize() {
		// 2. Read the length for variable length array.
		luv, n, err := bstio.ReadUint(br, options.Descending)
		if err != nil {
			return n, err
		}
		bytesRead += n

		// 3. Allocate the array values.
		//	  If the array has a fixed size, the length is already allocated.
		x.Values = make([]Value, luv)
		length = int(luv)
	}

	// 2. If the length is 0, return empty slice.
	if length == 0 {
		return bytesRead, nil
	}

	// 3. Read the elements.
	var boolBuf, boolPos byte
	for i := 0; i < length; i++ {
		if boolPos == 0 {
			var err error
			boolBuf, err = bstio.ReadByte(br)
			if err != nil {
				return bytesRead, err
			}
			bytesRead++
		}
		bv := boolBuf&(1<<boolPos) != 0
		x.Values[i] = &BoolValue{Value: bv}
		boolPos++
	}
	return bytesRead, nil
}

func (x *ArrayValue) write(w io.Writer, options bstio.ValueOptions) (int, error) {
	var (
		bytesWritten int
		err          error
	)
	if !x.ArrayType.HasFixedSize() {
		bytesWritten, err = bstio.WriteUint(w, uint(len(x.Values)), options.Descending)
		if err != nil {
			return bytesWritten, err
		}
	}
	var vw int
	for i, v := range x.Values {
		if v == nil {
			v = EmptyValueOf(x.ArrayType.Elem())
			x.Values[i] = v
		}
		vw, err = v.WriteValue(w, options)
		if err != nil {
			return bytesWritten + vw, err
		}
		bytesWritten += vw
	}
	return bytesWritten, nil
}

func (x *ArrayValue) writeBools(w io.Writer, options bstio.ValueOptions) (int, error) {
	var (
		bytesWritten     int
		err              error
		boolBuf, boolPos byte
	)

	if !x.ArrayType.HasFixedSize() {
		bytesWritten, err = bstio.WriteUint(w, uint(len(x.Values)), options.Descending)
		if err != nil {
			return bytesWritten, err
		}
	}

	for i, v := range x.Values {
		bv, ok := v.(*BoolValue)
		if !ok {
			return bytesWritten, bsterr.Err(bsterr.CodeTypeConstraintViolation, "array element is not a bool")
		}
		var bin byte
		if bv.Value {
			bin = 0x01
		}
		boolBuf |= bin << boolPos
		boolPos++

		if boolPos != 8 || i == len(x.Values)-1 {
			continue
		}

		// Write the buffer.
		_, err = w.Write([]byte{boolBuf})
		if err != nil {
			return bytesWritten, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write array element")
		}
		bytesWritten++
		boolBuf = 0x00
		boolPos = 0
	}
	return bytesWritten, nil
}
