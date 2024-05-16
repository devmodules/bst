package bstvalue

import (
	"bytes"
	"testing"
	"time"

	"github.com/devmodules/bst/bstio"
)

var timestampTestCases = []struct {
	Name   string
	Value  time.Time
	Binary []byte
}{
	{
		Name:   "2000-01-01 00:00:00 +0000 UTC",
		Value:  time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Binary: bstio.MarshalInt64(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano(), false),
	},
	{
		Name:   "2022-05-19 15:56:00 +0000 UTC",
		Value:  time.Date(2022, 5, 19, 15, 56, 0, 0, time.UTC),
		Binary: bstio.MarshalInt64(time.Date(2022, 5, 19, 15, 56, 0, 0, time.UTC).UnixNano(), false),
	},
}

func TestTimestampValue_ReadValue(t *testing.T) {
	for _, tc := range timestampTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var value TimestampValue
			n, err := value.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected to read %d bytes, got %d", len(tc.Binary), n)
			}

			if value.Value != tc.Value {
				t.Fatalf("unexpected value: %v", value.Value)
			}
		})
	}
}

func TestTimestampValue_WriteValue(t *testing.T) {
	for _, tc := range timestampTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var value TimestampValue
			value.Value = tc.Value

			var buf bytes.Buffer
			n, err := value.WriteValue(&buf, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("expected to write %d bytes, got %d", len(tc.Binary), n)
			}

			if !bytes.Equal(buf.Bytes(), tc.Binary) {
				t.Fatalf("unexpected value: %v", value.Value)
			}
		})
	}
}

func TestTimestampValue_Skip(t *testing.T) {
	for _, tc := range timestampTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var value TimestampValue
			n, err := value.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("expected to skip %d bytes, got %d", len(tc.Binary), n)
			}
		})
	}
}

func TestTimestampValue_UnmarshalValue(t *testing.T) {
	for _, tc := range timestampTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var value TimestampValue
			cp := make([]byte, len(tc.Binary))
			copy(cp, tc.Binary)

			err := value.UnmarshalValue(cp, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if value.Value != tc.Value {
				t.Fatalf("unexpected value: %v", value.Value)
			}
		})
	}
}

func TestTimestampValue_MarshalDB(t *testing.T) {
	for _, tc := range timestampTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var value TimestampValue
			value.Value = tc.Value

			data, err := value.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(data, tc.Binary) {
				t.Fatalf("unexpected value: %02x, expected: %02x", data, tc.Binary)
			}
		})
	}
}
