package bst

import (
	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// WriteEnumIndex writes an enum value to the composer.
func (x *Composer) WriteEnumIndex(index int) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	et, ok := x.elemType.(*bsttype.Enum)
	if !ok {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindEnum),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. Verify if the buffIndex is within the enum range.
	var found bool
	for _, e := range et.Elements {
		if e.Index == uint(index) {
			found = true
			break
		}
	}

	// 4. Return an error if the buffIndex is not found.
	if !found {
		return bsterr.Err(bsterr.CodeInvalidValue, "invalid enum value").
			WithDetails(
				bsterr.D("value", index),
			)
	}

	// 3. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		n, err := x.writeFieldHeader(x.w, x.fieldIndex(), uint(et.ValueBytes))
		if err != nil {
			return err
		}

		x.bytesWritten += n
	}

	// 5. Write the enum value.
	n, err := bstio.WriteEnumIndex(x.w, index, et.ValueBytes, x.elemDesc)
	if err != nil {
		return err
	}

	x.bytesWritten += n

	// 7. Mark the element as written.
	if err = x.finishElem(); err != nil {
		return err
	}

	return nil
}

// ReadEnumIndex reads the enum value from the extractor.
func (x *Extractor) ReadEnumIndex() (uint, error) {
	if x.err != nil {
		return 0, x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return 0, bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	// 2. Verify if current element matches the expected type.
	et, ok := x.elemType.(*bsttype.Enum)
	if !ok {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindEnum),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. Read the enum value.
	index, n, err := bstio.ReadEnumIndex(x.r, et.ValueBytes, x.elemDesc)
	x.bytesRead += n
	if err != nil {
		return 0, err
	}

	x.finishElem()
	return uint(index), nil
}
