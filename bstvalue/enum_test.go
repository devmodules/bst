package bstvalue

import (
	"bytes"
	"testing"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

var testEnumType = &bsttype.Enum{
	Elements: []bsttype.EnumElement{
		{String: "foo", Index: 0},
		{String: "bar", Index: 1},
		{String: "baz", Index: 2},
	},
	ValueBytes: bstio.BinarySizeUint8,
}

var enumValueTestCases = []struct {
	Name   string
	Value  string
	Binary []byte
}{
	{
		Name:   "foo",
		Value:  "foo",
		Binary: []byte{0x00},
	},
	{
		Name:   "bar",
		Value:  "bar",
		Binary: []byte{0x01},
	},
	{
		Name:   "baz",
		Value:  "baz",
		Binary: []byte{0x02},
	},
}

func TestEnumValue_ReadValue(t *testing.T) {
	for _, tc := range enumValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := EmptyEnumValue(testEnumType)
			n, err := v.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if n != len(tc.Binary) {
				t.Fatalf("unexpected number of bytes read: %d", n)
			}
			str, ok := v.IndexString()
			if !ok {
				t.Fatalf("buffIndex string not found")
			}
			if str != tc.Value {
				t.Fatalf("unexpected value: %q - expected: %q", str, tc.Value)
			}
		})
	}
}

func TestEnumValue_WriteValue(t *testing.T) {
	for _, tc := range enumValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v, err := NewEnumStringValue(testEnumType, tc.Value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var buf bytes.Buffer
			n, err := v.WriteValue(&buf, bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if n != len(tc.Binary) {
				t.Fatalf("unexpected number of bytes written: %d", n)
			}
			if !bytes.Equal(buf.Bytes(), tc.Binary) {
				t.Fatalf("unexpected bytes written: %v - expected: %v", buf.Bytes(), tc.Binary)
			}
		})
	}
}

func TestEnumValue_Skip(t *testing.T) {
	for _, tc := range enumValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := EmptyEnumValue(testEnumType)
			n, err := v.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if int(n) != len(tc.Binary) {
				t.Fatalf("unexpected number of bytes read: %d, wanted: %d", n, len(tc.Binary))
			}
		})
	}
}

func TestEnumValue_MarshalDB(t *testing.T) {
	for _, tc := range enumValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v, err := NewEnumStringValue(testEnumType, tc.Value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			data, err := v.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Equal(data, tc.Binary) {
				t.Fatalf("unexpected bytes written: %v - expected: %v", data, tc.Binary)
			}
		})
	}
}

func TestEnumValue_UnmarshalValue(t *testing.T) {
	for _, tc := range enumValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := EmptyEnumValue(testEnumType)
			err := v.UnmarshalValue(tc.Binary, bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			str, ok := v.IndexString()
			if !ok {
				t.Fatalf("buffIndex string not found")
			}
			if str != tc.Value {
				t.Fatalf("unexpected value: %q - expected: %q", str, tc.Value)
			}
		})
	}
}
