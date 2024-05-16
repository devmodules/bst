package bsttype

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/devmodules/bst/bstio"
)

var dateTimeTypeTestCases = []struct {
	Name   string
	Type   DateTime
	Binary []byte
}{
	{
		Name: "NoFixedZone",
		Type: DateTime{},
		Binary: []byte{
			// FixedZone Null value
			0x00,
		},
	},
	{
		Name: "FixedZone/UTC",
		Type: DateTime{
			HasFixedZone: true,
			FixedZone: DateTimeFixedZone{
				Name:   "UTC",
				Offset: 0,
			},
		},
		Binary: []byte{
			// FixedZone NotNull Header
			0x01,
			// FixedZone Name
			// Length
			bstio.BinarySizeUint8, byte(len("UTC")),
			// Value
			'U', 'T', 'C',
			// FixedZone Offset Int32 Value
			0x80, 0x00, 0x00, 0x00,
		},
	},
}

func TestDateTimeType_ReadType(t *testing.T) {
	for _, tc := range dateTimeTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := DateTime{}
			n, err := v.ReadType(bytes.NewReader(tc.Binary))
			if err != nil {
				t.Fatal(err)
			}
			if n != len(tc.Binary) {
				t.Errorf("got %d, want %d", n, len(tc.Binary))
			}

			if !reflect.DeepEqual(v, tc.Type) {
				t.Fatalf("unexpected type: %v, want %v", v, tc.Type)
			}
		})
	}
}

func TestDateTimeType_WriteType(t *testing.T) {
	for _, tc := range dateTimeTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := tc.Type
			var buf bytes.Buffer
			n, err := v.WriteType(&buf)
			if err != nil {
				t.Fatal(err)
			}
			if n != len(tc.Binary) {
				t.Errorf("got %d, want %d", n, len(tc.Binary))
			}
			if !bytes.Equal(buf.Bytes(), tc.Binary) {
				t.Errorf("got %v, want %v", buf.Bytes(), tc.Binary)
			}
		})
	}
}

func TestDateTimeType_SkipType(t *testing.T) {
	for _, tc := range dateTimeTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var v DateTime
			n, err := v.SkipType(bytes.NewReader(tc.Binary))
			if err != nil {
				t.Fatal(err)
			}
			if int(n) != len(tc.Binary) {
				t.Errorf("got %d, want %d", n, len(tc.Binary))
			}
		})
	}
}

func TestDateTimeType(t *testing.T) {
	t.Run("Kind", func(t *testing.T) {
		var dt DateTime
		if dt.Kind() != KindDateTime {
			t.Fatalf("got %v, want %v", dt.Kind(), KindDateTime)
		}
	})

	t.Run("Location", func(t *testing.T) {
		var dt DateTime
		if dt.Location() != nil {
			t.Fatalf("got %v, want nil", dt.Location())
		}

		dt.HasFixedZone = true
		dt.FixedZone = DateTimeFixedZone{
			Name:   "UTC",
			Offset: 0,
		}
		if dt.Location() == nil {
			t.Fatalf("got nil, want %v", time.FixedZone("UTC", 0))
		}
	})

	eqTestCases := []struct {
		Name string
		A, B *DateTime
		Eq   bool
	}{
		{
			Name: "Equal/NoFixedZone",
			A:    &DateTime{},
			B:    &DateTime{},
			Eq:   true,
		},
		{
			Name: "Equal/FixedZone/UTC",
			A:    &DateTime{HasFixedZone: true, FixedZone: DateTimeFixedZone{Name: "UTC", Offset: 0}},
			B:    &DateTime{HasFixedZone: true, FixedZone: DateTimeFixedZone{Name: "UTC", Offset: 0}},
			Eq:   true,
		},
		{
			Name: "NotEqual/FixedZone(UTC)/FixedZone(UTC-1)",
			A:    &DateTime{HasFixedZone: true, FixedZone: DateTimeFixedZone{Name: "UTC", Offset: 0}},
			B:    &DateTime{HasFixedZone: true, FixedZone: DateTimeFixedZone{Name: "UTC", Offset: -1 * 60 * 60}},
			Eq:   false,
		},
		{
			Name: "NotEqual/FixedZone(UTC-1)/NoFixedZone",
			A:    &DateTime{HasFixedZone: true, FixedZone: DateTimeFixedZone{Name: "UTC", Offset: -1 * 60 * 60}},
			B:    &DateTime{},
			Eq:   false,
		},
	}

	for _, tc := range eqTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.A.CompareType(tc.B) != tc.Eq {
				t.Fatalf("got %v, want %v", tc.A.CompareType(tc.B), tc.Eq)
			}
		})
	}
}
