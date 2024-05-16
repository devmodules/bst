package bstskip

import (
	"io"

	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// SkipMap skips the map type.
func SkipMap(br io.ReadSeeker, x *bsttype.Map, options bstio.ValueOptions) (int64, error) {
	return mapSkipFunc(x)(br, options)
}

func mapSkipFunc(x *bsttype.Map) SkipFunc {
	return func(br io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
		// 1. Decode the number of entries.
		length, n, err := bstio.ReadUint(br, options.Descending)
		if err != nil {
			return int64(n), err
		}
		bytesSkipped := int64(n)

		// 2. Initialize empty map key and value.
		ek, ev := SkipFuncOf(x.Key.Type), SkipFuncOf(x.Value.Type)

		// 3. Iterate over the map entries and skip each entry.
		var skipped int64
		for i := uint(0); i < length; i++ {
			skipped, err = ek(br, options)
			if err != nil {
				return bytesSkipped + skipped, err
			}
			bytesSkipped += skipped

			skipped, err = ev(br, options)
			if err != nil {
				return bytesSkipped + skipped, err
			}
			bytesSkipped += skipped
		}
		return bytesSkipped, nil
	}
}
