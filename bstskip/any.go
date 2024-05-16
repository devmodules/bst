package bstskip

import (
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// SkipAny skips the value of bsttype.Any.
func SkipAny(rs io.ReadSeeker, o bstio.ValueOptions) (int64, error) {
	rt, n, err := bsttype.ReadType(rs, false)
	if err != nil {
		return int64(n), err
	}

	v := SkipFuncOf(rt)
	var skipped int64
	skipped, err = v(rs, o)
	if err != nil {
		return int64(n) + skipped, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to skip value").
			WithDetail("type", rt.Kind())
	}
	return int64(n) + skipped, nil
}
