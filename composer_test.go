package bst

import (
	"bytes"
	"math"
	"testing"
	"time"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
	"github.com/devmodules/bst/internal/diff"
)

func TestComposerIntegers(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	t.Run("Int8", func(t *testing.T) {
		buf.Reset()
		c, err := NewComposer(buf, bsttype.Int8(), ComposerOptions{})
		if err != nil {
			t.Fatalf("creating composer failed: %v", err)
		}

		if err = c.WriteInt8(math.MaxInt8); err != nil {
			t.Fatalf("writing int8 failed: %v", err)
		}

		data := buf.Bytes()
		if len(data) != 2 {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		// The first byte of the data should contain a header - which in this option should be 0.
		if data[0] != 0 {
			t.Fatalf("unexpected header byte: %v", data[0])
		}
		if data[1] != math.MaxUint8 {
			t.Fatalf("unexpected int8 value binary value: %v, expected: %v", data[0], 0xff)
		}
		buf.Reset()
		c.Reset(ComposerOptions{})

		if err = c.WriteInt8(math.MinInt8); err != nil {
			t.Fatalf("writing int8 failed: %v", err)
		}

		data = buf.Bytes()
		if len(data) != 2 {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		// The first byte of the data should contain a header - which in this option should be 0.
		if data[0] != 0 {
			t.Fatalf("unexpected header byte: %v", data[0])
		}

		if data[1] != 0x00 {
			t.Fatalf("unexpected int8 value binary value: %v, expected: %v", data[0], 0x00)
		}

		buf.Reset()
		c.Reset(ComposerOptions{})
	})

	t.Run("Int16", func(t *testing.T) {
		buf.Reset()
		c, err := NewComposer(buf, bsttype.Int16(), ComposerOptions{})
		if err != nil {
			t.Fatalf("creating composer failed: %v", err)
		}

		if err = c.WriteInt16(math.MaxInt16); err != nil {
			t.Fatalf("writing int16 failed: %v", err)
		}

		data := buf.Bytes()
		if len(data) != 3 {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		if data[0] != 0x00 || data[1] != 0xff || data[2] != 0xff {
			t.Fatalf("unexpected int16 value binary value: %v, expected: %v", data, []byte{0x00, 0xff, 0xff})
		}

		buf.Reset()
		if err = c.Reset(ComposerOptions{}); err != nil {
			t.Fatalf("resetting composer failed: %v", err)
		}

		if err = c.WriteInt16(math.MinInt16); err != nil {
			t.Fatalf("writing int16 failed: %v", err)
		}

		data = buf.Bytes()
		if len(data) != 3 {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		if data[0] != 0x00 || data[1] != 0x00 || data[2] != 0x00 {
			t.Fatalf("unexpected int16 value binary value: %v, expected: %v", data, []byte{0x00, 0x00})
		}
	})

	t.Run("Int32", func(t *testing.T) {
		buf.Reset()
		c, err := NewComposer(buf, bsttype.Int32(), ComposerOptions{})
		if err != nil {
			t.Fatalf("creating composer failed: %v", err)
		}

		if err = c.WriteInt32(math.MaxInt32); err != nil {
			t.Fatalf("writing int32 failed: %v", err)
		}

		data := buf.Bytes()
		if len(data) != 5 {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		td := []byte{0x00, 0xff, 0xff, 0xff, 0xff}
		if !bytes.Equal(data, td) {
			t.Fatalf("unexpected int32 value binary value: %v, expected: %v", data, td)
		}

		buf.Reset()
		if err = c.Reset(ComposerOptions{}); err != nil {
			t.Fatalf("resetting composer failed: %v", err)
		}

		if err = c.WriteInt32(math.MinInt32); err != nil {
			t.Fatalf("writing int32 failed: %v", err)
		}

		data = buf.Bytes()
		if len(data) != 5 {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		if !bytes.Equal(data, []byte{0x00, 0x00, 0x00, 0x00, 0x00}) {
			t.Fatalf("unexpected int32 value binary value: %v, expected: %v", data, []byte{0x00, 0x00, 0x00, 0x00, 0x00})
		}

		buf.Reset()
		c.Reset(ComposerOptions{})
	})

	t.Run("Int64", func(t *testing.T) {
		buf.Reset()
		c, err := NewComposer(buf, bsttype.Int64(), ComposerOptions{})
		if err != nil {
			t.Fatalf("creating composer failed: %v", err)
		}

		if err = c.WriteInt64(math.MaxInt64); err != nil {
			t.Fatalf("writing int64 failed: %v", err)
		}

		data := buf.Bytes()
		if len(data) != 9 {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		if !bytes.Equal(data, []byte{0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
			t.Fatalf("unexpected int64 value binary value: %v, expected: %v", data, []byte{0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
		}

		buf.Reset()
		c.Reset(ComposerOptions{})

		if err = c.WriteInt64(math.MinInt64); err != nil {
			t.Fatalf("writing int64 failed: %v", err)
		}

		data = buf.Bytes()
		if len(data) != 9 {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		if !bytes.Equal(data, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) {
			t.Fatalf("unexpected int64 value binary value: %v, expected: %v", data, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		}

		buf.Reset()
		c.Reset(ComposerOptions{})
	})

	t.Run("Int", func(t *testing.T) {
		buf.Reset()
		c, err := NewComposer(buf, bsttype.Int(), ComposerOptions{})
		if err != nil {
			t.Fatalf("creating composer failed: %v", err)
		}

		if err = c.WriteInt(math.MaxInt64); err != nil {
			t.Fatalf("writing int failed: %v", err)
		}

		data := buf.Bytes()[1:]
		if len(data) != 9 {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		if !bytes.Equal(data, []byte{bstio.BinarySizeUint64, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
			t.Fatalf("unexpected int value binary value: %v, expected: %v", data, []byte{bstio.BinarySizeUint64, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
		}

		buf.Reset()
		c.Reset(ComposerOptions{})

		if err = c.WriteInt(math.MinInt64); err != nil {
			t.Fatalf("writing int failed: %v", err)
		}

		data = buf.Bytes()[1:]
		if len(data) != 9 {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		if !bytes.Equal(data, []byte{bstio.BinarySizeUint64, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) {
			t.Fatalf("unexpected int value binary value: %v, expected: %v", data, []byte{bstio.BinarySizeUint64, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		}

		buf.Reset()
		c.Reset(ComposerOptions{})
	})
}

func TestComposerArray(t *testing.T) {
	buf := &bytes.Buffer{}
	t.Run("Bool", func(t *testing.T) {
		buf.Reset()

		c, err := NewComposer(buf, bsttype.ArrayOf(bsttype.Boolean()), ComposerOptions{})
		if err != nil {
			t.Fatalf("creating composer failed: %v", err)
		}

		for i := 0; i < 10; i++ {
			b := i%2 == 0
			if err = c.WriteBoolean(b); err != nil {
				t.Fatalf("writing bool failed: %v", err)
			}
		}

		if err = c.Close(); err != nil {
			t.Fatalf("closing composer failed: %v", err)
		}

		data := buf.Bytes()
		// The data should be:
		// 0x00 - data header.
		// 0x1 - size flag for array
		// 10 - length
		// 0b01010101 0b00000001
		if len(data) != 5 {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		if !bytes.Equal(data, []byte{0x00, 0x1, 10, 0b01010101, 0b00000010}) {
			t.Fatalf("unexpected bool value binary value: %v, expected: %v", data, []byte{0x00, 0x1, 10, 0b01010101, 0b00000011})
		}

		buf.Reset()
		c.Reset(ComposerOptions{})
	})

	t.Run("ManualLength", func(t *testing.T) {
		t.Run("Failure", func(t *testing.T) {
			buf.Reset()

			_, err := NewComposer(buf, bsttype.ArrayOf(bsttype.Boolean()), ComposerOptions{Length: -1})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})

		t.Run("Success", func(t *testing.T) {
			buf.Reset()

			c, err := NewComposer(buf, bsttype.ArrayOf(bsttype.Boolean()), ComposerOptions{Length: 10})
			if err != nil {
				t.Fatalf("creating composer failed: %v", err)
			}

			for i := 0; i < 10; i++ {
				b := i%2 == 0
				if err = c.WriteBoolean(b); err != nil {
					t.Fatalf("writing bool failed: %v", err)
				}
			}

			data := buf.Bytes()
			// The data should be:
			// 0x00 - data header.
			// 0x1 - size flag for array
			// 10 - length
			// 0b01010101 0b00000001
			if len(data) != 5 {
				t.Fatalf("unexpected number of bytes written: %d", len(data))
			}

			if !bytes.Equal(data, []byte{0x00, 0x1, 10, 0b01010101, 0b00000010}) {
				t.Fatalf("unexpected bool value binary value: %v, expected: %v", data, []byte{0x00, 0x1, 10, 0b01010101, 0b00000011})
			}

			buf.Reset()
			if err = c.Close(); err != nil {
				t.Fatalf("closing composer failed: %v", err)
			}
		})
	})

	t.Run("Timestamp", func(t *testing.T) {
		buf.Reset()

		c, err := NewComposer(buf, bsttype.ArrayOf(bsttype.Timestamp()), ComposerOptions{Length: 2})
		if err != nil {
			t.Fatalf("creating composer failed: %v", err)
		}

		ts := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 2; i++ {
			if err = c.WriteTimestamp(ts); err != nil {
				t.Fatalf("writing timestamp failed: %v", err)
			}
		}

		data := buf.Bytes()
		// The data should be:
		// 0x00 - data header.
		// 0x1 - size flag for array
		// 2 - length
		// int64 value of timestamp in nanoseconds x 2
		if len(data) != 1+2+2*8 {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		tsBin := bstio.MarshalInt64(ts.UTC().UnixNano(), false)
		buf.Reset()
		buf.WriteByte(0x00)
		buf.Write([]byte{0x1, 2})
		buf.Write(tsBin)
		buf.Write(tsBin)

		if !bytes.Equal(data, buf.Bytes()) {
			t.Fatalf("unexpected timestamp value binary value: %v, expected: %v", data, buf.Bytes())
		}

		buf.Reset()
	})
}

func TestComposerMap(t *testing.T) {
	buf := &bytes.Buffer{}

	values := []struct {
		key   byte
		value bool
	}{
		{'a', true},
		{'b', false},
		{'c', true},
		{'d', false},
		{'e', true},
		{'f', false},
		{'g', true},
		{'h', false},
		{'i', true},
		{'j', false},
	}

	t.Run("Bool", func(t *testing.T) {
		buf.Reset()

		c, err := NewComposer(buf, bsttype.MapTypeOf(bsttype.Uint8(), bsttype.Boolean(), false, false), ComposerOptions{Length: 10})
		if err != nil {
			t.Fatalf("creating composer failed: %v", err)
		}

		// Let's write 10 pairs of rune : bool
		for _, v := range values {
			// Write Key
			if err = c.WriteUint8(v.key); err != nil {
				t.Fatalf("writing string failed: %v", err)
			}

			// Write Value
			if err = c.WriteBoolean(v.value); err != nil {
				t.Fatalf("writing bool failed: %v", err)
			}
		}

		data := buf.Bytes()
		// The data should be:
		// 0x00 - data header.
		// 0x1 - size flag for map
		// 10 - length
		// 'a', 0x01
		// 'b', 0x00
		// 'c', 0x01
		// 'd', 0x00
		// 'e', 0x01
		// 'f', 0x00
		// 'g', 0x01
		// 'h', 0x00
		// 'i', 0x01
		// 'j', 0x00
		if len(data) != 1+2+2*(10) {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		if !bytes.Equal(data, []byte{0x00, 0x1, 10, 'a', 0x01, 'b', 0x00, 'c', 0x01, 'd', 0x00, 'e', 0x01, 'f', 0x00, 'g', 0x01, 'h', 0x00, 'i', 0x01, 'j', 0x00}) {
			t.Fatalf("unexpected map value binary value: %v, expected: %v", data, []byte{0x00, 0x1, 10, 'a', 0x01, 'b', 0x00, 'c', 0x01, 'd', 0x00, 'e', 0x01, 'f', 0x00, 'g', 0x01, 'h', 0x00, 'i', 0x01, 'j', 0x00})
		}

		buf.Reset()
		c.Reset(ComposerOptions{})
	})
}

func TestComposerStruct(t *testing.T) {
	buf := &bytes.Buffer{}

	t.Run("MultiBool", func(t *testing.T) {
		st := bsttype.Struct{
			Fields: []bsttype.StructField{
				{Index: 1, Name: "a", Type: bsttype.Uint8()},
				{Index: 2, Name: "b", Type: bsttype.Boolean()},
				{Index: 3, Name: "c", Type: bsttype.Boolean(), Descending: true},
				{Index: 4, Name: "d", Type: bsttype.Boolean()},
				{Index: 5, Name: "e", Type: bsttype.String()},
			},
		}

		buf.Reset()
		defer buf.Reset()
		c, err := NewComposer(buf, &st, ComposerOptions{})
		if err != nil {
			t.Fatalf("creating composer failed: %v", err)
		}

		// a:
		if err = c.WriteUint8(1); err != nil {
			t.Fatalf("writing uint8 failed: %v", err)
		}

		// b:
		if err = c.WriteBoolean(true); err != nil {
			t.Fatalf("writing bool failed: %v", err)
		}

		// c:
		if err = c.WriteBoolean(false); err != nil {
			t.Fatalf("writing bool failed: %v", err)
		}

		// d:
		if err = c.WriteBoolean(true); err != nil {
			t.Fatalf("writing bool failed: %v", err)
		}

		// e:
		if err = c.WriteString("test"); err != nil {
			t.Fatalf("writing string failed: %v", err)
		}

		if !c.IsDone() {
			t.Fatalf("composer is not done")
		}

		data := buf.Bytes()
		// The data should be:
		// 0x00 - data header.
		// 0x01 - a
		// 0b00000111 - b, c (desc), d
		// 0x01 - size of length of string
		// 0x04 - length of string
		// 't', 'e', 's', 't'
		if len(data) != 1+1+1+1+1+4 {
			t.Fatalf("unexpected number of bytes written: %d", len(data))
		}

		if !bytes.Equal(data, []byte{0x00, 0x01, 0b00000111, 0x01, 0x04, 't', 'e', 's', 't'}) {
			t.Fatalf("unexpected struct value binary value: %v, expected: %v", data, []byte{0x00, 0x01, 0b00000111, 0x01, 0x04, 't', 'e', 's', 't'})
		}

		buf.Reset()
		c.Reset(ComposerOptions{})
	})

	t.Run("Compatibility", func(t *testing.T) {
		t.Run("Embedded", func(t *testing.T) {
			st := bsttype.Struct{
				Fields: []bsttype.StructField{
					{Index: 1, Name: "a", Type: bsttype.Uint8()},
					{Index: 2, Name: "b", Type: &bsttype.Struct{
						Fields: []bsttype.StructField{
							{Index: 1, Name: "c", Type: bsttype.Boolean()},
						},
					}},
					{Index: 4, Name: "d", Type: bsttype.Boolean()},
				},
			}

			defer buf.Reset()
			c, err := NewComposer(buf, &st, ComposerOptions{CompatibilityMode: true})
			if err != nil {
				t.Fatalf("creating composer failed: %v", err)
			}

			// a:
			if err = c.WriteUint8(8); err != nil {
				t.Fatalf("writing uint8 failed: %v", err)
			}

			// b:
			err = c.WriteStruct(func(xc *Composer) error {
				// c:
				if err = xc.WriteBoolean(true); err != nil {
					t.Fatalf("writing bool failed: %v", err)
				}

				if !xc.IsDone() {
					t.Fatalf("composer is not done")
				}
				return nil
			})
			if err != nil {
				t.Fatalf("creating sub composer failed: %v", err)
			}

			// d:
			if err = c.WriteBoolean(true); err != nil {
				t.Fatalf("writing bool failed: %v", err)
			}

			if !c.IsDone() {
				t.Fatalf("composer is not done")
			}
			if err = c.Close(); err != nil {
				t.Fatalf("closing composer failed: %v", err)
			}

			data := buf.Bytes()
			expected := []byte{
				// Header
				0b00000010, // Compatibility mode on.
				// Struct Header
				0x01, // Max Index Binary Size
				0x02, // Max Index Value
				// Field 'a'
				// Compatibility Field Header:
				0x01, // Field Index Binary Size
				0x01, // Field Index Value
				0x01, // Field Length Binary Size
				0x1,  // Field Length
				// Value:
				0x8,
				// Field 'b'
				// Compatibility Field Header:
				0x01, // Field Index Binary Size
				0x02, // Field Index Value
				0x01, // Field Length Binary Size
				0x06, // Field Length
				// Value:
				// Struct Header:
				0x00, // Max Index Binary Size and Value
				// Field 'c'
				// Compatibility Field Header:
				0x01, // Field Index Binary Size
				0x01, // Field Index Value
				0x01, // Field Length Binary Size
				0x01, // Field Length
				// Value:
				0b00000001,
				// Field 'd'
				// Compatibility Field Header:
				0x01, // Field Index Binary Size
				0x04, // Field Index Value
				0x01, // Field Length Binary Size
				0x01, // Field Length
				// Value:
				0b00000001,
			}
			if !bytes.Equal(data, expected) {
				t.Fatalf("unexpected struct value binary value: %v, expected: %v, diff: \n%v", data, expected, diff.DiffBytes(data, expected))
			}

			buf.Reset()
			c.Reset(ComposerOptions{})
		})

		t.Run("Simple", func(t *testing.T) {
			tp := &bsttype.Struct{
				Fields: []bsttype.StructField{
					{
						Name:  "ID",
						Index: 1,
						Type:  bsttype.Uint(),
					},
					{
						Name:  "Name",
						Index: 2,
						Type:  bsttype.String(),
					},
					{
						Name:  "Age",
						Index: 3,
						Type:  bsttype.Uint8(),
					},
					{
						Name:  "DayOfBirth",
						Index: 4,
						Type:  &bsttype.DateTime{},
					},
					{
						Name:  "Country",
						Index: 5,
						Type:  bsttype.String(),
					},
				},
			}

			c, err := NewComposer(buf, tp, ComposerOptions{CompatibilityMode: true})
			if err != nil {
				t.Fatalf("creating composer failed: %v", err)
			}

			if err = c.WriteUint(9); err != nil {
				t.Fatalf("writing uint failed: %v", err)
			}

			if err = c.WriteString("John"); err != nil {
				t.Fatalf("writing string failed: %v", err)
			}

			if err = c.WriteUint8(30); err != nil {
				t.Fatalf("writing uint8 failed: %v", err)
			}

			// 010000000e8a609d0000000000ffff
			birthDate := time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
			if err = c.WriteDateTime(birthDate); err != nil {
				t.Fatalf("writing date failed: %v", err)
			}

			if err = c.WriteString("Germany"); err != nil {
				t.Fatalf("writing string failed: %v", err)
			}

			if err = c.Close(); err != nil {
				t.Fatalf("closing composer failed: %v", err)
			}

			data := buf.Bytes()

			expected := []byte{
				// Header:
				0b00000010, // Compatibility mode on
				// Struct Compatibility header
				0x01,                     // Max Index binary size
				byte(len(tp.Fields) - 1), // Max Index
				// Field ID:
				// Compatibility Index:
				0x01, // Index binary integer size
				0x01, // Index
				0x01, // Field Binary length integer Size
				0x02, // Field Binary length
				// Value:
				0x01, // Value binary size
				0x09, // Value
				// Field Name:
				// Compatibility Index:
				0x01, // Index binary integer size
				0x02, // Index
				0x01, // Field Binary length integer Size
				0x06, // Field Binary length
				// Value:
				0x01, // Value binary size
				0x04, // Name length
				'J', 'o', 'h', 'n',
				// Field Age:
				// Compatibility Index:
				0x01, // Index binary integer size
				0x03, // Index
				0x01, // Field Binary length integer Size
				0x01, // Field Binary length
				// Value:
				byte(30), // Value
				// Field DayOfBirth:
				// Compatibility Index:
				0x01,     // Index binary integer size
				0x04,     // Index
				0x01,     // Field Binary length integer Size
				byte(15), // Field Binary length
				// Value:
				0x01, 0x00, 0x00, 0x00, 0x0e, 0x8a, 0x60,
				0x9d, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff,
				// Field Country:
				// Compatibility Index:
				0x01, // Index binary integer size
				0x05, // Index
				0x01, // Field Binary length integer Size
				0x09, // Field Binary length
				// Value:
				0x01, // Value binary size
				0x07, // Value length
				'G', 'e', 'r', 'm', 'a', 'n', 'y',
			}

			if !bytes.Equal(data, expected) {
				t.Fatalf("unexpected binary value: %v, expected: %v, \n%v", data, expected, diff.DiffBytes(expected, data))
			}
		})
	})
}

func TestComposerNamed(t *testing.T) {
	buf := &bytes.Buffer{}

	t.Run("NotResolved", func(t *testing.T) {
		nt := &bsttype.Named{
			Name:   "test",
			Module: "testing",
		}

		m := &bsttype.Module{
			Name: "testing",
			Definitions: []bsttype.ModuleDefinition{
				{
					Name: "test",
					Type: bsttype.Uint8(),
				},
			},
		}
		mds := bsttype.GetSharedModules()
		if err := mds.Add(m); err != nil {
			t.Fatalf("adding module failed: %v", err)
		}

		defer bsttype.PutSharedModules(mds)

		c, err := NewComposer(buf, nt, ComposerOptions{Modules: mds})
		if err != nil {
			t.Fatalf("creating composer failed: %v", err)
		}

		if err = c.WriteUint8(8); err != nil {
			t.Fatalf("writing uint8 failed: %v", err)
		}

		if err = c.Close(); err != nil {
			t.Fatalf("closing composer failed: %v", err)
		}

		data := buf.Bytes()
		// The data should be:
		// 0b00000000 - data header
		// 0x08       - named uint8 value
		if len(data) != 2 {
			t.Fatalf("unexpected number of bytes written: %d, expected: %d", len(data), 2)
		}

		expected := []byte{
			// Data header
			0b00000000,
			// Named Uint8 value
			0x08,
		}

		if !bytes.Equal(data, expected) {
			t.Fatalf("unexpected struct value binary value: %v, expected: %v", data, expected)
		}
		buf.Reset()
	})

	t.Run("NoModules", func(t *testing.T) {
		nt := &bsttype.Named{
			Name:   "test",
			Module: "testing",
			// Type: undefined on purpose
		}

		_, err := NewComposer(buf, nt, ComposerOptions{})
		if err == nil {
			t.Fatalf("creating composer should have failed")
		}
	})

	t.Run("BasicValid", func(t *testing.T) {
		nt := &bsttype.Named{
			Name:   "test",
			Module: "testing",
			Type:   bsttype.Uint8(),
		}

		c, err := NewComposer(buf, nt, ComposerOptions{})
		if err != nil {
			t.Fatalf("creating composer failed: %v", err)
		}

		if err = c.WriteUint8(8); err != nil {
			t.Fatalf("writing uint8 failed: %v", err)
		}

		if err = c.Close(); err != nil {
			t.Fatalf("closing composer failed: %v", err)
		}

		data := buf.Bytes()
		// The data should be:
		// 0b00000000 - data header
		// 0x08       - named uint8 value
		if len(data) != 2 {
			t.Fatalf("unexpected number of bytes written: %d, expected: %d", len(data), 2)
		}

		expected := []byte{
			// Data header
			0b00000000,
			// Named Uint8 value
			0x08,
		}

		if !bytes.Equal(data, expected) {
			t.Fatalf("unexpected struct value binary value: %v, expected: %v", data, expected)
		}
		buf.Reset()
	})

	t.Run("EmbedModules", func(t *testing.T) {
		nt := &bsttype.Named{
			Name:   "test",
			Module: "testing",
			Type:   bsttype.Uint(),
		}

		c, err := NewComposer(buf, nt, ComposerOptions{EmbedType: true})
		if err != nil {
			t.Fatalf("creating composer failed: %v", err)
		}

		if err = c.WriteUint(8); err != nil {
			t.Fatalf("writing uint failed: %v", err)
		}

		if err = c.Close(); err != nil {
			t.Fatalf("closing composer failed: %v", err)
		}

		data := buf.Bytes()
		// The data should be:
		// -----------------------------------------------------------
		// HEADER
		// -----------------------------------------------------------
		// 0b00010001              - data header
		// -----------------------------------------------------------
		// MODULES
		// -----------------------------------------------------------
		// 0x01                    - modules length binary size
		// 0x01                    - modules length
		// 0x01                    - module name length binary size
		// 0x07	                   - module name length
		// "testing"               - module name
		// 0x01                    - module definitions length binary size
		// 0x01                    - module definitions length
		// 0x01                    - module definition name length binary size
		// 0x04                    - module definition name length
		// "test"                  - module definition name
		// byte(bsttype.KindUint)  - module definition type kind
		// ------------------------------------------------------------
		// EMBEDDED TYPE
		// ------------------------------------------------------------
		// byte(bsttype.KindNamed) - embed  type kind
		// 0x01                    - embed type module name length binary size
		// 0x07	                   - embed type module name length
		// "testing"               - embed type module name
		// 0x01                    - embed type name length binary size
		// 0x04                    - embed type name length
		// "test"                  - embed type name
		// -----------------------------------------------------------
		// VALUE
		// -----------------------------------------------------------
		// 0x01                   - value binary size
		// 0x08                   - value
		expected := []byte{
			// Data header
			0b00010001,
			// Modules length binary size
			0x01,
			// Modules length
			0x01,
			// Module name length binary size
			0x01,
			// Module name length
			0x07,
			// Module name length
			't', 'e', 's', 't', 'i', 'n', 'g',
			// Module definitions length binary size
			0x01,
			// Module definitions length
			0x01,
			// Module definition name length binary size
			0x01,
			// Module definition name length
			0x04,
			// Module definition name
			't', 'e', 's', 't',
			// Module definition type
			byte(bsttype.KindUint),
			// Embed type kind
			byte(bsttype.KindNamed),
			// Module definitions length
			0x01,
			// Module name length binary size
			0x07,
			// Module name length
			't', 'e', 's', 't', 'i', 'n', 'g',
			// Module definition name length binary size
			0x01,
			// Module definition name length
			0x04,
			// Module definition name
			't', 'e', 's', 't',
			// Value binary size
			0x01,
			// Value
			0x08,
		}
		if !bytes.Equal(data, expected) {
			t.Fatalf("unexpected binary value: %v, expected: %v, \n%v", data, expected, diff.DiffBytes(expected, data))
		}
		buf.Reset()
	})
}
