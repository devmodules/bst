package bst

import (
	"time"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// WriteDateTime writes a datetime value to the composer.
func (x *Composer) WriteDateTime(v time.Time) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	dt, ok := x.elemType.(*bsttype.DateTime)
	if !ok {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindDateTime),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		n, err := x.writeFieldHeader(x.w, x.fieldIndex(), 15)
		if err != nil {
			return err
		}

		x.bytesWritten += n
	}

	// 4. Write the datetime value.
	n, err := bstio.WriteDateTime(x.w, v, x.elemDesc, dt.Location())
	if err != nil {
		return err
	}

	x.bytesWritten += n

	// 5. Mark the element as written.
	if err = x.finishElem(); err != nil {
		return err
	}
	return nil
}

// ReadDateTime reads the datetime value from the extractor.
func (x *Extractor) ReadDateTime() (time.Time, error) {
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
	dt, ok := x.elemType.(*bsttype.DateTime)
	if !ok {
		return time.Time{}, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindDateTime),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Read the datetime value.
	v, n, err := bstio.ReadDateTime(x.r, x.opts.Descending, dt.Location())
	x.bytesRead += n
	if err != nil {
		return time.Time{}, err
	}
	x.finishElem()
	return v, nil
}
