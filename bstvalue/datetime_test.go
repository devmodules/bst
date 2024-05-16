package bstvalue

import (
	"bytes"
	"testing"
	"time"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

func mustMarshalDateTime(t time.Time, options bstio.ValueOptions) []byte {
	v := &DateTime{Value: t, DateTimeType: &bsttype.DateTime{}}
	b, err := v.MarshalValue(options)
	if err != nil {
		panic(err)
	}
	return b
}

var testTimes = []time.Time{
	time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	time.Date(2022, 5, 19, 16, 56, 0, 1, time.FixedZone("Europe/Warsaw", 2*60*60)),
	time.Date(1990, 2, 15, 0, 0, 0, 2, time.FixedZone("Custom", -10*60*60)),
}

var dateTimeTestCases = []struct {
	Name   string
	Type   bsttype.DateTime
	Input  time.Time
	Value  time.Time
	Binary []byte
}{
	{
		Name:   "NoFixedZone/UTCInput",
		Type:   bsttype.DateTime{},
		Input:  testTimes[0],
		Value:  testTimes[0],
		Binary: mustMarshalDateTime(testTimes[0], bstio.ValueOptions{}),
	},
	{
		Name: "FixedZone/UTCInput",
		Type: bsttype.DateTime{
			HasFixedZone: true,
			FixedZone: bsttype.DateTimeFixedZone{
				Name:   "UTC-8",
				Offset: -8 * 60 * 60,
			},
		},
		Input:  testTimes[0],
		Value:  testTimes[0].In(time.FixedZone("UTC-8", -8*60*60)),
		Binary: mustMarshalDateTime(testTimes[0].In(time.FixedZone("UTC-8", -8*60*60)), bstio.ValueOptions{}),
	},
	{
		Name:   "FixedZone/Europe/Warsaw",
		Type:   bsttype.DateTime{HasFixedZone: true, FixedZone: bsttype.DateTimeFixedZone{Name: "Europe/Warsaw", Offset: 2 * 60 * 60}},
		Input:  testTimes[0],
		Value:  testTimes[0].In(time.FixedZone("Europe/Warsaw", 2*60*60)),
		Binary: mustMarshalDateTime(testTimes[0].In(time.FixedZone("Europe/Warsaw", 2*60*60)), bstio.ValueOptions{}),
	},
	{
		Name:   "NoFixedZone/Europe/Warsaw",
		Type:   bsttype.DateTime{},
		Input:  testTimes[1],
		Value:  testTimes[1],
		Binary: mustMarshalDateTime(testTimes[1], bstio.ValueOptions{}),
	},
	{
		Name:   "NoFixedZone/Custom",
		Type:   bsttype.DateTime{},
		Input:  testTimes[2],
		Value:  testTimes[2],
		Binary: mustMarshalDateTime(testTimes[2], bstio.ValueOptions{}),
	},
}

func TestDateTimeValue_ReadValue(t *testing.T) {
	for _, tc := range dateTimeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			tp := tc.Type
			v := EmptyDateTimeValue(&tp)
			n, err := v.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if n != len(tc.Binary) {
				t.Errorf("got %d, want %d", n, len(tc.Binary))
			}

			if !v.Value.Equal(tc.Value) {
				t.Fatalf("unexpected value: %v", v.Value)
			}
		})
	}
}

func TestDateTimeValue_WriteValue(t *testing.T) {
	for _, tc := range dateTimeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := NewDateTimeValue(&tc.Type, tc.Input)
			var buf bytes.Buffer
			n, err := v.WriteValue(&buf, bstio.ValueOptions{})
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

func TestDateTimeValue_MarshalDB(t *testing.T) {
	for _, tc := range dateTimeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := NewDateTimeValue(&tc.Type, tc.Input)
			b, err := v.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(b, tc.Binary) {
				t.Errorf("got %v, want %v", b, tc.Binary)
			}
		})
	}
}

func TestDateTimeValue_UnmarshalValue(t *testing.T) {
	for _, tc := range dateTimeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := EmptyDateTimeValue(&tc.Type)
			err := v.UnmarshalValue(tc.Binary, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if !v.Value.Equal(tc.Value) {
				t.Fatalf("unexpected value: %v", v.Value)
			}
		})
	}
}

func TestDateTimeValue_Skip(t *testing.T) {
	for _, tc := range dateTimeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			v := EmptyDateTimeValue(&tc.Type)
			n, err := v.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if int(n) != len(tc.Binary) {
				t.Errorf("got %d, want %d", n, len(tc.Binary))
			}
		})
	}
}
