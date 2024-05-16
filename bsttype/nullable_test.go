package bsttype

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/devmodules/bst/bstio"
)

var nullableTypeTestCases = []struct {
	Name   string
	Type   Nullable
	Binary []byte
}{
	{
		Name: "Nullable/Int",
		Type: Nullable{
			Type: Int(),
		},
		Binary: []byte{
			// Wrapped Type
			byte(KindInt),
		},
	},
	{
		Name: "Nullable/Bytes/Fixed",
		Type: Nullable{
			Type: &Bytes{
				FixedSize: 5,
			},
		},
		Binary: []byte{
			// Wrapped Type Kind
			byte(KindBytes),
			// Wrapped Type Content
			// Fixed Size binary length with the most significant bit set to 1
			bstio.BinarySizeUint8 | 1<<7,
			// Fixed Size length
			5,
		},
	},
	{
		Name: "Nullable/Any",
		Type: Nullable{Type: Any()},
		Binary: []byte{
			// Nullable Embed Type Header
			byte(KindAny),
		},
	},
}

func TestNullableType_ReadType(t *testing.T) {
	for _, tc := range nullableTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var nt Nullable
			n, err := nt.ReadType(bytes.NewReader(tc.Binary))
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected to read %d bytes, got %d", len(tc.Binary), n)
			}

			if !reflect.DeepEqual(nt, tc.Type) {
				t.Fatalf("expected type %v, got %v", tc.Type, nt)
			}
		})
	}
}

func TestNullableType_WriteType(t *testing.T) {
	for _, tc := range nullableTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var buf bytes.Buffer
			n, err := tc.Type.WriteType(&buf)
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected to write %d bytes, got %d", len(tc.Binary), n)
			}

			if !bytes.Equal(buf.Bytes(), tc.Binary) {
				t.Fatalf("expected to write %v, got %v", tc.Binary, buf.Bytes())
			}
		})
	}
}

func TestNullableType_SkipType(t *testing.T) {
	for _, tc := range nullableTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			n, err := tc.Type.SkipType(bytes.NewReader(tc.Binary))
			if err != nil {
				t.Fatal(err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("expected to skip %d bytes, got %d", len(tc.Binary), n)
			}
		})
	}
}

func TestNullableType_CompareType(t *testing.T) {
	testCases := []struct {
		Name  string
		Type  *Nullable
		Other Type
		Equal bool
	}{
		{
			Name:  "Nullable(Int)/Int",
			Type:  &Nullable{Type: Int()},
			Other: Int(),
			Equal: false,
		},
		{
			Name:  "Nullable(Int)/Nullable(Int)",
			Type:  &Nullable{Type: Int()},
			Other: &Nullable{Type: Int()},
			Equal: true,
		},
		{
			Name:  "Nullable(Int)/Nullable(Bytes)",
			Type:  &Nullable{Type: Int()},
			Other: &Nullable{Type: &Bytes{}},
			Equal: false,
		},
		{
			Name:  "Nullable(Bytes)/Nullable(Bytes(Fixed))",
			Type:  &Nullable{Type: &Bytes{}},
			Other: &Nullable{Type: &Bytes{FixedSize: 5}},
			Equal: false,
		},
		{
			Name:  "Nullable(Bytes)/Nullable(Bytes))",
			Type:  &Nullable{Type: &Bytes{}},
			Other: &Nullable{Type: &Bytes{}},
			Equal: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			res := TypesEqual(tc.Type, tc.Other)
			if res != tc.Equal {
				t.Fatalf("expected %v, got %v", tc.Equal, res)
			}
		})
	}
}
