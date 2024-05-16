package bstvalue

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

var _embedStructType = &bsttype.Struct{
	Fields: []bsttype.StructField{
		{
			Index: 0,
			Name:  "String",
			Type:  bsttype.String(),
		},
	},
}

var structTestCases = []struct {
	Name   string
	Type   bsttype.Struct
	Fields []Value
	Binary []byte
}{
	{
		// StructTest1 is a struct with two fields:
		// 	- String: string
		//  - Uint8:  uint8
		Name: "StructTest1",
		Type: bsttype.Struct{
			Fields: []bsttype.StructField{
				{
					Index: 0,
					Name:  "String",
					Type:  bsttype.String(),
				},
				{
					Index: 1,
					Name:  "Uint8",
					Type:  bsttype.Uint8(),
				},
			},
		},
		Fields: []Value{
			NewStringValue("hello"),
			NewUint8Value(1),
		},
		Binary: []byte{
			// String field
			// String length
			bstio.BinarySizeUint8, byte(len("hello")),
			// String value
			'h', 'e', 'l', 'l', 'o',
			// Uint8 field
			// Uint8 value
			1,
		},
	},
	{
		// Empty is a struct with no fields
		Name:   "Empty",
		Type:   bsttype.Struct{},
		Fields: []Value{},
		Binary: []byte{},
	},
	{
		// StructTest2 is a struct with embedded struct field.
		// 	- Embedded: Struct{String: string}
		Name: "StructTest2",
		Type: bsttype.Struct{
			Fields: []bsttype.StructField{
				{
					Name: "Embedded",
					Type: _embedStructType,
				},
			},
		},
		Fields: []Value{
			MustNewStructValue(_embedStructType, []Value{
				NewStringValue("hello"),
			}),
		},
		Binary: []byte{
			// Embedded struct field
			//    String: string
			//    String length
			bstio.BinarySizeUint8, byte(len("hello")),
			//    String value
			'h', 'e', 'l', 'l', 'o',
		},
	},
}

func TestStructValue_ReadValue(t *testing.T) {
	for _, testCase := range structTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// 1. Create a new StructValue.
			sv := EmptyStructValueOf(&testCase.Type)

			// 2. Read the value.
			bytesRead, err := sv.ReadValue(bytes.NewReader(testCase.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// 5. Check the number of bytes read.
			if bytesRead != len(testCase.Binary) {
				t.Fatalf("unexpected number of bytes read: %d", bytesRead)
			}

			// 6. Check the value.
			if !reflect.DeepEqual(sv.Fields, testCase.Fields) {
				t.Fatalf("unexpected value: %v", sv)
			}
		})
	}
}

func TestStructValue_WriteValue(t *testing.T) {
	for _, testCase := range structTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// 1. Create a new StructValue.
			sv, err := NewStructValue(&testCase.Type, testCase.Fields)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// 2. Write the value.
			var buf bytes.Buffer
			bytesWritten, err := sv.WriteValue(&buf, bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// 3. Check the number of bytes written.
			if bytesWritten != len(testCase.Binary) {
				t.Fatalf("unexpected number of bytes written: %d", bytesWritten)
			}

			// 4. Check the value.
			if !bytes.Equal(buf.Bytes(), testCase.Binary) {
				t.Fatalf("unexpected value: %v", buf.Bytes())
			}
		})
	}
}

func TestStructType_Skip(t *testing.T) {
	for _, tc := range structTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// 1. Create a new StructType.
			tp := tc.Type
			st := EmptyStructValueOf(&tp)

			// 2. Skip the value.
			skipped, err := st.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// 3. Check the number of bytes read.
			if int(skipped) != len(tc.Binary) {
				t.Fatalf("unexpected number of bytes read: %d, wanted: %d", skipped, len(tc.Binary))
			}
		})
	}
}
