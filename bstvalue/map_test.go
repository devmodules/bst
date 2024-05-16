package bstvalue

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

var mapValueTestCases = []struct {
	Name    string
	MapType *bsttype.Map
	Values  []MapValueKV
	Binary  []byte
}{
	{
		Name: "Empty",
		MapType: &bsttype.Map{
			Key:   bsttype.MapElement{Type: bsttype.String()},
			Value: bsttype.MapElement{Type: bsttype.Any()},
		},
		Binary: []byte{
			// 0-byte binary length.
			0x00,
		},
	},
	{
		Name: "String-Int32",
		MapType: &bsttype.Map{
			Key:   bsttype.MapElement{Type: bsttype.String()},
			Value: bsttype.MapElement{Type: bsttype.Int32()},
		},
		Values: []MapValueKV{
			{
				Key:   NewStringValue("foo"),
				Value: NewInt32Value(42),
			},
			{
				Key:   NewStringValue("bar"),
				Value: NewInt32Value(43),
			},
			{
				Key:   NewStringValue("goo"),
				Value: NewInt32Value(44),
			},
		},
		Binary: []byte{
			// Binary size of Uint for 3 kv pairs.
			bstio.BinarySizeUint8, 0x03,
			// The keys are sorted by the string value.
			// Thus, the first key is "bar".
			//
			// String length of "bar".
			bstio.BinarySizeUint8, 0x03,
			// String value of "bar".
			'b', 'a', 'r',
			// Int value of 43.
			0x80, 0x00, 0x00, 0x2b,
			// The second key is "foo".
			//
			// String length of "foo".
			bstio.BinarySizeUint8, 0x03,
			// String value of "foo".
			'f', 'o', 'o',
			// Int value of 42.
			0x80, 0x00, 0x00, 0x2a,
			// The third key is "goo".
			//
			// String length of "goo".
			bstio.BinarySizeUint8, 0x03,
			// String value of "goo".
			'g', 'o', 'o',
			// Int value of 44.
			0x80, 0x00, 0x00, 0x2c,
		},
	},
	{
		Name: "Int32-String/Bigger",
		MapType: &bsttype.Map{
			Key:   bsttype.MapElement{Type: bsttype.Int32()},
			Value: bsttype.MapElement{Type: bsttype.String()},
		},
		Values: func() []MapValueKV {
			values := make([]MapValueKV, 10000)
			for i := 0; i < 10000; i++ {
				values[i] = MapValueKV{
					Key:   NewInt32Value(int32(i)),
					Value: NewStringValue(fmt.Sprintf("%d", i)),
				}
			}
			return values
		}(),
		Binary: func() []byte {
			var buf bytes.Buffer
			buf.Write([]byte{
				// Binary size of Uint for 10000 kv pairs.
				bstio.BinarySizeUint16, 0x27, 0x10,
			})
			for i := 0; i < 10000; i++ {
				// Int Key value of i.
				buf.Write(bstio.MarshalInt32(int32(i), false))
				// String length of i.
				si := fmt.Sprintf("%d", i)
				buf.Write(bstio.EncodeStringNonComparable(si, false))
			}
			return buf.Bytes()
		}(),
	},
}

func TestMapValue_ReadValue(t *testing.T) {
	for _, tc := range mapValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			mv := EmptyMapValue(tc.MapType)
			n, err := mv.ReadValue(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if n != len(tc.Binary) {
				t.Fatalf("Expected to read %d bytes, but read %d", len(tc.Binary), n)
			}

			if mv.Len() != len(tc.Values) {
				t.Fatalf("Expected to read %d values, but read %d", len(tc.Values), mv.Len())
			}

			//  Get the KV pairs from the map.
			// The result is sorted by the key.
			kvs := mv.KeyValues()

			// Prepare sorted expected KV pairs.
			sortedKV := make([]MapValueKV, len(tc.Values))
			copy(sortedKV, tc.Values)

			sort.Slice(sortedKV, func(i, j int) bool {
				switch mv.MapType.Key.Type.Kind() {
				case bsttype.KindString:
					return sortedKV[i].Key.(*StringValue).Value < sortedKV[j].Key.(*StringValue).Value
				case bsttype.KindInt32:
					return sortedKV[i].Key.(*Int32Value).Value < sortedKV[j].Key.(*Int32Value).Value
				default:
					t.Fatalf("unexpected key type %v", mv.MapType.Key)
					return false
				}
			})

			// Compare the sorted KV pairs.
			if !reflect.DeepEqual(kvs, sortedKV) {
				t.Fatalf("Expected %v, but got %v", tc.Values, kvs)
			}
		})
	}
}

func TestMapValue_WriteValue(t *testing.T) {
	for _, tc := range mapValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			mv, err := NewMapValue(tc.MapType, tc.Values...)
			if err != nil {
				t.Fatal(err)
			}

			var buf bytes.Buffer
			n, err := mv.WriteValue(&buf, bstio.ValueOptions{})
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

func TestMapValue_Skip(t *testing.T) {
	for _, tc := range mapValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			mv := EmptyMapValue(tc.MapType)
			n, err := mv.Skip(bytes.NewReader(tc.Binary), bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if int(n) != len(tc.Binary) {
				t.Fatalf("Expected to skip %d bytes, but skipped %d", len(tc.Binary), n)
			}
		})
	}
}

func TestMapValue_MarshalDB(t *testing.T) {
	for _, tc := range mapValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			mv, err := NewMapValue(tc.MapType, tc.Values...)
			if err != nil {
				t.Fatal(err)
			}

			data, err := mv.MarshalValue(bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(data, tc.Binary) {
				t.Fatalf("Expected %v, but got %v", tc.Binary, data)
			}
		})
	}
}

func TestMapValue_UnmarshalValue(t *testing.T) {
	for _, tc := range mapValueTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			mv := EmptyMapValue(tc.MapType)
			err := mv.UnmarshalValue(tc.Binary, bstio.ValueOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if mv.Len() != len(tc.Values) {
				t.Fatalf("Expected to read %d values, but read %d", len(tc.Values), mv.Len())
			}

			//  Get the KV pairs from the map.
			// The result is sorted by the key.
			kvs := mv.KeyValues()

			// Prepare sorted expected KV pairs.
			sortedKV := make([]MapValueKV, len(tc.Values))
			copy(sortedKV, tc.Values)

			sort.Slice(sortedKV, func(i, j int) bool {
				switch mv.MapType.Key.Type.Kind() {
				case bsttype.KindString:
					return sortedKV[i].Key.(*StringValue).Value < sortedKV[j].Key.(*StringValue).Value
				case bsttype.KindInt32:
					return sortedKV[i].Key.(*Int32Value).Value < sortedKV[j].Key.(*Int32Value).Value
				default:
					t.Fatalf("unexpected key type %v", mv.MapType.Key)
					return false
				}
			})

			// Compare the sorted KV pairs.
			if !reflect.DeepEqual(kvs, sortedKV) {
				t.Fatalf("Expected %v, but got %v", tc.Values, kvs)
			}
		})
	}
}
