package bst

import (
	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// WriteUint8 writes an uint8 value to the composer.
func (x *Composer) WriteUint8(v uint8) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindUint8 {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindUint8),
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
	n, err := bstio.WriteUint8(x.w, v, x.elemDesc)
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

// WriteUint16 writes an uint16 value to the composer.
func (x *Composer) WriteUint16(v uint16) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindUint16 {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindUint16),
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
	n, err := bstio.WriteUint16(x.w, v, x.elemDesc)
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

// WriteUint32 writes an uint32 value to the composer.
func (x *Composer) WriteUint32(v uint32) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindUint32 {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindUint32),
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
	n, err := bstio.WriteUint32(x.w, v, x.elemDesc)
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

// WriteUint64 writes an uint64 value to the composer.
func (x *Composer) WriteUint64(v uint64) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindUint64 {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindUint64),
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
	n, err := bstio.WriteUint64(x.w, v, x.elemDesc)
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

// WriteUint writes an uint value to the composer.
func (x *Composer) WriteUint(v uint) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindUint {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindUint),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		n, err := x.writeFieldHeader(x.w, x.fieldIndex(), uint(bstio.UintBinarySize(v)))
		if err != nil {
			return err
		}

		x.bytesWritten += n
	}

	// 4. Write the value.
	n, err := bstio.WriteUint(x.w, v, x.elemDesc)
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
// Extract Unsigned Integers
//

// ReadUint8 reads the uint8 value from the extractor.
func (x *Extractor) ReadUint8() (uint8, error) {
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
	if x.elemType.Kind() != bsttype.KindUint8 {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindUint8),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Read the 8-bit unsigned integer..
	v, n, err := bstio.ReadUint8(x.r, x.elemDesc)
	if err != nil {
		return 0, err
	}

	x.bytesRead += n
	x.finishElem()
	return v, nil
}

// ReadUint16 reads the uint16 value from the extractor.
func (x *Extractor) ReadUint16() (uint16, error) {
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
	if x.elemType.Kind() != bsttype.KindUint16 {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindUint16),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Read the 16-bit unsigned integer.
	v, n, err := bstio.ReadUint16(x.r, x.elemDesc)
	if err != nil {
		return 0, err
	}

	x.bytesRead += n
	x.finishElem()
	return v, nil
}

// ReadUint32 reads the uint32 value from the extractor.
func (x *Extractor) ReadUint32() (uint32, error) {
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
	if x.elemType.Kind() != bsttype.KindUint32 {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindUint32),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. Read the 32-bit unsigned integer.
	v, n, err := bstio.ReadUint32(x.r, x.elemDesc)
	if err != nil {
		return 0, err
	}

	x.bytesRead += n

	x.finishElem()
	return v, nil
}

// ReadUint64 reads the uint64 value from the extractor.
func (x *Extractor) ReadUint64() (uint64, error) {
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
	if x.elemType.Kind() != bsttype.KindUint64 {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindUint64),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Read the 64-bit unsigned integer.
	v, n, err := bstio.ReadUint64(x.r, x.elemDesc)
	if err != nil {
		return 0, err
	}
	x.bytesRead += n

	x.finishElem()
	return v, nil
}

// ReadUint reads the uint value from the extractor.
func (x *Extractor) ReadUint() (uint, error) {
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
	if x.elemType.Kind() != bsttype.KindUint {
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindUint),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Read varying size unsigned integer.
	v, n, err := bstio.ReadUint(x.r, x.elemDesc)
	if err != nil {
		return 0, err
	}

	x.bytesRead += n

	x.finishElem()
	return v, nil
}

// Uint reads the uint value from the extractor.
// Current field type must be one of the following types:
//   - uint8
//   - uint16
//   - uint32
//   - uint64
//   - uint
func (x *Extractor) Uint() (uint64, error) {
	if x.err != nil {
		return 0, x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return 0, bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	var res uint64
	switch x.elemType.Kind() {
	case bsttype.KindUint8:
		v, n, err := bstio.ReadUint8(x.r, x.elemDesc)
		if err != nil {
			return 0, err
		}
		x.bytesRead += n
		res = uint64(v)
	case bsttype.KindUint16:
		v, n, err := bstio.ReadUint16(x.r, x.elemDesc)
		if err != nil {
			return 0, err
		}
		x.bytesRead += n
		res = uint64(v)
	case bsttype.KindUint32:
		v, n, err := bstio.ReadUint32(x.r, x.elemDesc)
		if err != nil {
			return 0, err
		}
		x.bytesRead += n
		res = uint64(v)
	case bsttype.KindUint64:
		v, n, err := bstio.ReadUint64(x.r, x.elemDesc)
		if err != nil {
			return 0, err
		}
		x.bytesRead += n
		res = v
	case bsttype.KindUint:
		v, n, err := bstio.ReadUint(x.r, x.elemDesc)
		if err != nil {
			return 0, err
		}
		x.bytesRead += n
		res = uint64(v)
	default:
		return 0, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", []bsttype.Kind{bsttype.KindUint8, bsttype.KindUint16, bsttype.KindUint32, bsttype.KindUint64, bsttype.KindUint}),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	x.finishElem()
	return res, nil
}
