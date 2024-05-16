package bstvalue

import (
	"bytes"
	"math"
	"reflect"
	"testing"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

var anyValueTestCases = []struct {
	Name   string
	Value  Value
	Binary []byte
}{
	{
		Name:  "UUID",
		Value: MustNewBytes(uuidTestValues[0][:], uuidBytesType),
		Binary: []byte{
			// Bytes Type Header
			byte(bsttype.KindBytes),
			// Fixed size header && size of the array.
			bstio.BinarySizeUint8 | 0x80, 16,
			// First UUID
			uuidTestValues[0][0], uuidTestValues[0][1], uuidTestValues[0][2], uuidTestValues[0][3],
			uuidTestValues[0][4], uuidTestValues[0][5], uuidTestValues[0][6], uuidTestValues[0][7],
			uuidTestValues[0][8], uuidTestValues[0][9], uuidTestValues[0][10], uuidTestValues[0][11],
			uuidTestValues[0][12], uuidTestValues[0][13], uuidTestValues[0][14], uuidTestValues[0][15],
		},
	},
	{
		Name:  "Uint8",
		Value: NewUint8Value(math.MaxUint8),
		Binary: []byte{
			// Uint8 Type Header
			byte(bsttype.KindUint8),
			// Uint8 value.
			0xFF,
		},
	},
	{
		Name:  "String",
		Value: NewStringValue("Hello World"),
		Binary: []byte{
			// String Type Header
			byte(bsttype.KindString),
			// Size of the string.
			bstio.BinarySizeUint8, byte(len("Hello World")),
			// String value.
			'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd',
		},
	},
}

func TestAnyValue_ReadValue(t *testing.T) {
	for _, tc := range anyValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var av AnyValue
			n, err := av.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if n != len(tc.Binary) {
				t.Errorf("Expected %d bytes read, got %d", len(tc.Binary), n)
			}
			if !bsttype.TypesEqual(av.Value.Type(), tc.Value.Type()) {
				t.Errorf("Expected type %v, got %v", tc.Value.Type(), av.Value.Type())
			}
			if !reflect.DeepEqual(av.Value, tc.Value) {
				t.Errorf("Expected value %v, got %v", tc.Value, av.Value)
			}
		})
	}
}

func TestAnyValue_WriteValue(t *testing.T) {
	for _, tc := range anyValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			av := AnyValueOf(tc.Value)

			var buf bytes.Buffer
			n, err := av.WriteValue(&buf, bstio.ValueOptions{})
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if n != len(tc.Binary) {
				t.Errorf("Expected %d bytes written, got %d", len(tc.Binary), n)
			}
			if !bytes.Equal(buf.Bytes(), tc.Binary) {
				t.Errorf("Expected value %v, got %v", tc.Binary, buf.Bytes())
			}
		})
	}
}

func TestAnyValue_MarshalDB(t *testing.T) {
	for _, tc := range anyValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			av := AnyValueOf(tc.Value)

			data, err := av.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !bytes.Equal(data, tc.Binary) {
				t.Errorf("Expected value %v, got %v", tc.Binary, data)
			}
		})
	}
}

func TestAnyValue_UnmarshalValue(t *testing.T) {
	for _, tc := range anyValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var av AnyValue
			err := av.UnmarshalValue(tc.Binary, bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !bsttype.TypesEqual(av.Value.Type(), tc.Value.Type()) {
				t.Errorf("Expected type %v, got %v", tc.Value.Type(), av.Value.Type())
			}
			if !reflect.DeepEqual(av.Value, tc.Value) {
				t.Errorf("Expected value %v, got %v", tc.Value, av.Value)
			}
		})
	}
}
