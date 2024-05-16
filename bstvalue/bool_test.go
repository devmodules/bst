package bstvalue

import (
	"bytes"
	"testing"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

var boolTestCases = []struct {
	Name   string
	Value  bool
	Binary []byte
}{
	{
		Name:   "true",
		Value:  true,
		Binary: []byte{0x01},
	},
	{
		Name:   "false",
		Value:  false,
		Binary: []byte{0x00},
	},
}

func TestBool(t *testing.T) {
	bt := bsttype.Boolean()
	t.Run("Value", func(t *testing.T) {
		v := emptyBoolValue(bt)
		if v.Kind() != bsttype.KindBoolean {
			t.Fatalf("expected kind %d, got %d", bsttype.KindBoolean, v.Kind())
		}
		if v.Kind() != bt.Kind() {
			t.Fatalf("expected type %v, got %v", bt, v.Type())
		}
	})

	t.Run("Type", func(t *testing.T) {
		if bt.Kind() != bsttype.KindBoolean {
			t.Fatalf("expected kind %d, got %d", bsttype.KindBoolean, bt.Kind())
		}
	})
}

func TestBoolValue_ReadValue(t *testing.T) {
	for _, tc := range boolTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v BoolValue
			n, err := v.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected value %t, got %t", tc.Value, v.Value)
			}
		})
	}
}

func TestBoolValue_ReadValueDescending(t *testing.T) {
	for _, tc := range boolTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v BoolValue
			n, err := v.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}

			if v.Value == tc.Value {
				t.Fatalf("expected value %t, got %t", !tc.Value, v.Value)
			}
		})
	}
}

func TestBoolValue_MarshalDB(t *testing.T) {
	for _, tc := range boolTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v BoolValue
			v.Value = tc.Value

			b, err := v.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !bytes.Equal(b, tc.Binary) {
				t.Fatalf("expected %v, got %v", tc.Binary, b)
			}
		})
	}
}

func TestBoolValue_UnmarshalValue(t *testing.T) {
	for _, tc := range boolTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v BoolValue
			err := v.UnmarshalValue(tc.Binary, bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if v.Value != tc.Value {
				t.Fatalf("expected value %t, got %t", tc.Value, v.Value)
			}
		})
	}
}

func TestBoolValue_SkipDB(t *testing.T) {
	for _, tc := range boolTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v BoolValue
			n, err := v.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("expected %d, got %d", len(tc.Binary), n)
			}
		})
	}
}

var lessThanBoolTestCases = []struct {
	Name    string
	Value1  *BoolValue
	Value2  *BoolValue
	Compare int
}{
	{
		Name:    "true>false",
		Value1:  &BoolValue{Value: true},
		Value2:  &BoolValue{Value: false},
		Compare: 1,
	},
	{
		Name:    "false<true",
		Value1:  &BoolValue{Value: false},
		Value2:  &BoolValue{Value: true},
		Compare: -1,
	},
	{
		Name:    "true=true",
		Value1:  &BoolValue{Value: true},
		Value2:  &BoolValue{Value: true},
		Compare: 0,
	},
	{
		Name:    "false=false",
		Value1:  &BoolValue{Value: false},
		Value2:  &BoolValue{Value: false},
		Compare: 0,
	},
}

func TestBoolValueLess(t *testing.T) {
	for _, tc := range lessThanBoolTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if tc.Compare != bytes.Compare(v1, v2) {
				t.Fatalf("expected %d, got %d", tc.Compare, bytes.Compare(v1, v2))
			}
		})
	}
}

func TestBoolValueLessDescending(t *testing.T) {
	for _, tc := range lessThanBoolTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if tc.Compare != -bytes.Compare(v1, v2) {
				t.Fatalf("expected %d, got %d", tc.Compare, -bytes.Compare(v1, v2))
			}
		})
	}
}
