package bsttype

import (
	"bytes"
	"math"
	"reflect"
	"testing"

	"github.com/devmodules/bst/bstio"
)

var enumTypeTestCases = []struct {
	Name   string
	Type   Enum
	Binary []byte
}{
	{
		Name: "Uint8Size/2-Elements",
		Type: Enum{
			Elements: []EnumElement{
				{String: "foo", Index: 0},
				{String: "bar", Index: 1},
			},
			ValueBytes: bstio.BinarySizeUint8,
		},
		Binary: []byte{
			// Value bits header.
			bstio.BinarySizeUint8,
			// Element number header.
			bstio.BinarySizeUint8, 0x02,
			//
			// Element 0.
			//
			// String number.
			bstio.BinarySizeUint8, byte(len("foo")),
			// String data.
			'f', 'o', 'o',
			// Index Value.
			0x00,
			//
			// Element 1.
			//
			// String number.
			bstio.BinarySizeUint8, byte(len("bar")),
			// String data.
			'b', 'a', 'r',
			// Index Value.
			0x01,
		},
	},
	{
		Name: "Uint16Size/2-Elements",
		Type: Enum{
			ValueBytes: bstio.BinarySizeUint16,
			Elements: []EnumElement{
				{String: "foo", Index: 0},
				{String: "bar", Index: math.MaxUint16},
			},
		},
		Binary: []byte{
			// Value bits header.
			bstio.BinarySizeUint16,
			// Element number header.
			bstio.BinarySizeUint8, 0x02,
			//
			// Element 0.
			//
			// String number.
			bstio.BinarySizeUint8, byte(len("foo")),
			// String data.
			'f', 'o', 'o',
			// Index Value.
			0x00, 0x00,
			//
			// Element 1.
			//
			// String number.
			bstio.BinarySizeUint8, byte(len("bar")),
			// String data.
			'b', 'a', 'r',
			// Index Value.
			0xff, 0xff,
		},
	},
	{
		Name: "Uint32Size/2-Elements",
		Type: Enum{
			ValueBytes: bstio.BinarySizeUint32,
			Elements: []EnumElement{
				{String: "foo", Index: 0},
				{String: "bar", Index: math.MaxUint32},
			},
		},
		Binary: []byte{
			// Value bits header.
			bstio.BinarySizeUint32,
			// Element number header.
			bstio.BinarySizeUint8, 0x02,
			//
			// Element 0.
			//
			// String number.
			bstio.BinarySizeUint8, byte(len("foo")),
			// String data.
			'f', 'o', 'o',
			// Index Value.
			0x00, 0x00, 0x00, 0x00,
			//
			// Element 1.
			//
			// String number.
			bstio.BinarySizeUint8, byte(len("bar")),
			// String data.
			'b', 'a', 'r',
			// Index Value.
			0xff, 0xff, 0xff, 0xff,
		},
	},
	{
		Name: "Uint64Size/2-Elements",
		Type: Enum{
			ValueBytes: bstio.BinarySizeUint64,
			Elements: []EnumElement{
				{String: "foo", Index: 0},
				{String: "bar", Index: math.MaxUint64},
			},
		},
		Binary: []byte{
			// Value bits header.
			bstio.BinarySizeUint64,
			// Element number header.
			bstio.BinarySizeUint8, 0x02,
			//
			// Element 0.
			//
			// String number.
			bstio.BinarySizeUint8, byte(len("foo")),
			// String data.
			'f', 'o', 'o',
			// Index Value.
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			//
			// Element 1.
			//
			// String number.
			bstio.BinarySizeUint8, byte(len("bar")),
			// String data.
			'b', 'a', 'r',
			// Index Value.
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		},
	},
}

func TestEnumType_ReadType(t *testing.T) {
	for _, tc := range enumTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var et Enum
			n, err := et.ReadType(bytes.NewReader(tc.Binary))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("unexpected read length: %d - expected: %d", n, len(tc.Binary))
			}

			if !reflect.DeepEqual(et, tc.Type) {
				t.Fatalf("unexpected type: %v - expected: %v", et, tc.Type)
			}
		})
	}
}

func TestEnumType_WriteType(t *testing.T) {
	for _, tc := range enumTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var buf bytes.Buffer
			n, err := tc.Type.WriteType(&buf)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("unexpected write length: %d - expected: %d", n, len(tc.Binary))
			}

			if !reflect.DeepEqual(buf.Bytes(), tc.Binary) {
				t.Fatalf("unexpected binary: %v - expected: %v", buf.Bytes(), tc.Binary)
			}
		})
	}
}

func TestEnumType_SkipType(t *testing.T) {
	for _, tc := range enumTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var et Enum
			n, err := et.SkipType(bytes.NewReader(tc.Binary))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("unexpected read length: %d - expected: %d", n, len(tc.Binary))
			}
		})
	}
}
