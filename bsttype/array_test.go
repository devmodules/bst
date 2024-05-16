package bsttype

import (
	"bytes"
	"testing"

	"github.com/devmodules/bst/bstio"
)

var arrayTypeTestCases = []struct {
	Name   string
	Type   Array
	Binary []byte
}{
	{
		Name: "ArrayType/FixedSize/Int32",
		Type: Array{
			Type:      Int32(),
			FixedSize: 3,
		},
		Binary: []byte{
			// Kind of the array type content.
			byte(KindInt32),
			// Fixed size of the array.
			bstio.BinarySizeUint8 | 0x80, 0x03,
		},
	},
	{
		Name: "ArrayType/FixedSize/UUID",
		Type: Array{
			Type:      &Bytes{FixedSize: 16},
			FixedSize: 4,
		},
		Binary: []byte{
			// Kind of the array type content.
			byte(KindBytes),
			// UUID fixed size content.
			bstio.BinarySizeUint8 | 0x80, 16,
			// Fixed size of the array.
			bstio.BinarySizeUint8 | 0x80, 4,
		},
	},
	{
		Name: "ArrayType/VarSize/Int32",
		Type: Array{
			Type: Int32(),
		},
		Binary: []byte{
			// Kind of the array type content.
			byte(KindInt32),
			// Empty array size byte.
			0x00,
		},
	},
}

func TestArrayType_WriteType(t *testing.T) {
	for _, tc := range arrayTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var buf bytes.Buffer
			n, err := tc.Type.WriteType(&buf)
			if err != nil {
				t.Fatal(err)
			}
			if n != len(tc.Binary) {
				t.Fatalf("expected to write %d bytes, but wrote %d", len(tc.Binary), n)
			}

			if !bytes.Equal(buf.Bytes(), tc.Binary) {
				t.Fatalf("expected %v, but got %v", tc.Binary, buf.Bytes())
			}
		})
	}
}
