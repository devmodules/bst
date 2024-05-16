package bstskip

import (
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// SkipStruct skips the struct type.
// NOTE: Add compatibility mode in the input.
//
//	This changes the behavior of the function.
func SkipStruct(br io.ReadSeeker, x *bsttype.Struct, options bstio.ValueOptions) (int64, error) {
	if options.CompatibilityMode {
		return structSkipCompatibilityStruct(x)(br, options)
	}
	return structSkipFunc(x)(br, options)
}

func structSkipCompatibilityStruct(x *bsttype.Struct) SkipFunc {
	return func(br io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
		// 1. Read struct header.
		maxIndex, n, err := bstio.ReadUint(br, false)
		if err != nil {
			return int64(n), bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to read struct header")
		}
		total := int64(n)

		var (
			n64         int64
			bytesToSkip uint
		)
		for i := uint(0); i < maxIndex; i++ {
			n64, err = bstio.SkipUint(br, false)
			if err != nil {
				return total, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to read compatibility field header index")
			}
			total += n64

			bytesToSkip, n, err = bstio.ReadUint(br, false)
			if err != nil {
				return total, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to read compatibility field header bytes")
			}
			total += int64(n)

			if bytesToSkip > 0 {
				_, err = br.Seek(int64(bytesToSkip), io.SeekCurrent)
				if err != nil {
					return total, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to skip compatibility field header bytes")
				}
				total += int64(bytesToSkip)
			}
		}
		return total, nil
	}
}

func structSkipFunc(x *bsttype.Struct) SkipFunc {
	return func(br io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
		var (
			total, n int64
			err      error
			boolPos  byte
		)

		for fi, f := range x.Fields {
			if f.Type.Kind() == bsttype.KindBoolean {
				prev, ok := x.PreviewPrevElemType(fi)
				if !ok || boolPos == 0 || (ok && prev.Kind() == bsttype.KindBoolean) {
					n, err = bstio.SkipUint8Value(br)
					if err != nil {
						return total, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to read bool value")
					}
					total += n
				}
				boolPos++

				if boolPos == 8 {
					boolPos = 0
				}
				continue
			}

			n, err = SkipFuncOf(f.Type)(br, options)
			if err != nil {
				return total, err
			}
			total += n
		}
		return total, nil
	}
}
