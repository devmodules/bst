package bstvalue

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
	"github.com/google/uuid"
)

var uuidTestValues = []uuid.UUID{
	uuid.MustParse("831da871-b26e-4864-8097-82630fa4b4a3"),
	uuid.MustParse("831da871-b26e-4864-8097-82630fa4b4a4"),
	uuid.MustParse("831da871-b26e-4864-8097-82630fa4b4a5"),
}

var arrayTestCases = []struct {
	Name   string
	Values []Value
	Binary []byte
	Type   bsttype.Array
}{
	{
		Name: "UUID/VarSize",
		Values: []Value{
			MustNewBytes(uuidTestValues[0][:], uuidBytesType),
			MustNewBytes(uuidTestValues[1][:], uuidBytesType),
			MustNewBytes(uuidTestValues[2][:], uuidBytesType),
		},
		Type: bsttype.Array{Type: uuidBytesType},
		Binary: []byte{
			// Size of the array.
			bstio.BinarySizeUint8, 0x03,
			// First UUID
			uuidTestValues[0][0], uuidTestValues[0][1], uuidTestValues[0][2], uuidTestValues[0][3],
			uuidTestValues[0][4], uuidTestValues[0][5], uuidTestValues[0][6], uuidTestValues[0][7],
			uuidTestValues[0][8], uuidTestValues[0][9], uuidTestValues[0][10], uuidTestValues[0][11],
			uuidTestValues[0][12], uuidTestValues[0][13], uuidTestValues[0][14], uuidTestValues[0][15],
			// Second UUID
			uuidTestValues[1][0], uuidTestValues[1][1], uuidTestValues[1][2], uuidTestValues[1][3],
			uuidTestValues[1][4], uuidTestValues[1][5], uuidTestValues[1][6], uuidTestValues[1][7],
			uuidTestValues[1][8], uuidTestValues[1][9], uuidTestValues[1][10], uuidTestValues[1][11],
			uuidTestValues[1][12], uuidTestValues[1][13], uuidTestValues[1][14], uuidTestValues[1][15],
			// Third UUID
			uuidTestValues[2][0], uuidTestValues[2][1], uuidTestValues[2][2], uuidTestValues[2][3],
			uuidTestValues[2][4], uuidTestValues[2][5], uuidTestValues[2][6], uuidTestValues[2][7],
			uuidTestValues[2][8], uuidTestValues[2][9], uuidTestValues[2][10], uuidTestValues[2][11],
			uuidTestValues[2][12], uuidTestValues[2][13], uuidTestValues[2][14], uuidTestValues[2][15],
		},
	},
	{
		Name: "UUID/FixedSize",
		Values: []Value{
			MustNewBytes(uuidTestValues[0][:], uuidBytesType),
			MustNewBytes(uuidTestValues[1][:], uuidBytesType),
			MustNewBytes(uuidTestValues[2][:], uuidBytesType),
			EmptyBytes(uuidBytesType),
		},
		Type: bsttype.Array{Type: uuidBytesType, FixedSize: 4},
		Binary: []byte{
			// First UUID
			uuidTestValues[0][0], uuidTestValues[0][1], uuidTestValues[0][2], uuidTestValues[0][3],
			uuidTestValues[0][4], uuidTestValues[0][5], uuidTestValues[0][6], uuidTestValues[0][7],
			uuidTestValues[0][8], uuidTestValues[0][9], uuidTestValues[0][10], uuidTestValues[0][11],
			uuidTestValues[0][12], uuidTestValues[0][13], uuidTestValues[0][14], uuidTestValues[0][15],
			// Second UUID
			uuidTestValues[1][0], uuidTestValues[1][1], uuidTestValues[1][2], uuidTestValues[1][3],
			uuidTestValues[1][4], uuidTestValues[1][5], uuidTestValues[1][6], uuidTestValues[1][7],
			uuidTestValues[1][8], uuidTestValues[1][9], uuidTestValues[1][10], uuidTestValues[1][11],
			uuidTestValues[1][12], uuidTestValues[1][13], uuidTestValues[1][14], uuidTestValues[1][15],
			// Third UUID
			uuidTestValues[2][0], uuidTestValues[2][1], uuidTestValues[2][2], uuidTestValues[2][3],
			uuidTestValues[2][4], uuidTestValues[2][5], uuidTestValues[2][6], uuidTestValues[2][7],
			uuidTestValues[2][8], uuidTestValues[2][9], uuidTestValues[2][10], uuidTestValues[2][11],
			uuidTestValues[2][12], uuidTestValues[2][13], uuidTestValues[2][14], uuidTestValues[2][15],
			// Empty UUID value
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
		},
	},
	{
		Name:   "UUID/VarSize/Empty",
		Values: []Value{},
		Type:   bsttype.Array{Type: uuidBytesType},
		Binary: []byte{
			// Size of the array.
			bstio.BinarySizeZero,
		},
	},
}

func TestArrayValue_ReadValue(t *testing.T) {
	for _, tc := range arrayTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			av := EmptyArrayValue(&tc.Type)
			n, err := av.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if n != len(tc.Binary) {
				t.Fatalf("expected to read %d bytes, but read %d", len(tc.Binary), n)
			}

			if !reflect.DeepEqual(av.Values, tc.Values) {
				t.Fatalf("expected %v, but got %v", tc.Values, av.Values)
			}
		})
	}
}

func TestArrayValue_WriteValue(t *testing.T) {
	for _, tc := range arrayTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			av, err := ArrayValueOf(&tc.Type, tc.Values)
			if err != nil {
				t.Fatal(err)
			}

			var buf bytes.Buffer
			n, err := av.WriteValue(&buf, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if n != len(tc.Binary) {
				t.Fatalf("expected to write %d bytes, but wrote %d", len(tc.Binary), n)
			}

			if !reflect.DeepEqual(buf.Bytes(), tc.Binary) {
				t.Fatalf("expected %v, but got %v", tc.Binary, buf.Bytes())
			}
		})
	}
}

func TestArrayValue_MarshalDB(t *testing.T) {
	for _, tc := range arrayTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			av, err := ArrayValueOf(&tc.Type, tc.Values)
			if err != nil {
				t.Fatal(err)
			}

			data, err := av.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(data, tc.Binary) {
				t.Fatalf("expected %v, but got %v", tc.Binary, data)
			}
		})
	}
}

func TestArrayValue_UnmarshalValue(t *testing.T) {
	for _, tc := range arrayTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			av := EmptyArrayValue(&tc.Type)
			err := av.UnmarshalValue(tc.Binary, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(av.Values, tc.Values) {
				t.Fatalf("expected %v, but got %v", tc.Values, av.Values)
			}
		})
	}
}

func TestArrayValue_Skip(t *testing.T) {
	for _, tc := range arrayTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			av, err := ArrayValueOf(&tc.Type, tc.Values)
			if err != nil {
				t.Fatal(err)
			}

			n, err := av.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if int(n) != len(tc.Binary) {
				t.Fatalf("expected to skip %d bytes, but skipped %d", len(tc.Binary), n)
			}
		})
	}
}
