package bst

import (
	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// WriteBytes writes a byte slice value to the composer.
// The byte slice for descending values will be modified,
// so it should be copied if it is needed later.
func (x *Composer) WriteBytes(v []byte) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	bt, ok := x.elemType.(*bsttype.Bytes)
	if !ok {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindBytes),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		n, err := x.writeFieldHeader(x.w, x.fieldIndex(), bstio.BytesBinarySize(bt.FixedSize, v, x.elemDesc, x.opts.Comparable))
		if err != nil {
			return err
		}

		x.bytesWritten += n
	}

	// 4. Write the value.
	n, err := bstio.WriteBytes(x.w, bt.FixedSize, v, x.elemDesc, x.opts.Comparable)
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

// ReadBytes reads the 'bytes' elem value from the extractor.
func (x *Extractor) ReadBytes() ([]byte, error) {
	if x.err != nil {
		return nil, x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return nil, bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	// 2. Verify if current element matches the expected type.
	bt, ok := x.elemType.(*bsttype.Bytes)
	if !ok {
		return nil, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindBytes),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. Read the bytes value.
	v, n, err := bstio.ReadBytes(x.r, bt.FixedSize, x.elemDesc, x.opts.Comparable)
	x.bytesRead += n
	if err != nil {
		return nil, err
	}
	x.finishElem()
	return v, nil
}
