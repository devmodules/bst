package bstvalue

import (
	"bytes"
	"strings"
	"testing"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
	"github.com/devmodules/bst/internal/diff"
)

const sampleString32Bytes = "This is a sample string of 32bts"

var stringValueTestCases = []struct {
	Name   string
	Value  string
	Binary []byte
}{
	{
		Name:   "Empty",
		Value:  "",
		Binary: []byte{bstio.BinarySizeZero},
	},
	{
		Name:   "Short",
		Value:  "short",
		Binary: []byte{bstio.BinarySizeUint8, 0x05, 's', 'h', 'o', 'r', 't'},
	},
	{
		Name: "MaxUint8+1Bytes",
		Value: func() string {
			var sb strings.Builder
			// 256 / 32 = 8
			for i := 0; i < 8; i++ {
				sb.WriteString(sampleString32Bytes)
			}
			return sb.String()
		}(),
		Binary: func() []byte {
			buf := bytes.Buffer{}
			buf.Write([]byte{bstio.BinarySizeUint16, 0x01, 0x00})
			for i := 0; i < 8; i++ {
				buf.Write([]byte(sampleString32Bytes))
			}
			return buf.Bytes()
		}(),
	},
	{
		Name: "MaxUint16+1Bytes",
		Value: func() string {
			var sb strings.Builder
			// 65536 / 32 = 2048
			for i := 0; i < 2048; i++ {
				sb.WriteString(sampleString32Bytes)
			}
			return sb.String()
		}(),
		Binary: func() []byte {
			buf := bytes.Buffer{}
			buf.Write([]byte{0x03, 0x01, 0x00, 0x00})
			for i := 0; i < 2048; i++ {
				buf.Write([]byte(sampleString32Bytes))
			}
			return buf.Bytes()
		}(),
	},
}

func TestString(t *testing.T) {
	vt := bsttype.String()
	t.Run("Value", func(t *testing.T) {
		v := emptyStringValue(vt)
		if v.Kind() != bsttype.KindString {
			t.Fatalf("expected kind %d, got %d", bsttype.KindString, v.Kind())
		}
		if v.Kind() != vt.Kind() {
			t.Fatalf("expected type %v, got %v", vt, v.Type())
		}
	})
	t.Run("Type", func(t *testing.T) {
		if vt.Kind() != bsttype.KindString {
			t.Fatalf("expected kind %d, got %d", bsttype.KindString, vt.Kind())
		}
	})
}

func TestStringValue_ReadValue(t *testing.T) {
	for _, tc := range stringValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := EmptyStringValue()
			n, err := v.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}

			if v.Value != tc.Value {
				t.Fatalf("reading string value failed. Diff: %s", diff.Diff(tc.Value, v.Value))
			}
		})
	}
}

func TestStringValue_ReadValueDescending(t *testing.T) {
	for _, tc := range stringValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := EmptyStringValue()
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
				t.Fatalf("expected %s, got %s", tc.Value, v.Value)
			}
		})
	}
}

func TestStringValue_WriteValue(t *testing.T) {
	for _, tc := range stringValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &StringValue{Value: tc.Value}
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

func TestStringValue_WriteValueDescending(t *testing.T) {
	for _, tc := range stringValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &StringValue{Value: tc.Value}
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

func TestStringValue_MarshalDB(t *testing.T) {
	for _, tc := range stringValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := &StringValue{Value: tc.Value}
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

func TestStringValue_UnmarshalValue(t *testing.T) {
	for _, tc := range stringValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := EmptyStringValue()
			if err := v.UnmarshalValue(tc.Binary, bstio.ValueOptions{}); err != nil {
				t.Fatal(err)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %s, got %s", tc.Value, v.Value)
			}
		})
	}
}

func TestStringValue_Skip(t *testing.T) {
	for _, tc := range stringValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := EmptyStringValue()
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

var lessThanStringTestCases = []struct {
	Name   string
	Value1 *StringValue
	Value2 *StringValue
}{
	{
		Name:   "a/b",
		Value1: &StringValue{Value: "a"},
		Value2: &StringValue{Value: "b"},
	},
	{
		Name:   "b/a",
		Value1: &StringValue{Value: "b"},
		Value2: &StringValue{Value: "a"},
	},
	{
		Name:   "a/a",
		Value1: &StringValue{Value: "a"},
		Value2: &StringValue{Value: "a"},
	},
	{
		Name:   "a/abc",
		Value1: &StringValue{Value: "a"},
		Value2: &StringValue{Value: "abc"},
	},
	{
		Name:   "abc/a",
		Value1: &StringValue{Value: "abc"},
		Value2: &StringValue{Value: "a"},
	},
	{
		Name:   "A/a",
		Value1: &StringValue{Value: "A"},
		Value2: &StringValue{Value: "a"},
	},
	{
		Name:   "abcdef/z",
		Value1: &StringValue{Value: "abcdef"},
		Value2: &StringValue{Value: "z"},
	},
	{
		Name:   "z/abcdef",
		Value1: &StringValue{Value: "z"},
		Value2: &StringValue{Value: "abcdef"},
	},
}

func TestStringValue_Less(t *testing.T) {
	for _, tc := range lessThanStringTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{Comparable: true})
			if err != nil {
				t.Fatal(err)
			}
			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{Comparable: true})
			if err != nil {
				t.Fatal(err)
			}

			comp := bytes.Compare([]byte(tc.Value1.Value), []byte(tc.Value2.Value))

			if bytes.Compare(v1, v2) != comp {
				t.Fatalf("expected %d, got %d", comp, bytes.Compare(v1, v2))
			}
		})
	}
}

func TestStringValue_LessDescending(t *testing.T) {
	for _, tc := range lessThanStringTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{Descending: true, Comparable: true})
			if err != nil {
				t.Fatal(err)
			}
			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{Descending: true, Comparable: true})
			if err != nil {
				t.Fatal(err)
			}

			comp := bytes.Compare([]byte(tc.Value1.Value), []byte(tc.Value2.Value))
			if bytes.Compare(v1, v2) != -comp {
				t.Fatalf("expected %d, got %d", -comp, bytes.Compare(v1, v2))
			}
		})
	}
}
