package bstvalue

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

var nullableValueTestCases = []struct {
	Name   string
	Type   bsttype.Type
	Value  Value
	IsNull bool
	Binary []byte
}{
	{
		Name:   "Null/String",
		Type:   bsttype.String(),
		Value:  EmptyStringValue(),
		IsNull: true,
		Binary: []byte{0x00},
	},
	{
		Name:  "NotNull/String",
		Type:  bsttype.String(),
		Value: NewStringValue("test value"),
		Binary: []byte{
			// Not Null flag
			0x01,
			// String Len
			bstio.BinarySizeUint8, byte(len("test value")),
			// String
			't', 'e', 's', 't', ' ', 'v', 'a', 'l', 'u', 'e',
		},
	},
	{
		Name:   "Null/Bytes",
		Type:   &bsttype.Bytes{},
		Value:  EmptyBytes(&bsttype.Bytes{}),
		IsNull: true,
		Binary: []byte{
			// Null Value header
			0x00,
		},
	},
	{
		Name:  "NotNull/Bytes",
		Type:  &bsttype.Bytes{},
		Value: MustNewBytes([]byte("test value"), &bsttype.Bytes{}),
		Binary: []byte{
			// Not Null flag
			0x01,
			// Bytes Len
			bstio.BinarySizeUint8, byte(len("test value")),
			// Bytes
			't', 'e', 's', 't', ' ', 'v', 'a', 'l', 'u', 'e',
		},
	},
	{
		Name:  "NotNull/UUID",
		Type:  uuidBytesType,
		Value: &Bytes{BytesType: uuidBytesType, Value: uuidTestValues[0][:]},
		Binary: []byte{
			// Not Null flag
			0x01,
			// UUID
			uuidTestValues[0][0], uuidTestValues[0][1], uuidTestValues[0][2], uuidTestValues[0][3],
			uuidTestValues[0][4], uuidTestValues[0][5], uuidTestValues[0][6], uuidTestValues[0][7],
			uuidTestValues[0][8], uuidTestValues[0][9], uuidTestValues[0][10], uuidTestValues[0][11],
			uuidTestValues[0][12], uuidTestValues[0][13], uuidTestValues[0][14], uuidTestValues[0][15],
		},
	},
	{
		Name:   "Null/UUID",
		Type:   uuidBytesType,
		Value:  EmptyBytes(uuidBytesType),
		IsNull: true,
		Binary: []byte{
			// Null Value header
			0x00,
		},
	},
	{
		Name:  "NotNull/Uint16",
		Type:  bsttype.Uint16(),
		Value: NewUint16Value(uint16(12345)),
		Binary: []byte{
			// Not Null flag
			0x01,
			// Uint16
			0x30, 0x39,
		},
	},
}

func TestNullableValue_ReadValue(t *testing.T) {
	for _, testCase := range nullableValueTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ev := EmptyNullableValue(testCase.Type)
			n, err := ev.ReadValue(bytes.NewReader(testCase.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(testCase.Binary) {
				t.Fatalf("expected to read %d bytes, got %d", len(testCase.Binary), n)
			}

			if ev.IsNull != testCase.IsNull {
				t.Fatalf("expected IsNull to be %t, got %t", testCase.IsNull, ev.IsNull)
			}

			if !reflect.DeepEqual(ev.Value, testCase.Value) {
				t.Fatalf("expected value %v, got %v", testCase.Value, ev.Value)
			}
		})
	}
}

func TestNullableValue_WriteValue(t *testing.T) {
	for _, testCase := range nullableValueTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			nv, err := NullableValueOf(testCase.Value, testCase.IsNull)
			if err != nil {
				t.Fatal(err)
			}
			var buf bytes.Buffer
			n, err := nv.WriteValue(&buf, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(testCase.Binary) {
				t.Fatalf("expected to write %d bytes, got %d", len(testCase.Binary), n)
			}

			if !bytes.Equal(buf.Bytes(), testCase.Binary) {
				t.Fatalf("expected to write %v, got %v", testCase.Binary, buf.Bytes())
			}
		})
	}
}

func TestNullableValue_Skip(t *testing.T) {
	for _, tc := range nullableValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			nv, err := NullableValueOf(tc.Value, tc.IsNull)
			if err != nil {
				t.Fatal(err)
			}
			n, err := nv.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("expected to skip %d bytes, got %d", len(tc.Binary), n)
			}
		})
	}
}
