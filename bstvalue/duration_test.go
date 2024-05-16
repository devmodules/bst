package bstvalue

import (
	"bytes"
	"testing"
	"time"

	"github.com/devmodules/bst/bstio"
)

var durationValueTestCases = []struct {
	Name   string
	Value  time.Duration
	Binary []byte
}{
	{
		Name:   "0",
		Value:  0,
		Binary: []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	},
	{
		Name:   "100s",
		Value:  100 * time.Second, // 100000000000
		Binary: bstio.MarshalInt64(100000000000, false),
	},
	{
		Name:   "1m",
		Value:  time.Minute, // 60000000000
		Binary: bstio.MarshalInt64(60000000000, false),
	},
	{
		Name:   "1h",
		Value:  time.Hour, // 3600000000000
		Binary: bstio.MarshalInt64(3600000000000, false),
	},
	{
		Name:   "-1h",
		Value:  -time.Hour, // -3600000000000
		Binary: bstio.MarshalInt64(-3600000000000, false),
	},
	{
		Name:   "1h1m",
		Value:  time.Hour + time.Minute, // 3600000000000 + 60000000000
		Binary: bstio.MarshalInt64(3600000000000+60000000000, false),
	},
}

func TestDurationValue_ReadValue(t *testing.T) {
	for _, tc := range durationValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v DurationValue
			n, err := v.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("unexpected read length: %d != %d", n, len(tc.Binary))
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %v, got %v", tc.Value, v.Value)
			}
		})
	}
}

func TestDurationValue_ReadValueDescending(t *testing.T) {
	for _, tc := range durationValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v DurationValue

			cp := make([]byte, len(tc.Binary))
			copy(cp, tc.Binary)
			bstio.ReverseBytes(cp)

			n, err := v.ReadValue(bytes.NewReader(cp), bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(cp) {
				t.Fatalf("unexpected read length: %d != %d", n, len(cp))
			}

			if v.Value != tc.Value {
				t.Fatalf("expected %v, got %v", tc.Value, v.Value)
			}
		})
	}
}

func TestDurationValue_WriteValue(t *testing.T) {
	for _, tc := range durationValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var buf bytes.Buffer

			v := NewDurationValue(tc.Value)
			n, err := v.WriteValue(&buf, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("unexpected write length: %d != %d", n, len(tc.Binary))
			}

			if !bytes.Equal(buf.Bytes(), tc.Binary) {
				t.Fatalf("expected %v, got %v", tc.Binary, buf.Bytes())
			}
		})
	}
}

func TestDurationValue_WriteValueDescending(t *testing.T) {
	for _, tc := range durationValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var buf bytes.Buffer

			v := NewDurationValue(tc.Value)
			n, err := v.WriteValue(&buf, bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("unexpected write length: %d != %d", n, len(tc.Binary))
			}

			cp := make([]byte, len(tc.Binary))
			copy(cp, tc.Binary)
			bstio.ReverseBytes(cp)

			if !bytes.Equal(buf.Bytes(), cp) {
				t.Fatalf("expected %v, got %v", tc.Binary, buf.Bytes())
			}
		})
	}
}

func TestDurationValue_Skip(t *testing.T) {
	for _, tc := range durationValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v DurationValue
			n, err := v.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("unexpected skip length: %d != %d", n, len(tc.Binary))
			}
		})
	}
}

var lessThanDurationValueTestCases = []struct {
	Name    string
	Value1  time.Duration
	Value2  time.Duration
	Compare int
}{
	{
		Name:    "0<1s",
		Value1:  0,
		Value2:  time.Second,
		Compare: -1,
	},
	{
		Name:    "1s=1s",
		Value1:  time.Second,
		Value2:  time.Second,
		Compare: 0,
	},
	{
		Name:    "3s>2s",
		Value1:  time.Second * 3,
		Value2:  time.Second * 2,
		Compare: 1,
	},
	{
		Name:    "-1h<1m",
		Value1:  -time.Hour,
		Value2:  time.Minute,
		Compare: -1,
	},
	{
		Name:    "1m>-1h",
		Value1:  time.Minute,
		Value2:  -time.Hour,
		Compare: 1,
	},
}

func TestDurationValue_Less(t *testing.T) {
	for _, tc := range lessThanDurationValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1 := NewDurationValue(tc.Value1)
			data1, err := v1.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			v2 := NewDurationValue(tc.Value2)
			data2, err := v2.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if bytes.Compare(data1, data2) != tc.Compare {
				t.Fatalf("expected %v, got %v", tc.Compare, bytes.Compare(data1, data2))
			}
		})
	}
}

func TestDurationValue_LessDescending(t *testing.T) {
	for _, tc := range lessThanDurationValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v1 := NewDurationValue(tc.Value1)
			data1, err := v1.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}
			v2 := NewDurationValue(tc.Value2)
			data2, err := v2.MarshalValue(bstio.ValueOptions{Descending: true})
			if err != nil {
				t.Fatal(err)
			}

			if bytes.Compare(data1, data2) != -tc.Compare {
				t.Fatalf("expected %v, got %v", tc.Compare, bytes.Compare(data1, data2))
			}
		})
	}
}
