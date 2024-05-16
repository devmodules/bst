package bstskip

import (
	"testing"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
	"github.com/devmodules/bst/internal/iopool"
)

func TestSkipStruct(t *testing.T) {
	t.Run("CompatibilityMode", func(t *testing.T) {
		data := []byte{
			// Struct Compatibility Header:
			0x01, // Max Index binary size
			0x04, // Max Index value
			// Field ID:
			// Compatibility Index:
			0x01, // Index binary size
			0x01, // Index :1
			0x01, // Field Binary Size
			0x02, // Field Binary length
			// Value:
			0x01, // ID Binary size
			0x11, // ID value
			// Field Name:
			// Compatibility Index:
			0x01, // Index binary size
			0x02, // Index :2
			0x01, // Field Binary Size
			0x09, // Field Binary length
			// Value:
			0x01,                              // Name Binary size
			0x07,                              // Name length
			't', 'e', 's', 't', 'i', 'n', 'g', // Name value
			// Field Timestamp:
			// Compatibility Index:
			0x01, // Index binary size
			0x03, // Index :3
			0x01, // Field Binary Size
			0x08, // Field Binary length
			// Value:
			0x16 | 0x80, 0xff, 0x98, 0x8d, 0x2c, 0x7f, 0x90, 0x00,
			// Field Uint8:
			// Compatibility Index:
			0x01, // Index binary size
			0x04, // Index :4
			0x01, // Field Binary Size
			0x01, // Field Binary length
			// Value:
			0xFF, // Uint8 Binary size
		}

		r := iopool.GetReadSeeker(data)
		defer iopool.ReleaseReadSeeker(r)

		st := &bsttype.Struct{
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
					Name:  "Timestamp",
					Index: 3,
					Type:  bsttype.Timestamp(),
				},
				{
					Name:  "Uint8",
					Index: 4,
					Type:  bsttype.Uint8(),
				},
			},
		}

		n, err := SkipStruct(r, st, bstio.ValueOptions{
			CompatibilityMode: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		if int(n) != len(data) {
			t.Fatalf("Expected %d, got %d", len(data), n)
		}
	})
}
