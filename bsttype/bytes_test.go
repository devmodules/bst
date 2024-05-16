package bsttype

import (
	"bytes"
	"math"
	"testing"

	"github.com/devmodules/bst/bstio"
)

var bytesTypeTestCases = []struct {
	Name   string
	Type   *Bytes
	Binary []byte
}{
	{
		Name:   "FixedSize",
		Type:   &Bytes{FixedSize: 30},
		Binary: []byte{bstio.BinarySizeUint8 | 0x80, 30},
	},
	{
		Name:   "FixedSizeMaxUint16",
		Type:   &Bytes{FixedSize: math.MaxUint16 + 1},
		Binary: []byte{0x03 | 0x80, 0x01, 0x00, 0x00},
	},
	{
		Name:   "VariableSize",
		Type:   &Bytes{},
		Binary: []byte{0x00},
	},
}

func TestBytesType_ReadType(t *testing.T) {
	for _, tc := range bytesTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			bt, ok := emptyKindType(KindBytes, false).(*Bytes)
			if !ok {
				t.Fatal("expected BytesType")
			}
			n, err := bt.ReadType(bytes.NewReader(tc.Binary))
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected to read %d bytes, got %d", len(tc.Binary), n)
			}

			if bt.FixedSize != tc.Type.FixedSize {
				t.Fatalf("expected FixedSize %d, got %d", tc.Type.FixedSize, bt.FixedSize)
			}
		})
	}
}

func TestBytesType_WriteType(t *testing.T) {
	for _, tc := range bytesTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			bt := tc.Type
			buf := bytes.Buffer{}
			n, err := bt.WriteType(&buf)
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected to write %d bytes, got %d", len(tc.Binary), n)
			}

			if !bytes.Equal(buf.Bytes(), tc.Binary) {
				t.Fatalf("expected value %v, got %v", tc.Binary, buf.Bytes())
			}
		})
	}
}

func TestBytesType_Skip(t *testing.T) {
	for _, tc := range bytesTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			bt := tc.Type
			n, err := bt.SkipType(bytes.NewReader(tc.Binary))
			if err != nil {
				t.Fatal(err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("expected to skip %d bytes, got %d", len(tc.Binary), n)
			}
		})
	}
}
