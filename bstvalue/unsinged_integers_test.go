package bstvalue

import (
	"bytes"
	"math"
	"testing"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

var uint8TestCases = []struct {
	Name   string
	Value  uint8
	Binary []byte
}{
	{
		Name:   "0",
		Value:  0,
		Binary: []byte{0x00},
	},
	{
		Name:   "1",
		Value:  1,
		Binary: []byte{0x01},
	},
	{
		Name:   "MaxUint8",
		Value:  math.MaxUint8,
		Binary: []byte{0xff},
	},
}

func TestUint8(t *testing.T) {
	vt := bsttype.Uint8()
	t.Run("Value", func(t *testing.T) {
		v := emptyUint8Value(vt)
		if v.Kind() != bsttype.KindUint8 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindUint8, v.Kind())
		}
		if v.Kind() != vt.Kind() {
			t.Fatalf("expected type %v, got %v", vt, v.Type())
		}
	})
	t.Run("Type", func(t *testing.T) {
		if vt.Kind() != bsttype.KindUint8 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindUint8, vt.Kind())
		}
	})
}

func TestUint8Value_ReadValue(t *testing.T) {
	for _, tc := range uint8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Uint8Value
			n, err := v.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if n != len(tc.Binary) {
				t.Fatalf("expected %d bytes read, got %d", len(tc.Binary), n)
			}
			if v.Value != tc.Value {
				t.Fatalf("expected %v, got %v", tc.Value, v)
			}
		})
	}
}

func TestUint8Value_ReadValueDesc(t *testing.T) {
	for _, tc := range uint8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Uint8Value
			cp := make([]byte, len(tc.Binary))
			copy(cp, tc.Binary)
			bstio.ReverseBytes(cp)
			n, err := v.ReadValue(bytes.NewReader(cp), bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			if n != len(cp) {
				t.Fatalf("expected %d bytes read, got %d", len(cp), n)
			}
			if v.Value != tc.Value {
				t.Fatalf("expected %v, got %v", tc.Value, v)
			}
		})
	}
}

func TestUint8Value_WriteValue(t *testing.T) {
	for _, tc := range uint8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := NewUint8Value(tc.Value)
			var buf bytes.Buffer
			n, err := v.WriteValue(&buf, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if n != len(tc.Binary) {
				t.Fatalf("expected %d bytes written, got %d", len(tc.Binary), n)
			}
			if !bytes.Equal(buf.Bytes(), tc.Binary) {
				t.Fatalf("expected %v, got %v", tc.Binary, buf.Bytes())
			}
		})
	}
}

func TestUint8Value_Skip(t *testing.T) {
	for _, tc := range uint8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Uint8Value
			n, err := v.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if int(n) != len(tc.Binary) {
				t.Fatalf("expected %d bytes skipped, got %d", len(tc.Binary), n)
			}
		})
	}
}

func TestUint8Value_MarshalDB(t *testing.T) {
	for _, tc := range uint8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := NewUint8Value(tc.Value)
			b, err := v.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(b, tc.Binary) {
				t.Fatalf("expected %x, got %x", tc.Binary, b)
			}
		})
	}
}

func TestUint8Value_MarshalDBDescending(t *testing.T) {
	for _, tc := range uint8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := NewUint8Value(tc.Value)
			b, err := v.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			cp := make([]byte, len(tc.Binary))
			copy(cp, tc.Binary)
			bstio.ReverseBytes(cp)

			if !bytes.Equal(b, cp) {
				t.Fatalf("expected %x, got %x", cp, b)
			}
		})
	}
}

func TestUint8Value_UnmarshalValue(t *testing.T) {
	for _, tc := range uint8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Uint8Value
			err := v.UnmarshalValue(tc.Binary, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if v.Value != tc.Value {
				t.Fatalf("expected %v, got %v", tc.Value, v)
			}
		})
	}
}

func TestUint8Value_UnmarshalValueDescending(t *testing.T) {
	for _, tc := range uint8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Uint8Value
			cp := make([]byte, len(tc.Binary))
			copy(cp, tc.Binary)
			bstio.ReverseBytes(cp)

			err := v.UnmarshalValue(cp, bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			if v.Value != tc.Value {
				t.Fatalf("expected %v, got %v", tc.Value, v)
			}
		})
	}
}

var lessThanUint8TestCases = []struct {
	Name    string
	Value1  *Uint8Value
	Value2  *Uint8Value
	Compare int
}{
	{
		Name:    "0/1",
		Value1:  &Uint8Value{Value: 0},
		Value2:  &Uint8Value{Value: 1},
		Compare: -1,
	},
	{
		Name:    "0/MaxUint8",
		Value1:  &Uint8Value{Value: 0},
		Value2:  &Uint8Value{Value: math.MaxUint8},
		Compare: -1,
	},
	{
		Name:    "1/0",
		Value1:  &Uint8Value{Value: 1},
		Value2:  &Uint8Value{Value: 0},
		Compare: 1,
	},
	{
		Name:    "1/1",
		Value1:  &Uint8Value{Value: 1},
		Value2:  &Uint8Value{Value: 1},
		Compare: 0,
	},
	{
		Name:    "1/MaxUint8",
		Value1:  &Uint8Value{Value: 1},
		Value2:  &Uint8Value{Value: math.MaxUint8},
		Compare: -1,
	},
}

func TestUint8Value_LessThan(t *testing.T) {
	for _, tc := range lessThanUint8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			res := bytes.Compare(v1, v2)
			if res != tc.Compare {
				t.Fatalf("expected %d, got %d", tc.Compare, res)
			}
		})
	}
}

func TestUint8Value_LessThanDescending(t *testing.T) {
	for _, tc := range lessThanUint8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			res := bytes.Compare(v1, v2)
			if res != -tc.Compare {
				t.Fatalf("expected %d, got %d", -tc.Compare, res)
			}
		})
	}
}

var uint16TestCases = []struct {
	Name   string
	Value  uint16
	Binary []byte
}{
	{
		Name:   "0",
		Value:  0,
		Binary: []byte{0x00, 0x00}, // 10000000 00000000
	},
	{
		Name:   "1",
		Value:  1,
		Binary: []byte{0x00, 0x01}, // 10000000 00000001
	},
	{
		Name:   "255",
		Value:  255,
		Binary: []byte{0x00, 0xFF}, // 10000000 11111111
	},
	{
		Name:   "256",
		Value:  256,
		Binary: []byte{0x01, 0x00}, // 00000001 00000000
	},
	{
		Name:   "MaxUint16",
		Value:  math.MaxUint16,
		Binary: []byte{0xFF, 0xFF}, // 11111111 11111111
	},
}

func TestUint16(t *testing.T) {
	vt := bsttype.Uint16()
	t.Run("Value", func(t *testing.T) {
		v := emptyUint16Value(vt)
		if v.Kind() != bsttype.KindUint16 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindUint16, v.Kind())
		}
		if v.Kind() != vt.Kind() {
			t.Fatalf("expected type %v, got %v", vt, v.Type())
		}
	})
	t.Run("Type", func(t *testing.T) {
		if vt.Kind() != bsttype.KindUint16 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindUint16, vt.Kind())
		}
	})
}

func TestUint16Value_ReadValue(t *testing.T) {
	for _, tc := range uint16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Uint16Value{}
			n, err := v.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %v, got %v", tc.Value, v.Value)
			}
		})
	}
}

func TestUint16Value_ReadValueDescending(t *testing.T) {
	for _, tc := range uint16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Uint16Value{}
			bin := make([]byte, len(tc.Binary))
			copy(bin, tc.Binary)
			bstio.ReverseBytes(bin)

			n, err := v.ReadValue(bytes.NewReader(bin), bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(bin) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %v, got %v", tc.Value, v.Value)
			}
		})
	}
}

func TestUint16Value_WriteValue(t *testing.T) {
	for _, tc := range uint16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Uint16Value{Value: tc.Value}
			buf := &bytes.Buffer{}
			n, err := v.WriteValue(buf, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}

			if !bytes.Equal(buf.Bytes(), tc.Binary) {
				t.Fatalf("expected %v, got %v", tc.Binary, buf.Bytes())
			}
		})
	}
}

func TestUint16Value_WriteValueDescending(t *testing.T) {
	for _, tc := range uint16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Uint16Value{Value: tc.Value}
			buf := &bytes.Buffer{}
			n, err := v.WriteValue(buf, bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}

			rev := make([]byte, len(tc.Binary))
			copy(rev, tc.Binary)
			bstio.ReverseBytes(rev)

			if !bytes.Equal(buf.Bytes(), rev) {
				t.Fatalf("expected %v, got %v", tc.Binary, buf.Bytes())
			}
		})
	}
}

func TestUint16Value_MarshalDB(t *testing.T) {
	for _, tc := range uint16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := NewUint16Value(tc.Value)
			bin, err := v.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(bin, tc.Binary) {
				t.Fatalf("expected %x, got %x", tc.Binary, bin)
			}
		})
	}
}

func TestUint16Value_UnmarshalValue(t *testing.T) {
	for _, tc := range uint16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Uint16Value
			err := v.UnmarshalValue(tc.Binary, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %v, got %v", tc.Value, v.Value)
			}
		})
	}
}

func TestUint16Value_Skip(t *testing.T) {
	for _, tc := range uint16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Uint16Value{}
			n, err := v.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}
		})
	}
}

var lessThanUint16TestCases = []struct {
	Name    string
	Value1  *Uint16Value
	Value2  *Uint16Value
	Compare int
}{
	{
		Name:    "0/1",
		Value1:  &Uint16Value{Value: 0},
		Value2:  &Uint16Value{Value: 1},
		Compare: -1,
	},
	{
		Name:    "1/0",
		Value1:  &Uint16Value{Value: 1},
		Value2:  &Uint16Value{Value: 0},
		Compare: 1,
	},
	{
		Name:    "1/1",
		Value1:  &Uint16Value{Value: 1},
		Value2:  &Uint16Value{Value: 1},
		Compare: 0,
	},
	{
		Name:    "0/0",
		Value1:  &Uint16Value{Value: 0},
		Value2:  &Uint16Value{Value: 0},
		Compare: 0,
	},
	{
		Name:    "MaxUint16/0",
		Value1:  &Uint16Value{Value: math.MaxUint16},
		Value2:  &Uint16Value{Value: 0},
		Compare: 1,
	},
	{
		Name:    "0/MaxUint16",
		Value1:  &Uint16Value{Value: 0},
		Value2:  &Uint16Value{Value: math.MaxUint16},
		Compare: -1,
	},
}

func TestUint16Value_Less(t *testing.T) {
	for _, tc := range lessThanUint16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(v1, v2) != tc.Compare {
				t.Fatalf("expected %d, got %d", tc.Compare, bytes.Compare(v1, v2))
			}
		})
	}
}

func TestUint16Value_LessDescending(t *testing.T) {
	for _, tc := range lessThanUint16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(v1, v2) != -tc.Compare {
				t.Fatalf("expected %d, got %d", -tc.Compare, bytes.Compare(v1, v2))
			}
		})
	}
}

var uint32ValueTestCases = []struct {
	Name   string
	Value  uint32
	Binary []byte
}{
	{
		Name:   "0",
		Value:  0,
		Binary: []byte{0x00, 0x00, 0x00, 0x00},
	},
	{
		Name:   "1",
		Value:  1,
		Binary: []byte{0x00, 0x00, 0x00, 0x01},
	},
	{
		Name:   "MaxUint32",
		Value:  math.MaxUint32,
		Binary: []byte{0xFF, 0xFF, 0xFF, 0xFF},
	},
}

func TestUint32(t *testing.T) {
	vt := bsttype.Uint32()
	t.Run("Value", func(t *testing.T) {
		v := emptyUint32Value(vt)
		if v.Kind() != bsttype.KindUint32 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindUint32, v.Kind())
		}
		if v.Kind() != vt.Kind() {
			t.Fatalf("expected type %v, got %v", vt, v.Type())
		}
	})
	t.Run("Type", func(t *testing.T) {
		if vt.Kind() != bsttype.KindUint32 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindUint32, vt.Kind())
		}
	})
}

func TestUint32Value_ReadValue(t *testing.T) {
	for _, tc := range uint32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Uint32Value
			n, err := v.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if n != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %v, got %v", tc.Value, v.Value)
			}
		})
	}
}

func TestUint32Value_ReadValueDescending(t *testing.T) {
	for _, tc := range uint32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Uint32Value
			bin := make([]byte, len(tc.Binary))
			copy(bin, tc.Binary)
			bstio.ReverseBytes(bin)

			n, err := v.ReadValue(bytes.NewReader(bin), bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(bin) {
				t.Fatalf("expected %d, got %d", len(bin), n)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %v, got %v", tc.Value, v.Value)
			}
		})
	}
}

func TestUint32Value_WriteValue(t *testing.T) {
	for _, tc := range uint32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Uint32Value
			v.Value = tc.Value

			buf := &bytes.Buffer{}
			n, err := v.WriteValue(buf, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if n != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}

			if !bytes.Equal(buf.Bytes(), tc.Binary) {
				t.Fatalf("expected %v, got %v", tc.Binary, buf.Bytes())
			}
		})
	}
}

func TestUint32Value_WriteValueDescending(t *testing.T) {
	for _, tc := range uint32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Uint32Value
			v.Value = tc.Value

			buf := &bytes.Buffer{}
			n, err := v.WriteValue(buf, bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}

			bin := make([]byte, len(tc.Binary))
			copy(bin, tc.Binary)
			bstio.ReverseBytes(bin)

			if n != len(bin) {
				t.Fatalf("expected %d, got %d", len(bin), n)
			}

			if !bytes.Equal(buf.Bytes(), bin) {
				t.Fatalf("expected %v, got %v", bin, buf.Bytes())
			}
		})
	}
}

func TestUint32Value_MarshalDB(t *testing.T) {
	for _, tc := range uint32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Uint32Value
			v.Value = tc.Value

			bin, err := v.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(bin, tc.Binary) {
				t.Fatalf("expected %v, got %v", tc.Binary, bin)
			}
		})
	}
}

func TestUint32Value_UnmarshalValue(t *testing.T) {
	for _, tc := range uint32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Uint32Value
			err := v.UnmarshalValue(tc.Binary, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if v.Value != tc.Value {
				t.Fatalf("expected %v, got %v", tc.Value, v.Value)
			}
		})
	}
}

func TestUint32Value_Skip(t *testing.T) {
	for _, tc := range uint32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Uint32Value
			n, err := v.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if int(n) != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}
		})
	}
}

var lessThanUint32TestCases = []struct {
	Name    string
	Value1  *Uint32Value
	Value2  *Uint32Value
	Compare int
}{
	{
		Name:    "0/1",
		Value1:  &Uint32Value{Value: 0},
		Value2:  &Uint32Value{Value: 1},
		Compare: -1,
	},
	{
		Name:    "1/0",
		Value1:  &Uint32Value{Value: 1},
		Value2:  &Uint32Value{Value: 0},
		Compare: 1,
	},
	{
		Name:    "1/1",
		Value1:  &Uint32Value{Value: 1},
		Value2:  &Uint32Value{Value: 1},
		Compare: 0,
	},
	{
		Name:    "0/0",
		Value1:  &Uint32Value{Value: 0},
		Value2:  &Uint32Value{Value: 0},
		Compare: 0,
	},
	{
		Name:    "MaxUint32/0",
		Value1:  &Uint32Value{Value: math.MaxUint32},
		Value2:  &Uint32Value{Value: 0},
		Compare: 1,
	},
	{
		Name:    "0/MaxUint32",
		Value1:  &Uint32Value{Value: 0},
		Value2:  &Uint32Value{Value: math.MaxUint32},
		Compare: -1,
	},
}

func TestUint32Value_Less(t *testing.T) {
	for _, tc := range lessThanUint32TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(v1, v2) != tc.Compare {
				t.Fatalf("expected %d, got %d", tc.Compare, bytes.Compare(v1, v2))
			}
		})
	}
}

func TestUint32Value_LessDescending(t *testing.T) {
	for _, tc := range lessThanUint32TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(v1, v2) != -tc.Compare {
				t.Fatalf("expected %d, got %d", -tc.Compare, bytes.Compare(v1, v2))
			}
		})
	}
}

var uint64ValueTestCases = []struct {
	Name   string
	Value  uint64
	Binary []byte
}{
	{
		Name:   "0",
		Value:  0,
		Binary: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	},
	{
		Name:   "1",
		Value:  1,
		Binary: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
	},
	{
		Name:   "MaxUint64",
		Value:  math.MaxUint64,
		Binary: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	},
}

func TestUint64(t *testing.T) {
	vt := bsttype.Uint64()
	t.Run("Value", func(t *testing.T) {
		v := emptyUint64Value(vt)
		if v.Kind() != bsttype.KindUint64 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindUint64, v.Kind())
		}
		if v.Kind() != vt.Kind() {
			t.Fatalf("expected type %v, got %v", vt, v.Type())
		}
	})
	t.Run("Type", func(t *testing.T) {
		if vt.Kind() != bsttype.KindUint64 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindUint64, vt.Kind())
		}
	})
}

func TestUint64Value_ReadValue(t *testing.T) {
	for _, tc := range uint64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Uint64Value{}
			n, err := v.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %d, got %d", tc.Value, v.Value)
			}
		})
	}
}

func TestUint64Value_ReadValueDescending(t *testing.T) {
	for _, tc := range uint64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Uint64Value{}
			cp := make([]byte, len(tc.Binary))
			copy(cp, tc.Binary)
			bstio.ReverseBytes(cp)

			n, err := v.ReadValue(bytes.NewReader(cp), bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(cp) {
				t.Fatalf("expected %d, got %d", len(cp), n)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %d, got %d", tc.Value, v.Value)
			}
		})
	}
}

func TestUint64Value_WriteValue(t *testing.T) {
	for _, tc := range uint64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := Uint64Value{Value: tc.Value}
			b := &bytes.Buffer{}
			n, err := v.WriteValue(b, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}

			if !bytes.Equal(b.Bytes(), tc.Binary) {
				t.Fatalf("expected %v, got %v", tc.Binary, b.Bytes())
			}
		})
	}
}

func TestUint64Value_WriteValueDescending(t *testing.T) {
	for _, tc := range uint64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := Uint64Value{Value: tc.Value}
			b := &bytes.Buffer{}
			n, err := v.WriteValue(b, bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}

			rev := make([]byte, len(tc.Binary))
			copy(rev, tc.Binary)
			bstio.ReverseBytes(rev)

			if n != len(rev) {
				t.Fatalf("expected %d, got %d", len(rev), n)
			}

			if !bytes.Equal(b.Bytes(), rev) {
				t.Fatalf("expected %v, got %v", rev, b.Bytes())
			}
		})
	}
}

func TestUint64Value_MarshalDB(t *testing.T) {
	for _, tc := range uint64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Uint64Value{Value: tc.Value}
			b, err := v.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(b, tc.Binary) {
				t.Fatalf("expected %v, got %v", tc.Binary, b)
			}
		})
	}
}

func TestUint64Value_UnmarshalValue(t *testing.T) {
	for _, tc := range uint64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Uint64Value{}
			if err := v.UnmarshalValue(tc.Binary, bstio.ValueOptions{}); err != nil {
				t.Fatal(err)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %d, got %d", tc.Value, v.Value)
			}
		})
	}
}

func TestUint64Value_Skip(t *testing.T) {
	for _, tc := range uint64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Uint64Value{}
			n, err := v.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}
		})
	}
}

var lessThanUint64TestCases = []struct {
	Name    string
	Value1  *Uint64Value
	Value2  *Uint64Value
	Compare int
}{
	{
		Name:    "0/1",
		Value1:  &Uint64Value{Value: 0},
		Value2:  &Uint64Value{Value: 1},
		Compare: -1,
	},
	{
		Name:    "1/0",
		Value1:  &Uint64Value{Value: 1},
		Value2:  &Uint64Value{Value: 0},
		Compare: 1,
	},
	{
		Name:    "1/1",
		Value1:  &Uint64Value{Value: 1},
		Value2:  &Uint64Value{Value: 1},
		Compare: 0,
	},
	{
		Name:    "0/0",
		Value1:  &Uint64Value{Value: 0},
		Value2:  &Uint64Value{Value: 0},
		Compare: 0,
	},
	{
		Name:    "MaxUint64/0",
		Value1:  &Uint64Value{Value: math.MaxUint64},
		Value2:  &Uint64Value{Value: 0},
		Compare: 1,
	},
	{
		Name:    "0/MaxUint64",
		Value1:  &Uint64Value{Value: 0},
		Value2:  &Uint64Value{Value: math.MaxUint64},
		Compare: -1,
	},
}

func TestUint64Value_Less(t *testing.T) {
	for _, tc := range lessThanUint64TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(v1, v2) != tc.Compare {
				t.Fatalf("expected %d, got %d", tc.Compare, bytes.Compare(v1, v2))
			}
		})
	}
}

func TestUint64Value_LessDescending(t *testing.T) {
	for _, tc := range lessThanUint64TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(v1, v2) != -tc.Compare {
				t.Fatalf("expected %d, got %d", -tc.Compare, bytes.Compare(v1, v2))
			}
		})
	}
}

var uintValueTestCases = []struct {
	Name   string
	Value  uint
	Binary []byte
}{
	{
		Name:   "0",
		Value:  0,
		Binary: []byte{0x00},
	},
	{
		Name:   "1",
		Value:  1,
		Binary: []byte{0x01, 0x01},
	},
	{
		Name:   "MaxUint8",
		Value:  math.MaxUint8,
		Binary: []byte{0x01, 0xff},
	},
	{
		Name:   "MaxUint16",
		Value:  math.MaxUint16,
		Binary: []byte{0x02, 0xff, 0xff},
	},
	{
		Name:   "MaxUint24",
		Value:  math.MaxUint16<<8 | math.MaxUint8,
		Binary: []byte{0x03, 0xff, 0xff, 0xff},
	},
	{
		Name:   "MaxUint32",
		Value:  math.MaxUint32,
		Binary: []byte{0x04, 0xff, 0xff, 0xff, 0xff},
	},
	{
		Name:   "MaxUint64",
		Value:  math.MaxUint64,
		Binary: []byte{0x08, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
	},
}

func TestUint(t *testing.T) {
	vt := bsttype.Uint()
	t.Run("Value", func(t *testing.T) {
		v := emptyUintValue(vt)
		if v.Kind() != bsttype.KindUint {
			t.Fatalf("expected kind %d, got %d", bsttype.KindUint, v.Kind())
		}
		if v.Kind() != vt.Kind() {
			t.Fatalf("expected type %v, got %v", vt, v.Type())
		}
	})
	t.Run("Type", func(t *testing.T) {
		if vt.Kind() != bsttype.KindUint {
			t.Fatalf("expected kind %d, got %d", bsttype.KindUint, vt.Kind())
		}
	})
}

func TestUintValue_ReadValue(t *testing.T) {
	for _, tc := range uintValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &UintValue{}
			n, err := v.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %d, got %d", tc.Value, v.Value)
			}
		})
	}
}

func TestUintValue_ReadValueDescending(t *testing.T) {
	for _, tc := range uintValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &UintValue{}
			cp := make([]byte, len(tc.Binary))
			copy(cp, tc.Binary)
			bstio.ReverseBytes(cp)

			n, err := v.ReadValue(bytes.NewReader(cp), bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(cp) {
				t.Fatalf("expected %d, got %d", len(cp), n)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %d, got %d", tc.Value, v.Value)
			}
		})
	}
}

func TestUintValue_WriteValue(t *testing.T) {
	for _, tc := range uintValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := UintValue{Value: tc.Value}
			b := &bytes.Buffer{}
			n, err := v.WriteValue(b, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}

			if !bytes.Equal(b.Bytes(), tc.Binary) {
				t.Fatalf("expected %v, got %v", tc.Binary, b.Bytes())
			}
		})
	}
}

func TestUintValue_WriteValueDescending(t *testing.T) {
	for _, tc := range uintValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := UintValue{Value: tc.Value}
			b := &bytes.Buffer{}

			n, err := v.WriteValue(b, bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}

			rev := make([]byte, len(tc.Binary))
			copy(rev, tc.Binary)
			bstio.ReverseBytes(rev)

			if n != len(rev) {
				t.Fatalf("expected %d, got %d", len(rev), n)
			}

			if !bytes.Equal(b.Bytes(), rev) {
				t.Fatalf("expected %v, got %v", rev, b.Bytes())
			}
		})
	}
}

func TestUintValue_MarshalDB(t *testing.T) {
	for _, tc := range uintValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &UintValue{Value: tc.Value}
			b, err := v.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(b, tc.Binary) {
				t.Fatalf("expected %v, got %v", tc.Binary, b)
			}
		})
	}
}

func TestUintValue_UnmarshalValue(t *testing.T) {
	for _, tc := range uintValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &UintValue{}
			if err := v.UnmarshalValue(tc.Binary, bstio.ValueOptions{}); err != nil {
				t.Fatal(err)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %d, got %d", tc.Value, v.Value)
			}
		})
	}
}

func TestUintValue_Skip(t *testing.T) {
	for _, tc := range uintValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &UintValue{}
			n, err := v.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}
		})
	}
}

var lessThanUintTestCases = []struct {
	Name    string
	Value1  *UintValue
	Value2  *UintValue
	Compare int
}{
	{
		Name:    "0/1",
		Value1:  &UintValue{Value: 0},
		Value2:  &UintValue{Value: 1},
		Compare: -1,
	},
	{
		Name:    "1/0",
		Value1:  &UintValue{Value: 1},
		Value2:  &UintValue{Value: 0},
		Compare: 1,
	},
	{
		Name:    "1/1",
		Value1:  &UintValue{Value: 1},
		Value2:  &UintValue{Value: 1},
		Compare: 0,
	},
	{
		Name:    "0/0",
		Value1:  &UintValue{Value: 0},
		Value2:  &UintValue{Value: 0},
		Compare: 0,
	},
	{
		Name:    "MaxUint64/0",
		Value1:  &UintValue{Value: math.MaxUint64},
		Value2:  &UintValue{Value: 0},
		Compare: 1,
	},
	{
		Name:    "0/MaxUint64",
		Value1:  &UintValue{Value: 0},
		Value2:  &UintValue{Value: math.MaxUint64},
		Compare: -1,
	},
}

func TestUintValue_Less(t *testing.T) {
	for _, tc := range lessThanUintTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(v1, v2) != tc.Compare {
				t.Fatalf("expected %d, got %d", tc.Compare, bytes.Compare(v1, v2))
			}
		})
	}
}

func TestUintValue_LessDescending(t *testing.T) {
	for _, tc := range lessThanUint64TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(v1, v2) != -tc.Compare {
				t.Fatalf("expected %d, got %d", -tc.Compare, bytes.Compare(v1, v2))
			}
		})
	}
}
