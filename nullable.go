package bst

import (
	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// WriteNull writes a null value to the composer.
func (x *Composer) WriteNull() error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindNullable {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindNullable),
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

	// 4. Prepare the null value.
	bt := bstio.NullableIsNull
	if x.elemDesc {
		bt = bstio.NullableIsNullDesc
	}

	// 4. Write the null value.
	if err := bstio.WriteByte(x.w, bt); err != nil {
		return bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write null")
	}
	x.bytesWritten++

	// 5. Mark the element as written.
	if err := x.finishElem(); err != nil {
		return err
	}
	return nil
}

// WriteNotNull writes a not null header to the composer.
func (x *Composer) WriteNotNull() error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	nt, ok := x.elemType.(*bsttype.Nullable)
	if !ok {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindNullable),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		x.setFieldBuffer()
	}

	// 4. Prepare the binary value for the not null header.
	bt := bstio.NullableIsNotNull
	if x.elemDesc {
		bt = bstio.NullableIsNotNullDesc
	}

	// 5. Write the not null value.
	if err := bstio.WriteByte(x.w, bt); err != nil {
		return bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write not null")
	}

	x.bytesWritten++

	// 6. Dereference the nullable type.
	x.elemType = nt.Elem()
	return nil
}

// IsNull checks if the element value is null.
// It works only if the current element type is NullableType.
// This method reads from the
func (x *Extractor) IsNull() (bool, error) {
	if x.err != nil {
		return false, x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return false, bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	// 2. Verify if current element matches the expected type.
	nt, ok := x.elemType.(*bsttype.Nullable)
	if !ok {
		return false, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindNullable),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. Read the null value.
	v, err := bstio.ReadNullableFlag(x.r, x.opts.Descending)
	if err != nil {
		return false, err
	}

	x.bytesRead += 1

	// 4. Finish nullable type.
	switch v {
	case bstio.NullableIsNotNull:
		// 4.1. A 1-bit indicates that the value is not-null.
		// Dereference the nullable value elem type.
		if x.embedType == x.opts.ExpectedType {
			x.elemType, x.err = x.derefType(nt.Type)
			if x.err != nil {
				return false, x.err
			}
			x.embed.elemType = x.elemType
		} else {
			var nx *bsttype.Nullable
			nx, ok = x.opts.ExpectedType.(*bsttype.Nullable)
			if !ok {
				return false, bsterr.Err(bsterr.CodeInvalidType, "invalid type expected element type").
					WithDetails(
						bsterr.D("expected", bsttype.KindNullable),
						bsterr.D("actual", x.opts.ExpectedType.Kind()),
					)
			}
			x.elemType, x.err = x.derefType(nx.Type)
			if x.err != nil {
				return false, x.err
			}
			x.embed.elemType, x.err = x.derefType(nt.Type)
			if x.err != nil {
				return false, x.err
			}
		}
		return false, nil
	case bstio.NullableIsNull:
		// A 0-bit indicates that the value is null.
		x.finishElem()
		return true, nil
	default:
		return false, bsterr.Err(bsterr.CodeInvalidValue, "invalid nullable flag value")
	}
}
