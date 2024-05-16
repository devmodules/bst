package bst

import (
	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// WriteFloat32 writes a float32 value to the composer.
func (x *Composer) WriteFloat32(v float32) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindFloat32 {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindFloat32),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		n, err := x.writeFieldHeader(x.w, x.fieldIndex(), 4)
		if err != nil {
			return err
		}

		x.bytesWritten += n
	}

	// 4. Write the value.
	n, err := bstio.WriteFloat32(x.w, v, x.elemDesc)
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

// WriteFloat64 writes a float64 value to the composer.
func (x *Composer) WriteFloat64(v float64) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindFloat64 {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindFloat64),
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

	// 4. Write the value.
	n, err := bstio.WriteFloat64(x.w, v, x.elemDesc)
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

// ReadFloat32 reads the float32 value from the extractor.
func (x *Extractor) ReadFloat32() (float32, error) {
	if x.err != nil {
		return 0, x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return 0, bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	// 2. Verify if current element matches the expected type.
	if x.elemType.Kind() != bsttype.KindFloat32 {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindFloat32),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. Read the float32 value.
	v, n, err := bstio.ReadFloat32(x.r, x.elemDesc)
	x.bytesRead += n
	if err != nil {
		return 0, err
	}
	x.finishElem()
	return v, nil
}

// ReadFloat64 reads the float64 value from the extractor.
func (x *Extractor) ReadFloat64() (float64, error) {
	if x.err != nil {
		return 0, x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return 0, bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	// 2. Verify if current element matches the expected type.
	if x.elemType.Kind() != bsttype.KindFloat64 {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindFloat64),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. Read the float64 value.
	v, n, err := bstio.ReadFloat64(x.r, x.elemDesc)
	x.bytesRead += n
	if err != nil {
		return 0, err
	}
	x.finishElem()
	return v, nil
}
