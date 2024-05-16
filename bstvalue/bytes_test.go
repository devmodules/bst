package bstvalue

import (
	"bytes"
	"testing"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
	"github.com/google/uuid"
)

var bytesTestCases = []struct {
	Name   string
	Value  []byte
	Binary []byte
	Type   *bsttype.Bytes
	IsNull bool
}{
	{
		Name:   "Empty",
		Value:  []byte{},
		Binary: []byte{0x00},
		Type:   &bsttype.Bytes{},
	},
	{
		Name:   "Short",
		Value:  []byte("short"),
		Binary: []byte{bstio.BinarySizeUint8, 0x05, 's', 'h', 'o', 'r', 't'},
		Type:   &bsttype.Bytes{},
	},
	{
		Name: "MaxUint8",
		Value: func() []byte {
			buf := make([]byte, 256)
			for i := 0; i < 256; i++ {
				buf[i] = byte(i)
			}
			return buf
		}(),
		Binary: func() []byte {
			buf := bytes.Buffer{}
			buf.WriteByte(bstio.BinarySizeUint16)
			buf.WriteByte(0x01)
			buf.WriteByte(0x00)
			for i := 0; i < 256; i++ {
				buf.WriteByte(byte(i))
			}
			return buf.Bytes()
		}(),
		Type: &bsttype.Bytes{},
	},
	{
		Name: "MaxUint16",
		Value: func() []byte {
			buf := make([]byte, 65536)
			for i := 0; i < 65536; i++ {
				buf[i] = byte(i)
			}
			return buf
		}(),
		Binary: func() []byte {
			buf := bytes.Buffer{}
			buf.WriteByte(0x03)
			buf.WriteByte(0x01)
			buf.WriteByte(0x00)
			buf.WriteByte(0x00)
			for i := 0; i < 65536; i++ {
				buf.WriteByte(byte(i))
			}
			return buf.Bytes()
		}(),
		Type: &bsttype.Bytes{},
	},
	{
		Name: "UUID",
		Value: func() []byte {
			uid := uuid.MustParse("ae24721e-0dd8-4787-981e-d453cfd0ff96")
			return uid[:]
		}(),
		Binary: func() []byte {
			uid := uuid.MustParse("ae24721e-0dd8-4787-981e-d453cfd0ff96")
			return uid[:]
		}(),
		Type: &bsttype.Bytes{FixedSize: 16},
	},
}

func TestBytes(t *testing.T) {
	t.Run("VarSize", func(t *testing.T) {
		v := MustNewBytes([]byte{0x1, 0x2}, &bsttype.Bytes{})
		if v.Kind() != bsttype.KindBytes {
			t.Fatalf("expected kind %d, got %d", bsttype.KindBytes, v.Kind())
		}

		if v.Type() == nil {
			t.Fatal("expected type, got nil")
		}

		if v.Type().Kind() != bsttype.KindBytes {
			t.Fatalf("expected type kind %d, got %d", bsttype.KindBytes, v.Type().Kind())
		}

		bt := v.Type().(*bsttype.Bytes)
		if bt.FixedSize != 0 {
			t.Fatalf("expected fixed size 0, got %d", bt.FixedSize)
		}
	})

	t.Run("FixedSize", func(t *testing.T) {
		v := MustNewBytes([]byte{0x1, 0x2}, &bsttype.Bytes{FixedSize: 2})
		if v.Kind() != bsttype.KindBytes {
			t.Fatalf("expected kind %d, got %d", bsttype.KindBytes, v.Kind())
		}

		if v.Type() == nil {
			t.Fatal("expected type, got nil")
		}

		if v.Type().Kind() != bsttype.KindBytes {
			t.Fatalf("expected type kind %d, got %d", bsttype.KindBytes, v.Type().Kind())
		}

		bt := v.Type().(*bsttype.Bytes)
		if bt.FixedSize != 2 {
			t.Fatalf("expected fixed size 2, got %d", bt.FixedSize)
		}
	})
}

func TestBytesValue_ReadValue(t *testing.T) {
	for _, tc := range bytesTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			optsCases := []struct {
				Name                   string
				Descending, Comparable bool
			}{
				{"AscComp", false, true},
				{"DescComp", true, true},
				{"AscNonComp", false, false},
				{"DescNonComp", true, false},
			}

			for _, opts := range optsCases {
				t.Run(opts.Name, func(t *testing.T) {
					v := EmptyBytes(tc.Type)

					n, err := v.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
					if err != nil {
						t.Fatal(err)
					}

					if n != len(tc.Binary) {
						t.Fatalf("expected to read %d bytes, got %d", len(tc.Binary), n)
					}

					if !bytes.Equal(v.Value, tc.Value) {
						t.Fatalf("expected value %v, got %v", tc.Value, v.Value)
					}
				})
			}
		})
	}
}

func TestBytesValue_ReadValueDescending(t *testing.T) {
	for _, tc := range bytesTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := EmptyBytes(tc.Type)

			bin := make([]byte, len(tc.Binary))
			copy(bin, tc.Binary)
			bstio.ReverseBytes(bin)

			n, err := v.ReadValue(bytes.NewReader(bin), bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(bin) {
				t.Fatalf("expected to read %d bytes, got %d", len(bin), n)
			}

			if !bytes.Equal(v.Value, tc.Value) {
				t.Fatalf("expected value %v, got %v", tc.Value, v.Value)
			}
		})
	}
}

func TestBytesValue_WriteValue(t *testing.T) {
	for _, tc := range bytesTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := MustNewBytes(tc.Value, tc.Type)

			buf := bytes.Buffer{}
			n, err := v.WriteValue(&buf, bstio.ValueOptions{})
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

func TestBytesValue_WriteValueDescending(t *testing.T) {
	for _, tc := range bytesTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := MustNewBytes(tc.Value, tc.Type)

			buf := bytes.Buffer{}
			n, err := v.WriteValue(&buf, bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}

			cp := make([]byte, len(tc.Binary))
			copy(cp, tc.Binary)
			bstio.ReverseBytes(cp)

			if n != len(cp) {
				t.Fatalf("expected to write %d bytes, got %d", len(cp), n)
			}

			if !bytes.Equal(buf.Bytes(), cp) {
				t.Fatalf("expected value %v, got %v", cp, buf.Bytes())
			}
		})
	}
}

func TestBytesValue_MarshalDB(t *testing.T) {
	for _, tc := range bytesTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := MustNewBytes(tc.Value, tc.Type)

			buf, err := v.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(buf, tc.Binary) {
				t.Fatalf("expected value %v, got %v", tc.Binary, buf)
			}
		})
	}
}

func TestBytesValue_UnmarshalValue(t *testing.T) {
	for _, tc := range bytesTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := EmptyBytes(tc.Type)

			err := v.UnmarshalValue(tc.Binary, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(v.Value, tc.Value) {
				t.Fatalf("expected value %v, got %v", tc.Value, v.Value)
			}
		})
	}
}

func TestBytesValue_Skip(t *testing.T) {
	for _, tc := range bytesTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := EmptyBytes(tc.Type)

			n, err := v.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("expected to skip %d bytes, got %d", len(tc.Binary), n)
			}
		})
	}
}

// uuidBytesType represents a UUID bytes type.
// TODO: move this to test-cases.
var uuidBytesType = &bsttype.Bytes{FixedSize: 16}
