package bstvalue

import (
	"bytes"
	"math"
	"testing"

	"github.com/devmodules/bst/bstio"
)

var float32ValueTestCases = []struct {
	Name   string
	Value  float32
	Binary []byte
}{
	{
		Name:   "Zero",
		Value:  0,
		Binary: []byte{0x80, 0x00, 0x00, 0x00},
	},
	{
		Name:   "Positive",
		Value:  1.0,
		Binary: bstio.MarshalFloat32(1.0, false),
	},
	{
		Name:   "Negative",
		Value:  -1.0,
		Binary: bstio.MarshalFloat32(-1.0, false),
	},
	{
		Name:   "Max",
		Value:  math.MaxFloat32,
		Binary: bstio.MarshalFloat32(math.MaxFloat32, false),
	},
	{
		Name:   "Min",
		Value:  math.SmallestNonzeroFloat32,
		Binary: bstio.MarshalFloat32(math.SmallestNonzeroFloat32, false),
	},
}

func TestFloat32Value_ReadValue(t *testing.T) {
	for _, testCase := range float32ValueTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var value Float32Value
			n, err := value.ReadValue(bytes.NewReader(testCase.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(testCase.Binary) {
				t.Fatalf("expected to read %d bytes, got %d", len(testCase.Binary), n)
			}

			if value.Value != testCase.Value {
				t.Errorf("value = %v, want %v", value.Value, testCase.Value)
			}
		})
	}
}

func TestFloat32Value_WriteValue(t *testing.T) {
	for _, testCase := range float32ValueTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			value := Float32Value{Value: testCase.Value}
			buf := bytes.Buffer{}
			n, err := value.WriteValue(&buf, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(testCase.Binary) {
				t.Fatalf("expected to write %d bytes, got %d", len(testCase.Binary), n)
			}

			if !bytes.Equal(buf.Bytes(), testCase.Binary) {
				t.Fatalf("expected to write %v, got %v", testCase.Binary, buf.Bytes())
			}
		})
	}
}

func TestFloat32Value_MarshalDB(t *testing.T) {
	for _, testCase := range float32ValueTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			value := Float32Value{Value: testCase.Value}
			data, err := value.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(data, testCase.Binary) {
				t.Fatalf("expected to marshal %v, got %v", testCase.Binary, data)
			}
		})
	}
}

func TestFloat32Value_UnmarshalValue(t *testing.T) {
	for _, testCase := range float32ValueTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var value Float32Value
			err := value.UnmarshalValue(testCase.Binary, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if value.Value != testCase.Value {
				t.Errorf("value = %v, want %v", value.Value, testCase.Value)
			}
		})
	}
}

func TestFloat32Value_Skip(t *testing.T) {
	for _, testCase := range float32ValueTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var value Float32Value
			n, err := value.Skip(bytes.NewReader(testCase.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if int(n) != len(testCase.Binary) {
				t.Fatalf("expected to skip %d bytes, got %d", len(testCase.Binary), n)
			}
		})
	}
}

var lessThanFloat32ValueTestCases = []struct {
	Name    string
	Value1  Float32Value
	Value2  Float32Value
	Compare int
}{
	{
		Name:    "Zero/1.3",
		Value1:  Float32Value{Value: 0},
		Value2:  Float32Value{Value: 1.3},
		Compare: -1,
	},
	{
		Name:    "1.3/Zero",
		Value1:  Float32Value{Value: 1.3},
		Value2:  Float32Value{Value: 0},
		Compare: 1,
	},
	{
		Name:    "1.3/1.3",
		Value1:  Float32Value{Value: 1.3},
		Value2:  Float32Value{Value: 1.3},
		Compare: 0,
	},
	{
		Name:    "Max/Min",
		Value1:  Float32Value{Value: math.MaxFloat32},
		Value2:  Float32Value{Value: math.SmallestNonzeroFloat32},
		Compare: 1,
	},
	{
		Name:    "Min/Max",
		Value1:  Float32Value{Value: math.SmallestNonzeroFloat32},
		Value2:  Float32Value{Value: math.MaxFloat32},
		Compare: -1,
	},
}

func TestFloat32Value_Less(t *testing.T) {
	for _, tc := range lessThanFloat32ValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1, err := tc.Value1.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			v2, err := tc.Value2.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if tc.Compare != bytes.Compare(v1, v2) {
				t.Errorf("expected %d, got %d", tc.Compare, bytes.Compare(v1, v2))
			}
		})
	}
}

func TestFloat32Value_LessDescending(t *testing.T) {
	for _, tc := range lessThanFloat32ValueTestCases {
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
				t.Errorf("expected %d, got %d", -tc.Compare, bytes.Compare(v1, v2))
			}
		})
	}
}

var float64ValueTestCases = []struct {
	Name   string
	Value  float64
	Binary []byte
}{
	{
		Name:   "Zero",
		Value:  0,
		Binary: []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	},
	{
		Name:   "1.3",
		Value:  1.3,
		Binary: bstio.MarshalFloat64(1.3, false),
	},
	{
		Name:   "Max",
		Value:  math.MaxFloat64,
		Binary: bstio.MarshalFloat64(math.MaxFloat64, false),
	},
	{
		Name:   "Min",
		Value:  math.SmallestNonzeroFloat64,
		Binary: bstio.MarshalFloat64(math.SmallestNonzeroFloat64, false),
	},
	{
		Name:   "NaN",
		Value:  math.NaN(),
		Binary: bstio.MarshalFloat64(math.NaN(), false),
	},
	{
		Name:   "Inf",
		Value:  math.Inf(1),
		Binary: bstio.MarshalFloat64(math.Inf(1), false),
	},
	{
		Name:   "-Inf",
		Value:  math.Inf(-1),
		Binary: bstio.MarshalFloat64(math.Inf(-1), false),
	},
}

func TestFloat64Value_ReadValue(t *testing.T) {
	for _, testCase := range float64ValueTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var value Float64Value
			n, err := value.ReadValue(bytes.NewReader(testCase.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(testCase.Binary) {
				t.Fatalf("expected to read %d bytes, got %d", len(testCase.Binary), n)
			}

			if math.IsNaN(testCase.Value) {
				if !math.IsNaN(value.Value) {
					t.Errorf("expected NaN, got %v", value.Value)
				}
				return
			}

			if value.Value != testCase.Value {
				t.Errorf("value = %v, want %v", value.Value, testCase.Value)
			}
		})
	}
}

func TestFloat64Value_WriteValue(t *testing.T) {
	for _, testCase := range float64ValueTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var buf bytes.Buffer
			v := Float64Value{Value: testCase.Value}
			n, err := v.WriteValue(&buf, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(testCase.Binary) {
				t.Fatalf("expected to write %d bytes, got %d", len(testCase.Binary), n)
			}

			if !bytes.Equal(buf.Bytes(), testCase.Binary) {
				t.Errorf("expected %v, got %v", testCase.Binary, buf.Bytes())
			}
		})
	}
}
