package bst

import (
	"time"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// WriteTimestamp writes a timestamp value to the composer.
func (x *Composer) WriteTimestamp(v time.Time) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindTimestamp {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindTimestamp),
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

	// 4. Write the timestamp value.
	n, err := bstio.WriteInt64(x.w, v.UTC().UnixNano(), x.elemDesc)
	if err != nil {
		return bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write timestamp")
	}

	x.bytesWritten += n

	// 5. Mark the element as written.
	if err = x.finishElem(); err != nil {
		return err
	}
	return nil
}

// ReadTimestamp reads the timestamp value from the extractor.
func (x *Extractor) ReadTimestamp() (time.Time, error) {
	if x.err != nil {
		return time.Time{}, x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return time.Time{}, bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	// 2. Check if current element is still in range.
	if x.index > x.maxIndex {
		return time.Time{}, bsterr.Err(bsterr.CodeOutOfBounds, "buffIndex out of bounds")
	}

	// 3. Verify if current element matches the expected type.
	if x.elemType.Kind() != bsttype.KindTimestamp {
		return time.Time{}, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindTimestamp),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Read the timestamp value.
	v, n, err := bstio.ReadInt64(x.r, x.elemDesc)
	x.bytesRead += n
	if err != nil {
		return time.Time{}, err
	}

	x.finishElem()
	return time.Unix(0, v).UTC(), nil
}
