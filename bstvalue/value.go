package bstvalue

import (
	"io"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// Value is a value of the type.
type Value interface {
	// Type returns the type of the value.
	Type() bsttype.Type
	// Kind returns the basic kind of the value.
	Kind() bsttype.Kind
	// Skip the bytes in the reader to the next value.
	Skip(br io.ReadSeeker, options bstio.ValueOptions) (int64, error)
	// MarshalValue marshals the value to a binary database format.
	MarshalValue(options bstio.ValueOptions) ([]byte, error)
	// UnmarshalValue unmarshals the value from a binary database format.
	UnmarshalValue(in []byte, options bstio.ValueOptions) error
	// ReadValue reads the value from a reader.
	ReadValue(r io.Reader, options bstio.ValueOptions) (int, error)
	// WriteValue writes the value to a writer.
	WriteValue(w io.Writer, options bstio.ValueOptions) (int, error)
	// String returns a human-readable string representation of the value.
	String() string
}

var _StdTypeValues = [bsttype.KindOneOf + 1]func(bsttype.Type) Value{
	bsttype.KindUndefined: emptyUndefinedValue,
	bsttype.KindBoolean:   emptyBoolValue,
	bsttype.KindInt:       emptyIntValue,
	bsttype.KindInt8:      emptyInt8Value,
	bsttype.KindInt16:     emptyInt16Value,
	bsttype.KindInt32:     emptyInt32Value,
	bsttype.KindInt64:     emptyInt64Value,
	bsttype.KindUint:      emptyUintValue,
	bsttype.KindUint8:     emptyUint8Value,
	bsttype.KindUint16:    emptyUint16Value,
	bsttype.KindUint32:    emptyUint32Value,
	bsttype.KindUint64:    emptyUint64Value,
	bsttype.KindFloat32:   emptyFloat32Value,
	bsttype.KindFloat64:   emptyFloat64Value,
	bsttype.KindString:    emptyStringValue,
	bsttype.KindBytes:     emptyBytesValue,
	bsttype.KindArray:     emptyArrayValue,
	bsttype.KindDuration:  emptyDurationValue,
	bsttype.KindTimestamp: emptyTimestampValue,
	bsttype.KindAny:       emptyAnyValue,
}

func init() {
	// The value initializers that depends on the _StdTypes directly or indirectly
	// needs to be initialized in runtime - as the compiler does not allow dependency cycles.
	_StdTypeValues[bsttype.KindNullable] = emptyNullableValue
	_StdTypeValues[bsttype.KindMap] = emptyMapValue
	_StdTypeValues[bsttype.KindEnum] = emptyEnumValue
	_StdTypeValues[bsttype.KindStruct] = emptyStructValue
	_StdTypeValues[bsttype.KindDateTime] = emptyDateTimeValue
	_StdTypeValues[bsttype.KindOneOf] = emptyOneOfValue
	_StdTypeValues[bsttype.KindNamed] = emptyNamedValue
}

func emptyNamedValue(t bsttype.Type) Value {
	nt, ok := t.(*bsttype.Named)
	if !ok {
		return nil
	}
	return EmptyValueOf(nt.Type)
}

func emptyStructValue(t bsttype.Type) Value {
	st := t.(*bsttype.Struct)
	fv := make([]Value, len(st.Fields))
	for i, f := range st.Fields {
		fv[i] = EmptyValueOf(f.Type)
	}
	return &StructValue{StructType: st, Fields: fv}
}

// EmptyValueOf creates an empty value of the given type.
func EmptyValueOf(t bsttype.Type) Value {
	k := t.Kind()
	return _StdTypeValues[k](t)
}

// ElementValue returns the element value of the array value.
// Implements the Value interface.
type ElementValue interface {
	Value
	// Elem dereferences the pointer wrapped Value.
	Elem() Value
}

// Marshaler is the interface wrapper for overwritten bst.Marshaler.
type Marshaler interface {
	// MarshalValue marshals the value to a binary database format.
	MarshalValue(options bstio.ValueOptions) ([]byte, error)
}

// Unmarshaler is the interface wrapper for overwritten bst.Unmarshaler.
type Unmarshaler interface {
	// UnmarshalValue unmarshals the value from a binary database format.
	UnmarshalValue(in []byte, options bstio.ValueOptions) error
}

// ValueWriter is the interface used by the Value that could write itself to a writer.
type ValueWriter interface {
	// WriteValue writes the value to a writer.
	WriteValue(w io.Writer, options bstio.ValueOptions) (int, error)
}

// ValueReader is the interface used by the Value that could read itself from a reader.
type ValueReader interface {
	// ReadValue reads the value from a reader.
	ReadValue(r io.Reader, options bstio.ValueOptions) (int, error)
}
