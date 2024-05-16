package bsttype

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/internal/diff"
)

var _embedStructType = &Struct{
	Fields: []StructField{
		{
			Index: 0,
			Name:  "String",
			Type:  String(),
		},
	},
}

var structTypeTestCases = []struct {
	Name   string
	Type   Struct
	Binary []byte
}{
	{
		Name: "StructTypeTest1",
		Type: Struct{
			Fields: []StructField{
				{Index: 0, Name: "String", Type: String()},
				{Index: 1, Name: "Uint8", Type: Uint8()},
			}},
		Binary: []byte{
			// Fields
			// Fields length
			bstio.BinarySizeUint8, byte(2),
			// String.Index
			bstio.BinarySizeZero,
			// String.Name
			bstio.BinarySizeUint8, byte(len("String")),
			// String value
			'S', 't', 'r', 'i', 'n', 'g',
			// String Type
			byte(KindString),
			// Uint8 field
			// Uint8.Index
			bstio.BinarySizeUint8, byte(1),
			// Uint8 Name
			bstio.BinarySizeUint8, byte(len("Uint8")),
			// Uint8 value
			'U', 'i', 'n', 't', '8',
			// Uint8 Type
			byte(KindUint8),
		},
	},
	{
		Name: "Empty",
		Type: Struct{},
		Binary: []byte{
			// Fields
			// Fields length
			bstio.BinarySizeZero,
		},
	},
	{
		Name: "Embedded",
		Type: Struct{Fields: []StructField{
			{Index: 0, Name: "EmbeddedStruct", Type: _embedStructType}},
		},
		Binary: []byte{
			// Fields
			// Fields length
			bstio.BinarySizeUint8, byte(1),
			// Field 1: EmbeddedStruct
			// Field.Index
			bstio.BinarySizeZero,
			// Field.Name
			// Field.Name length
			bstio.BinarySizeUint8, byte(len("EmbeddedStruct")),
			// FieldName value
			'E', 'm', 'b', 'e', 'd', 'd', 'e', 'd', 'S', 't', 'r', 'u', 'c', 't',
			// Field.Type
			// Field.Type Kind
			byte(KindStruct),
			// Embedded.Fields length (1)
			bstio.BinarySizeUint8, byte(1),
			// Embedded.Fields.Field 1: String
			// Embedded.Fields.Field.Index
			bstio.BinarySizeZero,
			// Embedded.Fields.Field.Name
			// Embedded.Fields.Field.Name length
			bstio.BinarySizeUint8, byte(len("String")),
			// Embedded.Fields.FieldName value
			'S', 't', 'r', 'i', 'n', 'g',
			// Embedded.Fields.Field.Type
			// Embedded.Fields.Field.Type Kind
			byte(KindString),
		},
	},
}

func TestStructType_ReadType(t *testing.T) {
	for _, testCase := range structTypeTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// 1. Create a new StructType.
			st := Struct{}

			// 2. Read the value.
			bytesRead, err := st.ReadType(bytes.NewReader(testCase.Binary))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// 3. Check the number of bytes read.
			if bytesRead != len(testCase.Binary) {
				t.Fatalf("unexpected number of bytes read: %d, expected: %d", bytesRead, len(testCase.Binary))
			}

			// 4. Check the value.
			if !reflect.DeepEqual(st, testCase.Type) {
				t.Fatalf("unexpected value: %v, wanted: %v", st, testCase.Type)
			}
		})
	}
}

func TestStructType_WriteType(t *testing.T) {
	for _, testCase := range structTypeTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// 1. Create a new StructType.
			cp := testCase.Type
			st := &cp

			// 2. Write the value.
			var buf bytes.Buffer
			bytesWritten, err := st.WriteType(&buf)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// 3. Check the number of bytes written.
			if bytesWritten != len(testCase.Binary) {
				t.Fatalf("unexpected number of bytes written: %d - expected: %d", bytesWritten, len(testCase.Binary))
			}

			// 4. Check the value.
			if !bytes.Equal(buf.Bytes(), testCase.Binary) {
				t.Fatalf("unexpected value: %v", diff.DiffBytes(testCase.Binary, buf.Bytes()))
			}
		})
	}
}

func TestStructType_SkipType(t *testing.T) {
	for _, tc := range structTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// 1. Create a new StructType.
			st := Struct{}

			// 2. Skip the value.
			skipped, err := st.SkipType(bytes.NewReader(tc.Binary))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// 3. Check the number of bytes read.
			if int(skipped) != len(tc.Binary) {
				t.Fatalf("unexpected number of bytes read: %d, expected: %d", skipped, len(tc.Binary))
			}
		})
	}
}
