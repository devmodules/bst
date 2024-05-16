package bst

import (
	"time"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// WriteDuration writes a duration value to the composer.
func (x *Composer) WriteDuration(v time.Duration) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindDuration {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindDuration),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		n, err := x.writeFieldHeader(x.w, x.fieldIndex(), 8)
		if err != nil {
			return err
		}

		x.bytesWritten += n
	}

	// 4. Write the duration value.
	n, err := bstio.WriteInt64(x.w, v.Nanoseconds(), x.elemDesc)
	if err != nil {
		return bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write duration")
	}

	x.bytesWritten += n

	// 5. Mark the element as written.
	if err = x.finishElem(); err != nil {
		return err
	}
	return nil
}

// ReadDuration reads the duration value from the extractor.
func (x *Extractor) ReadDuration() (time.Duration, error) {
	if x.err != nil {
		return 0, x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return 0, bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	// 2. Check if current element is still in range.
	if x.index > x.maxIndex {
		return 0, bsterr.Err(bsterr.CodeOutOfBounds, "buffIndex out of bounds")
	}

	// 3. Verify if current element matches the expected type.
	if x.elemType.Kind() != bsttype.KindDuration {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindDuration),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Read the duration value.
	v, n, err := bstio.ReadInt64(x.r, x.elemDesc)
	x.bytesRead += n
	if err != nil {
		return 0, err
	}

	x.finishElem()

	return time.Duration(v), nil
}
