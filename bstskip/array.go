package bstskip

import (
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// SkipArray skips the array value binary from the input reader.
func SkipArray(rs io.ReadSeeker, at *bsttype.Array, options bstio.ValueOptions) (int64, error) {
	return arraySkipFunc(at)(rs, options)
}

func arraySkipFunc(at *bsttype.Array) SkipFunc {
	return func(rs io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
		var (
			n   int64
			err error
		)
		length := at.FixedSize
		if !at.HasFixedSize() {
			var ni int
			length, ni, err = bstio.ReadUint(rs, options.Descending)
			if err != nil {
				return int64(ni), err
			}
			n += int64(ni)
		}

		if length == 0 {
			return n, nil
		}

		elem := at.Elem()
		switch elem.Kind() {
		case bsttype.KindUndefined:
			// Nothing to skip for undefined array type.
			return n, bsterr.Err(bsterr.CodeDecodingBinaryValue, "undefined array type")
		case bsttype.KindBoolean:
			// Boolean arrays are skipped differently as the number of bytes written is in fact
			// the number of elements divided by 8.
			bytesNo := (length + 7) >> 3
			_, err = rs.Seek(int64(bytesNo), io.SeekCurrent)
			if err != nil {
				return n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip array")
			}
			return n + int64(bytesNo), nil
		default:
			skipFunc := SkipFuncOf(elem)
			total := n
			for i := uint(0); i < length; i++ {
				n, err = skipFunc(rs, options)
				if err != nil {
					return total, err
				}
				total += n
			}
			return total, nil
		}
	}
}
