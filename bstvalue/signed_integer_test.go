package bstvalue

import (
	"bytes"
	"math"
	"testing"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

var int8TestCases = []struct {
	Name   string
	Value  int8
	Binary []byte
}{
	{
		Name:   "0",
		Value:  0,
		Binary: []byte{0x00 | bstio.PositiveBit8Mask},
	},
	{
		Name:   "1",
		Value:  1,
		Binary: []byte{0x01 | bstio.PositiveBit8Mask},
	},
	{
		Name:   "-1",
		Value:  -1,
		Binary: []byte{(0x00 | bstio.PositiveBit8Mask) - 1},
	},
	{
		Name:   "MaxInt8",
		Value:  math.MaxInt8,
		Binary: []byte{0x7f | bstio.PositiveBit8Mask},
	},
	{
		Name:   "MinInt8",
		Value:  math.MinInt8,
		Binary: []byte{0x00},
	},
}

func TestInt8(t *testing.T) {
	vt := bsttype.Int8()
	t.Run("Value", func(t *testing.T) {
		v := emptyInt8Value(vt)
		if v.Kind() != bsttype.KindInt8 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindInt8, v.Kind())
		}
		if v.Kind() != vt.Kind() {
			t.Fatalf("expected type %v, got %v", vt, v.Type())
		}
	})
	t.Run("Type", func(t *testing.T) {
		if vt.Kind() != bsttype.KindInt8 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindInt8, vt.Kind())
		}
	})
}

func TestInt8Value_ReadValue(t *testing.T) {
	for _, tc := range int8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Int8Value
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

func TestInt8Value_ReadValueDesc(t *testing.T) {
	for _, tc := range int8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Int8Value
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

func TestInt8Value_WriteValue(t *testing.T) {
	for _, tc := range int8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := NewInt8Value(tc.Value)
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

func TestInt8Value_Skip(t *testing.T) {
	for _, tc := range int8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Int8Value
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

func TestInt8Value_MarshalDB(t *testing.T) {
	for _, tc := range int8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := NewInt8Value(tc.Value)
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

func TestInt8Value_MarshalDBDescending(t *testing.T) {
	for _, tc := range int8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := NewInt8Value(tc.Value)
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

func TestInt8Value_UnmarshalValue(t *testing.T) {
	for _, tc := range int8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Int8Value
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

func TestInt8Value_UnmarshalValueDescending(t *testing.T) {
	for _, tc := range int8TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Int8Value
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

var lessThanInt8TestCases = []struct {
	Name    string
	Value1  *Int8Value
	Value2  *Int8Value
	Compare int
}{
	{
		Name:    "0/1",
		Value1:  &Int8Value{Value: 0},
		Value2:  &Int8Value{Value: 1},
		Compare: -1,
	},
	{
		Name:    "-1/0",
		Value1:  &Int8Value{Value: -1},
		Value2:  &Int8Value{Value: 0},
		Compare: -1,
	},
	{
		Name:    "-1/1",
		Value1:  &Int8Value{Value: -1},
		Value2:  &Int8Value{Value: 1},
		Compare: -1,
	},
	{
		Name:    "1/-1",
		Value1:  &Int8Value{Value: 1},
		Value2:  &Int8Value{Value: -1},
		Compare: 1,
	},
	{
		Name:    "127/-127",
		Value1:  &Int8Value{Value: 127},
		Value2:  &Int8Value{Value: -127},
		Compare: 1,
	},
}

func TestInt8Value_LessThan(t *testing.T) {
	for _, tc := range lessThanInt8TestCases {
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

func TestInt8Value_LessThanDescending(t *testing.T) {
	for _, tc := range lessThanInt8TestCases {
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

var int16TestCases = []struct {
	Name   string
	Value  int16
	Binary []byte
}{
	{
		Name:   "0",
		Value:  0,
		Binary: []byte{0x00 | bstio.PositiveBit8Mask, 0x00}, // 10000000 00000000
	},
	{
		Name:   "1",
		Value:  1,
		Binary: []byte{0x00 | bstio.PositiveBit8Mask, 0x01}, // 10000000 00000001
	},
	{
		Name:   "-1",
		Value:  -1,
		Binary: []byte{0x7F, 0xFF}, // 01111111 11111111
	},
	{
		Name:   "255",
		Value:  255,
		Binary: []byte{0x00 | bstio.PositiveBit8Mask, 0xFF}, // 10000000 11111111
	},
	{
		Name:   "256",
		Value:  256,
		Binary: []byte{0x01 | bstio.PositiveBit8Mask, 0x00}, // 10000001 00000000
	},
	{
		Name:   "-255",
		Value:  -255,
		Binary: []byte{0x7F, 0x01}, // 00000000 11111111
	},
	{
		Name:   "-256",
		Value:  -256,
		Binary: []byte{0x7F, 0x00}, // 00000000 00000000
	},
	{
		Name:   "MaxInt16",
		Value:  math.MaxInt16,
		Binary: []byte{0x7F | bstio.PositiveBit8Mask, 0xFF}, // 11111111 11111111
	},
	{
		Name:   "MinInt16",
		Value:  math.MinInt16,
		Binary: []byte{0x00, 0x00}, // 00000000 00000000
	},
}

func TestInt16(t *testing.T) {
	vt := bsttype.Int16()
	t.Run("Value", func(t *testing.T) {
		v := emptyInt16Value(vt)
		if v.Kind() != bsttype.KindInt16 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindInt16, v.Kind())
		}
		if v.Kind() != vt.Kind() {
			t.Fatalf("expected type %v, got %v", vt, v.Type())
		}
	})
	t.Run("Type", func(t *testing.T) {
		if vt.Kind() != bsttype.KindInt16 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindInt16, vt.Kind())
		}
	})
}

func TestInt16Value_ReadValue(t *testing.T) {
	for _, tc := range int16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Int16Value{}
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

func TestInt16Value_ReadValueDescending(t *testing.T) {
	for _, tc := range int16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Int16Value{}
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

func TestInt16Value_WriteValue(t *testing.T) {
	for _, tc := range int16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Int16Value{Value: tc.Value}
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

func TestInt16Value_WriteValueDescending(t *testing.T) {
	for _, tc := range int16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Int16Value{Value: tc.Value}
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

func TestInt16Value_MarshalDB(t *testing.T) {
	for _, tc := range int16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := NewInt16Value(tc.Value)
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

func TestInt16Value_UnmarshalValue(t *testing.T) {
	for _, tc := range int16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Int16Value
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

func TestInt16Value_Skip(t *testing.T) {
	for _, tc := range int16TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Int16Value{}
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

var lessThanInt16TestCases = []struct {
	Name    string
	Value1  *Int16Value
	Value2  *Int16Value
	Compare int
}{
	{
		Name:    "0/1",
		Value1:  &Int16Value{Value: 0},
		Value2:  &Int16Value{Value: 1},
		Compare: -1,
	},
	{
		Name:    "-1/0",
		Value1:  &Int16Value{Value: -1},
		Value2:  &Int16Value{Value: 0},
		Compare: -1,
	},
	{
		Name:    "-1/1",
		Value1:  &Int16Value{Value: -1},
		Value2:  &Int16Value{Value: 1},
		Compare: -1,
	},
	{
		Name:    "1/-1",
		Value1:  &Int16Value{Value: 1},
		Value2:  &Int16Value{Value: -1},
		Compare: 1,
	},
	{
		Name:    "MaxInt16/MinInt16",
		Value1:  &Int16Value{Value: math.MaxInt16},
		Value2:  &Int16Value{Value: math.MinInt16},
		Compare: 1,
	},
	{
		Name:    "MinInt16/MaxInt16",
		Value1:  &Int16Value{Value: math.MinInt16},
		Value2:  &Int16Value{Value: math.MaxInt16},
		Compare: -1,
	},
	{
		Name:    "0/0",
		Value1:  &Int16Value{Value: 0},
		Value2:  &Int16Value{Value: 0},
		Compare: 0,
	},
	{
		Name:    "1/1",
		Value1:  &Int16Value{Value: 1},
		Value2:  &Int16Value{Value: 1},
		Compare: 0,
	},
	{
		Name:   "-1/-1",
		Value1: &Int16Value{Value: -1},
		Value2: &Int16Value{Value: -1},
	},
}

func TestInt16Value_Less(t *testing.T) {
	for _, tc := range lessThanInt16TestCases {
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

func TestInt16Value_LessDescending(t *testing.T) {
	for _, tc := range lessThanInt16TestCases {
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

var int32ValueTestCases = []struct {
	Name   string
	Value  int32
	Binary []byte
}{
	{
		Name:   "0",
		Value:  0,
		Binary: []byte{0x00 | bstio.PositiveBit8Mask, 0x00, 0x00, 0x00},
	},
	{
		Name:   "1",
		Value:  1,
		Binary: []byte{0x00 | bstio.PositiveBit8Mask, 0x00, 0x00, 0x01},
	},
	{
		Name:   "-1",
		Value:  -1,
		Binary: []byte{0x7F, 0xFF, 0xFF, 0xFF},
	},
	{
		Name:   "MaxInt32",
		Value:  math.MaxInt32,
		Binary: []byte{0x7F | bstio.PositiveBit8Mask, 0xFF, 0xFF, 0xFF},
	},
	{
		Name:   "MinInt32",
		Value:  math.MinInt32,
		Binary: []byte{0x00, 0x00, 0x00, 0x00},
	},
}

func TestInt32(t *testing.T) {
	vt := bsttype.Int32()
	t.Run("Value", func(t *testing.T) {
		v := emptyInt32Value(vt)
		if v.Kind() != bsttype.KindInt32 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindInt32, v.Kind())
		}
		if v.Kind() != vt.Kind() {
			t.Fatalf("expected type %v, got %v", vt, v.Type())
		}
	})
	t.Run("Type", func(t *testing.T) {
		if vt.Kind() != bsttype.KindInt32 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindInt32, vt.Kind())
		}
	})
}

func TestInt32Value_ReadValue(t *testing.T) {
	for _, tc := range int32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Int32Value
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

func TestInt32Value_ReadValueDescending(t *testing.T) {
	for _, tc := range int32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Int32Value
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

func TestInt32Value_WriteValue(t *testing.T) {
	for _, tc := range int32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Int32Value
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

func TestInt32Value_WriteValueDescending(t *testing.T) {
	for _, tc := range int32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Int32Value
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

func TestInt32Value_MarshalDB(t *testing.T) {
	for _, tc := range int32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Int32Value
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

func TestInt32Value_UnmarshalValue(t *testing.T) {
	for _, tc := range int32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Int32Value
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

func TestInt32Value_Skip(t *testing.T) {
	for _, tc := range int32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v Int32Value
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

var lessThanInt32TestCases = []struct {
	Name    string
	Value1  *Int32Value
	Value2  *Int32Value
	Compare int
}{
	{
		Name:    "0/1",
		Value1:  &Int32Value{Value: 0},
		Value2:  &Int32Value{Value: 1},
		Compare: -1,
	},
	{
		Name:    "-1/0",
		Value1:  &Int32Value{Value: -1},
		Value2:  &Int32Value{Value: 0},
		Compare: -1,
	},
	{
		Name:    "-1/1",
		Value1:  &Int32Value{Value: -1},
		Value2:  &Int32Value{Value: 1},
		Compare: -1,
	},
	{
		Name:    "1/-1",
		Value1:  &Int32Value{Value: 1},
		Value2:  &Int32Value{Value: -1},
		Compare: 1,
	},
	{
		Name:    "MaxInt32/MinInt32",
		Value1:  &Int32Value{Value: math.MaxInt32},
		Value2:  &Int32Value{Value: math.MinInt32},
		Compare: 1,
	},
	{
		Name:    "MinInt32/MaxInt32",
		Value1:  &Int32Value{Value: math.MinInt32},
		Value2:  &Int32Value{Value: math.MaxInt32},
		Compare: -1,
	},
	{
		Name:    "0/0",
		Value1:  &Int32Value{Value: 0},
		Value2:  &Int32Value{Value: 0},
		Compare: 0,
	},
	{
		Name:    "1/1",
		Value1:  &Int32Value{Value: 1},
		Value2:  &Int32Value{Value: 1},
		Compare: 0,
	},
	{
		Name:   "-1/-1",
		Value1: &Int32Value{Value: -1},
		Value2: &Int32Value{Value: -1},
	},
}

func TestInt32Value_Less(t *testing.T) {
	for _, tc := range lessThanInt32TestCases {
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

func TestInt32Value_LessDescending(t *testing.T) {
	for _, tc := range lessThanInt32TestCases {
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

var int64ValueTestCases = []struct {
	Name   string
	Value  int64
	Binary []byte
}{
	{
		Name:   "0",
		Value:  0,
		Binary: []byte{0x00 | bstio.PositiveBit8Mask, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	},
	{
		Name:   "1",
		Value:  1,
		Binary: []byte{0x00 | bstio.PositiveBit8Mask, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
	},
	{
		Name:   "-1",
		Value:  -1,
		Binary: []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	},
	{
		Name:   "MaxInt64",
		Value:  math.MaxInt64,
		Binary: []byte{0x7F | bstio.PositiveBit8Mask, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	},
	{
		Name:   "MinInt64",
		Value:  math.MinInt64,
		Binary: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	},
}

func TestInt64(t *testing.T) {
	vt := bsttype.Int64()
	t.Run("Value", func(t *testing.T) {
		v := emptyInt64Value(vt)
		if v.Kind() != bsttype.KindInt64 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindInt64, v.Kind())
		}
		if v.Kind() != vt.Kind() {
			t.Fatalf("expected type %v, got %v", vt, v.Type())
		}
	})
	t.Run("Type", func(t *testing.T) {
		if vt.Kind() != bsttype.KindInt64 {
			t.Fatalf("expected kind %d, got %d", bsttype.KindInt64, vt.Kind())
		}
	})
}

func TestInt64Value_ReadValue(t *testing.T) {
	for _, tc := range int64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Int64Value{}
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

func TestInt64Value_ReadValueDescending(t *testing.T) {
	for _, tc := range int64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Int64Value{}
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

func TestInt64Value_WriteValue(t *testing.T) {
	for _, tc := range int64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := Int64Value{Value: tc.Value}
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

func TestInt64Value_WriteValueDescending(t *testing.T) {
	for _, tc := range int64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := Int64Value{Value: tc.Value}
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

func TestInt64Value_MarshalDB(t *testing.T) {
	for _, tc := range int64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Int64Value{Value: tc.Value}
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

func TestInt64Value_UnmarshalValue(t *testing.T) {
	for _, tc := range int64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Int64Value{}
			if err := v.UnmarshalValue(tc.Binary, bstio.ValueOptions{}); err != nil {
				t.Fatal(err)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %d, got %d", tc.Value, v.Value)
			}
		})
	}
}

func TestInt64Value_Skip(t *testing.T) {
	for _, tc := range int64ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &Int64Value{}
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

var lessThanInt64TestCases = []struct {
	Name    string
	Value1  *Int64Value
	Value2  *Int64Value
	Compare int
}{
	{
		Name:    "0/1",
		Value1:  &Int64Value{Value: 0},
		Value2:  &Int64Value{Value: 1},
		Compare: -1,
	},
	{
		Name:    "-1/0",
		Value1:  &Int64Value{Value: -1},
		Value2:  &Int64Value{Value: 0},
		Compare: -1,
	},
	{
		Name:    "-1/1",
		Value1:  &Int64Value{Value: -1},
		Value2:  &Int64Value{Value: 1},
		Compare: -1,
	},
	{
		Name:    "1/-1",
		Value1:  &Int64Value{Value: 1},
		Value2:  &Int64Value{Value: -1},
		Compare: 1,
	},
	{
		Name:    "MaxInt64/MinInt64",
		Value1:  &Int64Value{Value: math.MaxInt64},
		Value2:  &Int64Value{Value: math.MinInt64},
		Compare: 1,
	},
	{
		Name:    "MinInt64/MaxInt64",
		Value1:  &Int64Value{Value: math.MinInt64},
		Value2:  &Int64Value{Value: math.MaxInt64},
		Compare: -1,
	},
	{
		Name:    "0/0",
		Value1:  &Int64Value{Value: 0},
		Value2:  &Int64Value{Value: 0},
		Compare: 0,
	},
	{
		Name:    "1/1",
		Value1:  &Int64Value{Value: 1},
		Value2:  &Int64Value{Value: 1},
		Compare: 0,
	},
	{
		Name:   "-1/-1",
		Value1: &Int64Value{Value: -1},
		Value2: &Int64Value{Value: -1},
	},
}

func TestInt64Value_Less(t *testing.T) {
	for _, tc := range lessThanInt64TestCases {
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

func TestInt64Value_LessDescending(t *testing.T) {
	for _, tc := range lessThanInt64TestCases {
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

var intComparableTestCases = []struct {
	Name   string
	Value  int
	Binary []byte
}{
	{
		Name:   "0",
		Value:  0,
		Binary: []byte{0x00 | bstio.PositiveBit8Mask, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	},
	{
		Name:   "1",
		Value:  1,
		Binary: []byte{0x00 | bstio.PositiveBit8Mask, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
	},
	{
		Name:   "-1",
		Value:  -1,
		Binary: []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	},
	{
		Name:   "MaxInt64",
		Value:  math.MaxInt64,
		Binary: []byte{0x7F | bstio.PositiveBit8Mask, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	},
	{
		Name:   "MinInt64",
		Value:  math.MinInt64,
		Binary: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	},
}

func TestInt(t *testing.T) {
	vt := bsttype.Int()
	t.Run("Value", func(t *testing.T) {
		v := emptyIntValue(vt)
		if v.Kind() != bsttype.KindInt {
			t.Fatalf("expected kind %d, got %d", bsttype.KindInt, v.Kind())
		}
		if v.Kind() != vt.Kind() {
			t.Fatalf("expected type %v, got %v", vt, v.Type())
		}
	})
	t.Run("Type", func(t *testing.T) {
		if vt.Kind() != bsttype.KindInt {
			t.Fatalf("expected kind %d, got %d", bsttype.KindInt, vt.Kind())
		}
	})
}

func TestIntComparable_ReadValue(t *testing.T) {
	for _, tc := range intComparableTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &IntValue{}
			n, err := v.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{Comparable: true})
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

func TestIntComparable_ReadValueDescending(t *testing.T) {
	for _, tc := range intComparableTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &IntValue{}
			cp := make([]byte, len(tc.Binary))
			copy(cp, tc.Binary)
			bstio.ReverseBytes(cp)

			n, err := v.ReadValue(bytes.NewReader(cp), bstio.ValueOptions{Descending: true, Comparable: true})
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

func TestIntComparable_WriteValue(t *testing.T) {
	for _, tc := range intComparableTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := IntValue{Value: tc.Value}
			b := &bytes.Buffer{}
			n, err := v.WriteValue(b, bstio.ValueOptions{Comparable: true})
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

func TestIntComparable_WriteValueDescending(t *testing.T) {
	for _, tc := range intComparableTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := IntValue{Value: tc.Value}
			b := &bytes.Buffer{}
			n, err := v.WriteValue(b, bstio.ValueOptions{Descending: true, Comparable: true})
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

func TestIntComparable_MarshalDB(t *testing.T) {
	for _, tc := range intComparableTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &IntValue{Value: tc.Value}
			b, err := v.MarshalValue(bstio.ValueOptions{Comparable: true})
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(b, tc.Binary) {
				t.Fatalf("expected %v, got %v", tc.Binary, b)
			}
		})
	}
}

func TestIntValue_UnmarshalValue(t *testing.T) {
	for _, tc := range intComparableTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &IntValue{}
			if err := v.UnmarshalValue(tc.Binary, bstio.ValueOptions{Comparable: true}); err != nil {
				t.Fatal(err)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %d, got %d", tc.Value, v.Value)
			}
		})
	}
}

func TestIntComparable_Skip(t *testing.T) {
	for _, tc := range intComparableTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &IntValue{}
			n, err := v.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{Comparable: true})
			if err != nil {
				t.Fatal(err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}
		})
	}
}

var lessThanIntTestCases = []struct {
	Name    string
	Value1  *IntValue
	Value2  *IntValue
	Compare int
}{
	{
		Name:    "0/1",
		Value1:  &IntValue{Value: 0},
		Value2:  &IntValue{Value: 1},
		Compare: -1,
	},
	{
		Name:    "-1/0",
		Value1:  &IntValue{Value: -1},
		Value2:  &IntValue{Value: 0},
		Compare: -1,
	},
	{
		Name:    "-1/1",
		Value1:  &IntValue{Value: -1},
		Value2:  &IntValue{Value: 1},
		Compare: -1,
	},
	{
		Name:    "1/-1",
		Value1:  &IntValue{Value: 1},
		Value2:  &IntValue{Value: -1},
		Compare: 1,
	},
	{
		Name:    "MaxInt64/MinInt64",
		Value1:  &IntValue{Value: math.MaxInt64},
		Value2:  &IntValue{Value: math.MinInt64},
		Compare: 1,
	},
	{
		Name:    "MinInt64/MaxInt64",
		Value1:  &IntValue{Value: math.MinInt64},
		Value2:  &IntValue{Value: math.MaxInt64},
		Compare: -1,
	},
	{
		Name:    "0/0",
		Value1:  &IntValue{Value: 0},
		Value2:  &IntValue{Value: 0},
		Compare: 0,
	},
	{
		Name:    "1/1",
		Value1:  &IntValue{Value: 1},
		Value2:  &IntValue{Value: 1},
		Compare: 0,
	},
	{
		Name:   "-1/-1",
		Value1: &IntValue{Value: -1},
		Value2: &IntValue{Value: -1},
	},
}

func TestIntValue_Less(t *testing.T) {
	for _, tc := range lessThanIntTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{Comparable: true})
			if err != nil {
				t.Fatal(err)
			}
			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{Comparable: true})
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(v1, v2) != tc.Compare {
				t.Fatalf("expected %d, got %d", tc.Compare, bytes.Compare(v1, v2))
			}
		})
	}
}

func TestIntValue_LessDescending(t *testing.T) {
	for _, tc := range lessThanInt64TestCases {
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
