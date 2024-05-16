package bstskip

import (
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// SkipOneOf skips the value of bsttype.OneOf.
func SkipOneOf(rs io.ReadSeeker, tp *bsttype.OneOf, o bstio.ValueOptions) (int64, error) {
	return oneOfSkipFunc(tp)(rs, o)
}

func oneOfSkipFunc(tp *bsttype.OneOf) SkipFunc {
	return func(rs io.ReadSeeker, o bstio.ValueOptions) (int64, error) {
		// 1. Read the buffIndex.
		idx, bytesRead, err := bstio.ReadOneOfIndex(rs, tp.IndexBytes, o.Descending)
		if err != nil {
			return int64(bytesRead), err
		}
		bytesSkipped := int64(bytesRead)

		// 2. Match the buffIndex to the one of element.
		var elem bsttype.Type
		for i := range tp.Elements {
			if tp.Elements[i].Index == idx {
				elem = tp.Elements[i].Type
				break
			}
		}

		// 3. If the buffIndex did not match, return an error.
		if elem == nil {
			return bytesSkipped, bsterr.Err(bsterr.CodeTypeConstraintViolation, "oneOfType buffIndex doesn't match to the elements")
		}

		// 4. Initialize empty value.
		v := SkipFuncOf(elem)

		// 5. Skip the value.
		var n int64
		n, err = v(rs, o)
		if err != nil {
			return bytesSkipped + n, err
		}

		// 6. Return the number of bytes skipped.
		return bytesSkipped + n, nil
	}
}
