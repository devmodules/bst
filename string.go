package bst

import (
	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// WriteString writes a string value to the composer.
func (x *Composer) WriteString(v string) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindString {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindString),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		n, err := x.writeFieldHeader(x.w, x.fieldIndex(), bstio.StringBinarySize(v, x.opts.Comparable))
		if err != nil {
			return err
		}

		x.bytesWritten += n
	}

	// 4. Write the value.
	n, err := bstio.WriteString(x.w, v, x.elemDesc, x.opts.Comparable)
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

// ReadString reads the string value from the extractor.
func (x *Extractor) ReadString() (string, error) {
	if x.err != nil {
		return "", x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return "", bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	// 2. Check if current element is still in range.
	if x.index > x.maxIndex {
		return "", bsterr.Err(bsterr.CodeOutOfBounds, "buffIndex out of bounds")
	}

	// 3. Verify if current element matches the expected type.
	if x.elemType.Kind() != bsttype.KindString {
		return "", bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindString),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Read the string value.
	v, n, err := bstio.ReadString(x.r, x.elemDesc, x.opts.Comparable)
	if err != nil {
		return "", err
	}

	x.bytesRead += n

	x.finishElem()
	return v, nil
}
