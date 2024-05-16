package bst

import (
	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// WriteInt8 writes an int8 value to the composer.
func (x *Composer) WriteInt8(v int8) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindInt8 {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindInt8),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		n, err := x.writeFieldHeader(x.w, x.fieldIndex(), 1)
		if err != nil {
			return err
		}

		x.bytesWritten += n
	}

	// 4. Write the value.
	n, err := bstio.WriteInt8(x.w, v, x.elemDesc)
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

// WriteInt16 writes an int16 value to the composer.
func (x *Composer) WriteInt16(v int16) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindInt16 {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindInt16),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		n, err := x.writeFieldHeader(x.w, x.fieldIndex(), 2)
		if err != nil {
			return err
		}

		x.bytesWritten += n
	}

	// 4. Write the value.
	n, err := bstio.WriteInt16(x.w, v, x.elemDesc)
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

// WriteInt32 writes an int32 value to the composer.
func (x *Composer) WriteInt32(v int32) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindInt32 {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", x.elemType.Kind()),
				bsterr.D("actual", bsttype.KindInt32),
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
	n, err := bstio.WriteInt32(x.w, v, x.elemDesc)
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

// WriteInt64 writes an int64 value to the composer.
func (x *Composer) WriteInt64(v int64) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindInt64 {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindInt64),
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
	n, err := bstio.WriteInt64(x.w, v, x.elemDesc)
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

// WriteInt writes an integer value to the composer.
func (x *Composer) WriteInt(v int) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindInt {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindInt),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		binSize := 8
		if !x.opts.Comparable {
			binSize = bstio.UintBinarySize(uint(v))
		}

		n, err := x.writeFieldHeader(x.w, x.fieldIndex(), uint(binSize))
		if err != nil {
			return err
		}

		x.bytesWritten += n
	}

	// 4. Write the value.
	n, err := bstio.WriteInt(x.w, v, x.elemDesc, x.opts.Comparable)
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

//
// Extract Signed Integers
//

// ReadInt8 reads the int8 value from the extractor.
func (x *Extractor) ReadInt8() (int8, error) {
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
	if x.elemType.Kind() != bsttype.KindInt8 {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindInt8),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Read the 8-bit signed integers.
	v, n, err := bstio.ReadInt8(x.r, x.elemDesc)
	if err != nil {
		return 0, err
	}

	x.bytesRead += n

	x.finishElem()
	return v, nil
}

// ReadInt16 reads the int16 value from the extractor.
func (x *Extractor) ReadInt16() (int16, error) {
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
	if x.elemType.Kind() != bsttype.KindInt16 {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindInt16),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Read the 16-bit signed integers.
	v, n, err := bstio.ReadInt16(x.r, x.elemDesc)
	if err != nil {
		return 0, err
	}

	x.bytesRead += n

	x.finishElem()
	return v, nil
}

// ReadInt32 reads the int32 value from the extractor.
func (x *Extractor) ReadInt32() (int32, error) {
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
	if x.elemType.Kind() != bsttype.KindInt32 {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindInt32),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Read the 32-bit signed integers.
	v, n, err := bstio.ReadInt32(x.r, x.elemDesc)
	if err != nil {
		return 0, err
	}

	x.bytesRead += n
	x.finishElem()
	return v, nil
}

// ReadInt64 reads the int64 value from the extractor.
func (x *Extractor) ReadInt64() (int64, error) {
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
	if x.elemType.Kind() != bsttype.KindInt64 {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindInt64),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Read the 64-bit signed integers.
	v, n, err := bstio.ReadInt64(x.r, x.elemDesc)
	if err != nil {
		return 0, err
	}

	x.bytesRead += n

	x.finishElem()
	return v, nil
}

// ReadInt reads the int value from the extractor.
func (x *Extractor) ReadInt() (int, error) {
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
	if x.elemType.Kind() != bsttype.KindInt {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindInt),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Read the int value.
	v, n, err := bstio.ReadInt(x.r, x.elemDesc, x.opts.Comparable)
	if err != nil {
		return 0, err
	}

	x.bytesRead += n
	x.finishElem()

	return v, nil
}

// Int reads the int value from the extractor.
// Current field type must be one of the following types:
//   - int8
//   - int16
//   - int32
//   - int64
//   - int
func (x *Extractor) Int() (int64, error) {
	if x.err != nil {
		return 0, x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return 0, bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	var res int64
	switch x.elemType.Kind() {
	case bsttype.KindInt8:
		v, n, err := bstio.ReadInt8(x.r, x.elemDesc)
		if err != nil {
			return 0, err
		}
		x.bytesRead += n
		res = int64(v)
	case bsttype.KindInt16:
		v, n, err := bstio.ReadInt16(x.r, x.elemDesc)
		if err != nil {
			return 0, err
		}
		x.bytesRead += n
		res = int64(v)
	case bsttype.KindInt32:
		v, n, err := bstio.ReadInt32(x.r, x.elemDesc)
		if err != nil {
			return 0, err
		}
		x.bytesRead += n
		res = int64(v)
	case bsttype.KindInt64:
		v, n, err := bstio.ReadInt64(x.r, x.elemDesc)
		if err != nil {
			return 0, err
		}
		x.bytesRead += n
		res = v
	case bsttype.KindInt:
		v, n, err := bstio.ReadInt(x.r, x.elemDesc, x.opts.Comparable)
		if err != nil {
			return 0, err
		}
		x.bytesRead += n
		res = int64(v)
	default:
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", []bsttype.Kind{bsttype.KindInt8, bsttype.KindInt16, bsttype.KindInt32, bsttype.KindInt64, bsttype.KindInt}),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	x.finishElem()
	return res, nil
}
