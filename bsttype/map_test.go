package bsttype

import (
	"bytes"
	"reflect"
	"testing"
)

var mapTypeTestCases = []struct {
	Name    string
	MapType Map
	Binary  []byte
}{
	{
		Name: "Key(String)/Value(String)",
		MapType: Map{
			Key:   MapElement{Type: String()},
			Value: MapElement{Type: String()},
		},
		Binary: []byte{
			// Key Type Header
			byte(KindString),
			// Value Type Header
			byte(KindString),
		},
	},
	{
		Name: "Key(String)/Value(Int32)",
		MapType: Map{
			Key:   MapElement{Type: String()},
			Value: MapElement{Type: Int32()},
		},
		Binary: []byte{
			// Key Type Header
			byte(KindString),
			// Value Type Header
			byte(KindInt32),
		},
	},
	{
		Name: "Key(Int32)/Value(String)/Descending",
		MapType: Map{
			Key:   MapElement{Type: Int32(), Descending: true},
			Value: MapElement{Type: String()},
		},
		Binary: []byte{
			// Key Type Header
			byte(KindInt32) | 0x80,
			// Value Type Header
			byte(KindString),
		},
	},
	{
		Name: "Key(Int32)/Bytes",
		MapType: Map{
			Key:   MapElement{Type: Int32()},
			Value: MapElement{Type: &Bytes{}},
		},
		Binary: []byte{
			// Key Type Header
			byte(KindInt32),
			// Value Type Header
			byte(KindBytes),
			// Value Type Content
			// No Fixed Length Flag
			0x00,
		},
	},
	{
		Name: "Key(Bytes)/Value(Int32)",
		MapType: Map{
			Key:   MapElement{Type: &Bytes{}},
			Value: MapElement{Type: Int32()},
		},
		Binary: []byte{
			// Key Type Header
			byte(KindBytes),
			// Key Type Content
			// No Fixed Length Flag
			0x00,
			// Value Type Header
			byte(KindInt32),
		},
	},
}

func TestMapType_ReadType(t *testing.T) {
	for _, tc := range mapTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var mt Map
			n, err := mt.ReadType(bytes.NewReader(tc.Binary))
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("Expected to read %d bytes, but read %d", len(tc.Binary), n)
			}

			if !reflect.DeepEqual(mt, tc.MapType) {
				t.Fatalf("Expected %v, but got %v", tc.MapType, mt)
			}
		})
	}
}

func TestMapType_WriteType(t *testing.T) {
	for _, tc := range mapTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			var buf bytes.Buffer
			n, err := tc.MapType.WriteType(&buf)
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("Expected to write %d bytes, but wrote %d", len(tc.Binary), n)
			}

			if !bytes.Equal(buf.Bytes(), tc.Binary) {
				t.Fatalf("Expected %v, but got %v", tc.Binary, buf.Bytes())
			}
		})
	}
}

func TestMapType_SkipType(t *testing.T) {
	for _, tc := range mapTypeTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			n, err := tc.MapType.SkipType(bytes.NewReader(tc.Binary))
			if err != nil {
				t.Fatal(err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("Expected to skip %d bytes, but skipped %d", len(tc.Binary), n)
			}
		})
	}
}

func TestMapType_CompareType(t *testing.T) {
	equalTestCases := []struct {
		Name string
		A, B *Map
		Eq   bool
	}{
		{
			Name: "Equal/Key(String)/Value(String)",
			A: &Map{
				Key:   MapElement{Type: String()},
				Value: MapElement{Type: String()},
			},
			B: &Map{
				Key:   MapElement{Type: String()},
				Value: MapElement{Type: String()},
			},
			Eq: true,
		},
		{
			Name: "Equal/Key(String)/Value(Int32)",
			A: &Map{
				Key:   MapElement{Type: String()},
				Value: MapElement{Type: Int32()},
			},
			B: &Map{
				Key:   MapElement{Type: String()},
				Value: MapElement{Type: Int32()},
			},
			Eq: true,
		},
		{
			Name: "NotEqual/Key(String)/Value(String)/OneDescending",
			A: &Map{
				Key:   MapElement{Type: String(), Descending: true},
				Value: MapElement{Type: String()},
			},
			B: &Map{
				Key:   MapElement{Type: String()},
				Value: MapElement{Type: String()},
			},
			Eq: false,
		},
		{
			Name: "Equal/Key(String)/Value(String)/BothDescending",
			A: &Map{
				Key:   MapElement{Type: String(), Descending: true},
				Value: MapElement{Type: String()},
			},
			B: &Map{
				Key:   MapElement{Type: String(), Descending: true},
				Value: MapElement{Type: String()},
			},
			Eq: true,
		},
		{
			Name: "NotEqual/K(String)V(Int)/K(String)V(String)",
			A: &Map{
				Key:   MapElement{Type: String()},
				Value: MapElement{Type: Int32()},
			},
			B: &Map{
				Key:   MapElement{Type: String()},
				Value: MapElement{Type: String()},
			},
			Eq: false,
		},
	}

	for _, tc := range equalTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.A.CompareType(tc.B) != tc.Eq {
				t.Fatalf("Expected %v, but got %v", tc.Eq, !tc.Eq)
			}
		})
	}
}
