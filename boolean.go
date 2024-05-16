package bst

import (
	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// WriteBoolean writes a bool value to the composer.
func (x *Composer) WriteBoolean(v bool) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindBoolean {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindBoolean),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		// 3.1. When the fields headers are written, no boolean compaction could be provided.
		// Thus, each boolean is written as a separate field - separate byte.
		n, err := x.writeFieldHeader(x.w, x.fieldIndex(), 1)
		if err != nil {
			return err
		}

		x.bytesWritten += n

		// 3.2. If the value is descending, inverse it.
		buf := byte(0)
		if v {
			buf = 1
		}
		if x.elemDesc {
			buf = ^buf
		}

		// 3.3. Write a buffered byte.
		if err = bstio.WriteByte(x.w, buf); err != nil {
			return err
		}

		x.bytesWritten++

		// 3.3. Mark the element as written.
		if err = x.finishElem(); err != nil {
			return err
		}
		return nil
	}

	// 3. Write the binary value to the boolean buffer.
	//    For booleans the positive value is defined as '1'
	if v || x.elemDesc {
		x.boolBuf |= 1 << x.boolBufPos
	}

	// 4. Increment the boolean buffer position.
	x.boolBufPos++

	// 5. Check if the boolean buffer is full or if the next element is not a boolean.
	//    If so, write the boolean buffer to the writer and resetWithRoot it.
	//	  Otherwise, increment the boolean buffer position.
	e, ok := x.previewNextElem()
	if x.boolBufPos == 7 || !ok || (ok && e.Kind() != bsttype.KindBoolean) {
		if err := bstio.WriteByte(x.w, x.boolBuf); err != nil {
			return bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write bool")
		}

		x.bytesWritten++
		x.boolBuf = 0
		x.boolBufPos = 0
	}

	// 6. Mark the element as written.
	if err := x.finishElem(); err != nil {
		return err
	}
	return nil
}

// ReadBoolean reads the bool value from the extractor.
func (x *Extractor) ReadBoolean() (bool, error) {
	if x.err != nil {
		return false, x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return false, bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	// 2. Verify if current element matches the expected type.
	if x.elemType.Kind() != bsttype.KindBoolean {
		return false, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindBoolean),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. Read the bool value.
	prev, ok := x.previewPrevElem()
	if !ok || x.boolBufPosition == 0 || (ok && prev.Kind() != bsttype.KindBoolean) {
		buf, err := bstio.ReadByte(x.r)
		if err != nil {
			return false, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to read bool value")
		}

		x.boolBuf = buf
	}

	// 4. Extract the bool value.
	v := x.boolBuf&(1<<uint(x.boolBufPosition)) != 0

	// 5. Increment current buffer position.
	x.boolBufPosition++

	// 6. Check if current buffer is full.
	if x.boolBufPosition == 8 {
		// 6.1. Reset current buffer and its position.
		x.boolBuf = 0x00
		x.boolBufPosition = 0
	}

	if x.elemDesc {
		v = !v
	}
	x.finishElem()
	return v, nil
}
