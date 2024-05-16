package bst

import (
	"bytes"
	"testing"
	"time"

	"github.com/devmodules/bst/bsttype"
	"github.com/devmodules/bst/internal/iopool"
)

func TestExtractorNamed(t *testing.T) {
	t.Run("BasicHeadless", func(t *testing.T) {
		nt := &bsttype.Named{
			Name:   "test",
			Module: "testing",
			Type:   bsttype.Uint8(),
		}

		// The data should be:
		// 0x08       - named uint8 value
		data := []byte{0x08}

		r := iopool.GetReadSeeker(data)
		defer iopool.ReleaseReadSeeker(r)

		e, err := NewExtractor(r, ExtractorOptions{ExpectedType: nt, Headless: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var v uint8
		v, err = e.ReadUint8()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != 0x08 {
			t.Fatalf("unexpected value: %v", v)
		}

		if e.BytesRead() != 1 {
			t.Fatalf("unexpected bytes read: %v", e.BytesRead())
		}

		e.Close()
	})

	t.Run("BasicNotResolved", func(t *testing.T) {
		nt := bsttype.Named{
			Name:   "test",
			Module: "testing",
			// Type: undefined on purpose
		}
		t.Run("NoModules", func(t *testing.T) {
			r := iopool.GetReadSeeker([]byte{0x00, 0x08})
			tp := nt
			_, err := NewExtractor(r, ExtractorOptions{ExpectedType: &tp})
			if err == nil {
				t.Fatalf("expected error")
			}
		})

		t.Run("WithModules", func(t *testing.T) {
			m := bsttype.GetSharedModule()
			m.Name = "testing"
			m.Definitions = append(m.Definitions, bsttype.ModuleDefinition{
				Name: "test",
				Type: bsttype.Uint8(),
			})

			mds := bsttype.GetSharedModules()
			if err := mds.Add(m); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer mds.Free()

			r := iopool.GetReadSeeker([]byte{0x00, 0x08})
			defer iopool.ReleaseReadSeeker(r)

			x, err := NewExtractor(r, ExtractorOptions{ExpectedType: &nt, Modules: mds})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var v uint8
			if x.Next() {
				v, err = x.ReadUint8()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if v != 0x08 {
					t.Fatalf("unexpected value: %v", v)
				}
			}

			if x.BytesRead() != 2 {
				t.Fatalf("unexpected bytes read: %v", x.BytesRead())
			}

			x.Close()
		})

		t.Run("EmbedModules", func(t *testing.T) {
			// The data composition:
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
			data := []byte{
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
			r := iopool.GetReadSeeker(data)
			defer iopool.ReleaseReadSeeker(r)

			tp := nt
			x, err := NewExtractor(r, ExtractorOptions{ExpectedType: &tp})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var v uint
			if x.Next() {
				v, err = x.ReadUint()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if v != 0x08 {
					t.Fatalf("unexpected value: %v", v)
				}
			}
			if x.BytesRead() != len(data) {
				t.Fatalf("unexpected bytes read: %v", x.BytesRead())
			}
			x.Close()
		})
	})
}

func TestExtractorStruct(t *testing.T) {
	tp := bsttype.Struct{
		Fields: []bsttype.StructField{
			{
				Index:      1,
				Name:       "ID",
				Descending: false,
				Type:       bsttype.Uint(),
			},
			{
				Index:      2,
				Name:       "Name",
				Descending: false,
				Type:       bsttype.String(),
			},
			{
				Index:      3,
				Name:       "NullableNamed",
				Descending: false,
				Type: bsttype.NullableOf(&bsttype.Named{
					Module: "testing",
					Name:   "name",
					// Type: undefined on purpose
				}),
			},
			{
				Index: 4,
				Name:  "NamedArray",
				Type: bsttype.ArrayOf(&bsttype.Named{
					Module: "testing",
					Name:   "name",
				}),
			},
			{
				Index:      5,
				Name:       "NamedMap",
				Descending: false,
				Type: bsttype.MapTypeOf(
					bsttype.Uint8(),
					&bsttype.Named{Module: "testing", Name: "name"},
					false,
					false,
				),
			},
		},
	}

	md := &bsttype.Modules{
		List: []*bsttype.Module{
			{
				Name: "testing",
				Definitions: []bsttype.ModuleDefinition{
					{
						Name: "name",
						Type: &tp,
					},
				},
			},
		},
	}

	if err := md.Resolve(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("Headless", func(t *testing.T) {
		t.Run("NoNamedRecursion", func(t *testing.T) {
			data := []byte{
				// Field ID:
				0x01, // Uint binary size
				0x08, // Uint

				// Field Name:
				0x01, // String binary size
				0x04, // String length
				't', 'e', 's', 't',

				// Field NullableNamed:
				0x00, // Null flag

				// Field NamedArray:
				0x00, // Array binary size

				// Field NamedMap:
				0x00, // Map binary size
			}

			r := iopool.GetReadSeeker(data)
			defer iopool.ReleaseReadSeeker(r)

			x, err := NewExtractor(r, ExtractorOptions{ExpectedType: &tp, Modules: md, Headless: true})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			type temp struct {
				ID            uint
				Name          string
				NullableNamed *temp
			}

			var v temp
			// Field ID:
			if x.Next() {
				v.ID, err = x.ReadUint()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if v.ID != 0x08 {
					t.Fatalf("unexpected value: %v, wanted: %d", v.ID, 0x08)
				}
			}

			// Field Name:
			if x.Next() {
				v.Name, err = x.ReadString()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if v.Name != "test" {
					t.Fatalf("unexpected Name value: %v, wanted: %s", v.Name, "test")
				}
			}

			// Field NullableNamed:
			if x.Next() {
				isNull, err := x.IsNull()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if !isNull {
					t.Fatal("unexpected non-null value")
				}
			}
			if x.Next() {
				err = x.ReadArray(func(ax *Extractor) error {
					if ax.Length() != 0 {
						t.Fatalf("unexpected array length: %v", ax.Length())
					}

					ax.Close()
					return nil
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			if x.Next() {
				err = x.ReadMap(func(mx *Extractor) error {
					if mx.Length() != 0 {
						t.Fatalf("unexpected map length: %v", mx.Length())
					}
					mx.Close()
					return nil
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			x.Close()
		})

		t.Run("NamedRecursion", func(t *testing.T) {
			data := []byte{
				// Field ID:
				0x01, // Uint binary size
				0x08, // Uint

				// Field Name:
				0x01, // String binary size
				0x04, // String length
				't', 'e', 's', 't',

				// Field NullableNamed:
				0x01, // NotNull flag
				// Field NullableNamed.ID:
				0x01, // Uint binary size
				0x09, // Uint

				// Field NullableNamed.Name:
				0x01, // String binary size
				0x05, // String length
				'n', 'a', 'm', 'e', 'd',

				// Field NullableNamed.NullableNamed:
				0x00, // Null flag

				// Field NullableNamed.NamedArray:
				0x00, // Array binary size

				// Field NullableNamed.NamedMap:
				0x00, // Map binary size

				// Field NamedArray:
				0x01, // Array binary size
				0x01, // Array length

				// Field NamedArray[0].ID:
				0x01, // Uint binary size
				0x10, // Uint

				// Field NamedArray[0].Name:
				0x01, // String binary size
				0x06, // String length
				'n', 'a', 'm', 'e', 'd', '0',

				// Field NamedArray[0].NullableNamed:
				0x00, // Null flag

				// Field NamedArray[0].NamedArray:
				0x00, // Array binary size

				// Field NamedArray[0].NamedMap:
				0x00, // Map binary size

				// Field NamedMap:
				0x01, // Map binary size
				0x01, // Map length

				// NamedMap.Key:
				0x11, // Uint8 value
				// NamedMap.Value
				// NamedMap.Value.ID:
				0x01, // Uint binary size
				0x11, // Uint
				// NamedMap.Value.Name:
				0x01, // String binary size
				0x06, // String length
				'n', 'a', 'm', 'e', 'd', '1',
				// NamedMap.Value.NullableNamed:
				0x00, // Null flag
				// NamedMap.Value.NamedArray:
				0x00, // Array binary size
				// NamedMap.Value.NamedMap:
				0x00, // Map binary size
			}

			r := iopool.GetReadSeeker(data)
			defer iopool.ReleaseReadSeeker(r)

			x, err := NewExtractor(r, ExtractorOptions{
				ExpectedType: &tp,
				Modules:      md,
				Headless:     true,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			type temp struct {
				ID            uint
				Name          string
				NullableNamed *temp
			}

			var v temp
			// Field ID:
			if x.Next() {
				v.ID, err = x.ReadUint()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if v.ID != 0x08 {
					t.Fatalf("unexpected value: %v, wanted: %d", v.ID, 0x08)
				}
			}

			// Field Name:
			if x.Next() {
				v.Name, err = x.ReadString()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if v.Name != "test" {
					t.Fatalf("unexpected Name value: %v, wanted: %s", v.Name, "test")
				}
			}

			// Field NullableNamed:
			if x.Next() {
				isNull, err := x.IsNull()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if isNull {
					t.Fatal("unexpected null value")
				}

				err = x.ReadStruct(func(sx *Extractor) error {
					var vt temp
					// Field ID:
					if sx.Next() {
						vt.ID, err = sx.ReadUint()
						if err != nil {
							t.Fatalf("unexpected error: %v", err)
						}
						if vt.ID != 0x09 {
							t.Fatalf("unexpected value: %v, wanted: %d", vt.ID, 0x09)
						}
					}

					// Field Name:
					if sx.Next() {
						vt.Name, err = sx.ReadString()
						if err != nil {
							t.Fatalf("unexpected error: %v", err)
						}
						if vt.Name != "named" {
							t.Fatalf("unexpected Name value: %v, wanted: %s", vt.Name, "named")
						}
					}

					// Field NullableNamed:
					if sx.Next() {
						isNull, err := sx.IsNull()
						if err != nil {
							t.Fatalf("unexpected error: %v", err)
						}

						if !isNull {
							t.Fatal("unexpected non-null value")
						}
					}

					if sx.Next() {
						err = sx.ReadArray(func(ax *Extractor) error {
							if ax.Length() != 0 {
								t.Fatalf("unexpected array length: %d", ax.Length())
							}
							ax.Close()
							return nil
						})

						if err != nil {
							t.Fatalf("unexpected error: %v", err)
						}
					}

					if sx.Next() {
						err = sx.ReadMap(func(mx *Extractor) error {
							if mx.Length() != 0 {
								t.Fatalf("unexpected map length: %d", mx.Length())
							}
							mx.Close()
							return nil
						})
					}

					if sx.BytesRead() != 12 {
						t.Fatalf("unexpected bytes read: %d, wanted: %d", sx.BytesRead(), 12)
					}
					sx.Close()
					return nil
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			// Field NamedArray:
			if x.Next() {
				err = x.ReadArray(func(ax *Extractor) error {
					if ax.Length() != 1 {
						t.Fatalf("unexpected array length: %d", ax.Length())
					}

					for ax.Next() {
						var tmp temp
						err = ax.ReadStruct(func(asx *Extractor) error {
							// Field ID:
							if asx.Next() {
								tmp.ID, err = asx.ReadUint()
								if err != nil {
									t.Fatalf("unexpected error: %v", err)
								}
								if tmp.ID != 0x10 {
									t.Fatalf("unexpected value: %v, wanted: %d", tmp.ID, 0x10)
								}
							}

							// Field Name:
							if asx.Next() {
								tmp.Name, err = asx.ReadString()
								if err != nil {
									t.Fatalf("unexpected error: %v", err)
								}
								if tmp.Name != "named0" {
									t.Fatalf("unexpected Name value: %v, wanted: %s", tmp.Name, "named0")
								}
							}

							// Field NullableNamed:
							if asx.Next() {
								var isNull bool
								isNull, err = asx.IsNull()
								if err != nil {
									t.Fatalf("unexpected error: %v", err)
								}

								if !isNull {
									t.Fatal("unexpected null value")
								}
							}

							// Field NamedArray
							if asx.Next() {
								err = asx.ReadArray(func(ax *Extractor) error {
									if ax.Length() != 0 {
										t.Fatalf("unexpected array length: %d", ax.Length())
									}
									ax.Close()
									return nil
								})
								if err != nil {
									t.Fatalf("unexpected error: %v", err)
								}
							}

							// Field NamedMap
							if asx.Next() {
								err = asx.ReadMap(func(mx *Extractor) error {
									if mx.Length() != 0 {
										t.Fatalf("unexpected map length: %d", mx.Length())
									}
									mx.Close()
									return nil
								})
								if err != nil {
									t.Fatalf("unexpected error: %v", err)
								}
							}
							if asx.BytesRead() != 13 {
								t.Fatalf("unexpected bytes read: %d, wanted: %d", asx.BytesRead(), 13)
							}
							return nil
						})

						if err != nil {
							t.Fatalf("unexpected error: %v", err)
						}
					}
					if ax.BytesRead() != 15 {
						t.Fatalf("unexpected bytes read: %d, wanted: %d", ax.BytesRead(), 15)
					}
					return nil
				})
			}

			// Field NamedMap:
			if x.Next() {
				err = x.ReadMap(func(mx *Extractor) error {
					if mx.Length() != 1 {
						t.Fatalf("unexpected map length: %d", mx.Length())
					}
					if mx.Next() {
						// Map Key:
						var k uint8
						k, err = mx.ReadUint8()
						if err != nil {
							t.Fatalf("unexpected error: %v", err)
						}
						if k != 0x11 {
							t.Fatalf("unexpected key value: %v, wanted: %d", k, 0x11)
						}

						var tmp temp
						err = mx.ReadStruct(func(msx *Extractor) error {
							// Field ID:
							if msx.Next() {
								tmp.ID, err = msx.ReadUint()
								if err != nil {
									t.Fatalf("unexpected error: %v", err)
								}
								if tmp.ID != 0x11 {
									t.Fatalf("unexpected value: %v, wanted: %d", tmp.ID, 0x11)
								}
							}

							// Field Name:
							if msx.Next() {
								tmp.Name, err = msx.ReadString()
								if err != nil {
									t.Fatalf("unexpected error: %v", err)
								}
								if tmp.Name != "named1" {
									t.Fatalf("unexpected Name value: %v, wanted: %s", tmp.Name, "named1")
								}
							}

							// Field NullableNamed:
							if msx.Next() {
								var isNull bool
								isNull, err = msx.IsNull()
								if err != nil {
									t.Fatalf("unexpected error: %v", err)
								}

								if !isNull {
									t.Fatal("unexpected not-null value")
								}
							}

							// Field NamedArray
							if msx.Next() {
								err = msx.ReadArray(func(ax *Extractor) error {
									if ax.Length() != 0 {
										t.Fatalf("unexpected array length: %d", ax.Length())
									}
									ax.Close()
									return nil
								})
								if err != nil {
									t.Fatalf("unexpected error: %v", err)
								}

							}

							// Field NamedMap
							if msx.Next() {
								err = msx.ReadMap(func(mx *Extractor) error {
									if mx.Length() != 0 {
										t.Fatalf("unexpected map length: %d", mx.Length())
									}
									mx.Close()
									return nil
								})
								if err != nil {
									t.Fatalf("unexpected error: %v", err)
								}
							}
							return nil
						})
					}
					mx.Close()
					return nil
				})
			}

			if x.BytesRead() != len(data) {
				t.Fatalf("unexpected bytes read: %d, wanted: %d", x.BytesRead(), len(data))
			}

			x.Close()
		})
	})

	t.Run("Compatibility", func(t *testing.T) {
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

		t.Run("HeadlessFull", func(t *testing.T) {
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

			x, err := NewExtractor(r, ExtractorOptions{ExpectedType: st, Headless: true, CompatibilityMode: true})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if x.Next() {
				var id uint
				id, err = x.ReadUint()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if id != 0x11 {
					t.Fatalf("unexpected ID value: %d, wanted: %d", id, 0x11)
				}
			}

			if x.Next() {
				var name string
				name, err = x.ReadString()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if name != "testing" {
					t.Fatalf("unexpected name value: %s, wanted: %s", name, "testing")
				}
			}
			if x.Next() {
				var ts time.Time
				ts, err = x.ReadTimestamp()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				expected := time.Date(2022, 07, 07, 16, 22, 00, 00, time.UTC)
				if !ts.Equal(expected) {
					t.Fatalf("unexpected timestamp value: %v, wanted: %v", ts, expected)
				}
			}
			if x.Next() {
				var u uint8
				u, err = x.ReadUint8()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if u != 0xFF {
					t.Fatalf("unexpected uint8 value: %d, wanted: %d", u, 0xFF)
				}
			}
			x.Close()
		})

		lt := &bsttype.Struct{
			Fields: []bsttype.StructField{
				{
					Name:  "ID",
					Type:  bsttype.Uint(),
					Index: 1,
				},
				// This type is missing the name field.
				// {
				// 	Name: "Name",
				// 	Type: bsttype.String(),
				// 	Index: 2,
				// },
				{
					Name:  "Timestamp",
					Type:  bsttype.Timestamp(),
					Index: 3,
				},
				// This field is also missing in the expected type.
				// {
				// 	Name:  "Uint8",
				// 	Type:  bsttype.Uint8(),
				// 	Index: 4,
				// },
				{
					Name:  "NotInTheEmbedded",
					Type:  bsttype.Uint8(),
					Index: 5,
				},
			},
		}

		t.Run("LimitedEmbedded", func(t *testing.T) {
			buf := &bytes.Buffer{}
			// 1. Write Header
			buf.WriteByte(0b00000011) // Embedded Type and compatibility mode.

			// 2. Write the structure type.
			buf.WriteByte(byte(bsttype.KindStruct))
			_, err := st.WriteType(buf)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// 3. Write field values.
			// This test provides embedded struct data in compatibility mode,
			// where the expected type does not have all the fields which exists in the embedded one.
			data := []byte{
				// Struct Compatibility Header:
				0x01, // Max Index binary size
				0x03, // Max Index value
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

			_, err = buf.Write(data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			r := iopool.GetReadSeeker(buf.Bytes())
			defer iopool.ReleaseReadSeeker(r)

			x, err := NewExtractor(r, ExtractorOptions{ExpectedType: lt})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			hasNext := x.Next()
			if !hasNext {
				t.Fatalf("expected field ID to be present")
			}
			var id uint
			id, err = x.ReadUint()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != 0x11 {
				t.Fatalf("unexpected ID value: %d, wanted: %d", id, 0x11)
			}

			hasNext = x.Next()
			if !hasNext {
				t.Fatalf("expected field Timestamp to be present")
			}
			var ts time.Time
			ts, err = x.ReadTimestamp()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expected := time.Date(2022, 07, 07, 16, 22, 00, 00, time.UTC)
			if !ts.Equal(expected) {
				t.Fatalf("unexpected timestamp value: %v, wanted: %v", ts, expected)
			}

			hasNext = x.Next()
			if hasNext {
				t.Fatalf("expected field NotInTheEmbedded to be absent")
			}
			x.Close()
		})

		t.Run("LimitedNotEmbedded", func(t *testing.T) {
			// 3. Write field values.
			// This test provides embedded struct data in compatibility mode,
			// where the expected type does not have all the fields which exists in the embedded one.
			data := []byte{
				// Header:
				0b00000010, // Compatibility mode only.
				// Struct Compatibility Header:
				0x01, // Max Index binary size
				0x03, // Max Index value
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

			x, err := NewExtractor(r, ExtractorOptions{ExpectedType: lt})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			hasNext := x.Next()
			if !hasNext {
				t.Fatalf("expected field ID to be present")
			}
			var id uint
			id, err = x.ReadUint()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != 0x11 {
				t.Fatalf("unexpected ID value: %d, wanted: %d", id, 0x11)
			}

			hasNext = x.Next()
			if !hasNext {
				t.Fatalf("expected field Timestamp to be present")
			}
			var ts time.Time
			ts, err = x.ReadTimestamp()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expected := time.Date(2022, 07, 07, 16, 22, 00, 00, time.UTC)
			if !ts.Equal(expected) {
				t.Fatalf("unexpected timestamp value: %v, wanted: %v", ts, expected)
			}

			hasNext = x.Next()
			if hasNext {
				t.Fatalf("expected field NotInTheEmbedded to be absent")
			}
			x.Close()
		})
	})
}
