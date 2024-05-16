package bst

import (
	"errors"
	"io"
	"math"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bstskip"
	"github.com/devmodules/bst/bsttype"
	"github.com/devmodules/bst/internal/iopool"
)

// WriteArray writes an array value to the composer.
// The optional length argument can be used to specify the length of the array.
func (x *Composer) WriteArray(fn func(c *Composer) error, optLength int) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		x.setFieldBuffer()
	}

	// 3. Create a savepoint and resetWithRoot given composer.
	sp := *x

	// 4. Verify if current element matches expected type.
	at, ok := x.elemType.(*bsttype.Array)
	if !ok {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindArray),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 5. resetWithRoot current composer state.
	x.reset()

	// 6. Set up the length of the array.
	if optLength > 0 {
		x.maxIndex = optLength - 1
		x.definedLength = true
	}

	// 7. Initialize array composer.
	if err := x.initializeArrayComposer(at, false); err != nil {
		return err
	}

	// 8. Call input function.
	if err := fn(x); err != nil {
		return err
	}

	// 9. Verify if writing was completed
	if x.index <= x.maxIndex && !x.definedLength {
		return bsterr.Err(bsterr.CodeWritingFailed, "sub-composer didn't write all elements")
	}

	// 10. Close the array composer.
	if err := x.closeArray(at); err != nil {
		return err
	}

	// 11. Store the number of bytes written to array composer.
	bw := x.bytesWritten

	// 12. Restore the savepoint.
	*x = sp

	// 13. Increase the number of bytes written by the array composer.
	x.bytesWritten += bw

	// 14. Finish the element.
	if err := x.finishElem(); err != nil {
		return err
	}
	return nil
}

// ReadArray reads an array value from the extractor. It creates a sub-extractor which
// should be used for the element array type.
func (x *Extractor) ReadArray(fn func(x *Extractor) error) error {
	// 1. Check if any previous reading had failed.
	if x.err != nil {
		return x.err
	}

	// 2. Check if reading element value is already finished.
	if x.elemDone {
		return bsterr.Err(bsterr.CodeAlreadyRead, "element already read")
	}

	// 3. Create a snapshot of the current state.
	sp := *x

	// 3. Ensure that the element is an array.
	if x.elemType.Kind() != bsttype.KindArray {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type element to read").
			WithDetails(
				bsterr.D("expected", bsttype.KindArray),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Keep embedded and expected array types.
	xt := x.elemType
	et := x.embed.elemType

	// 5. Reset the extractor
	x.reset()

	// 6. Set up base type for the new extractor composer.
	x.opts.ExpectedType = xt
	x.embedType = et

	// 7. Initialize the extractor for the array.
	if err := x.initializeArray(); err != nil {
		return err
	}

	// 8. Execute the extraction function.
	if err := fn(x); err != nil {
		return err
	}

	// 9. Check if the array was fully extracted.
	if err := x.finishArray(); err != nil {
		return err
	}

	// 10. Keep the number of bytes read from the array.
	br := x.bytesRead

	// 11. Restore an extractor from the snapshot.
	*x = sp

	// 12. Update the number of bytes read.
	x.bytesRead += br

	// 13. Finish this element.
	x.finishElem()

	return nil
}

func (x *Extractor) initializeArray() error {
	tt, ok := x.embedType.(*bsttype.Array)
	if !ok {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to read").
			WithDetails(
				bsterr.D("expected", bsttype.KindArray),
				bsterr.D("actual", x.embedType.Kind()),
			)
	}

	// 1. Set up common array fields.
	x.index = -1
	x.embed.elemType = tt.Type
	x.elemType = tt.Type
	if et, ok := x.opts.ExpectedType.(*bsttype.Array); ok {
		x.elemType = et.Type
	}
	x.elemType, x.err = x.derefType(x.elemType)
	if x.err != nil {
		return x.err
	}
	x.embed.elemType, x.err = x.derefType(x.embed.elemType)
	if x.err != nil {
		return x.err
	}

	x.elemDesc = x.opts.Descending

	// 2.If the array is of fixed size we already know the length and directly start the extraction.
	if tt.FixedSize != 0 {
		x.maxIndex = int(tt.FixedSize) - 1
		return nil
	}

	// 3. If the array is of variable size, and the extractor is not in comparable format,
	//    we need to read the length of the array.
	if !x.opts.Comparable {
		// 3.1. Read the length of the array.
		ln, n, err := bstio.ReadUint(x.r, x.opts.Descending)
		if err != nil {
			return err
		}
		x.bytesRead += n

		// 3.2. Set the maximum index of the array.
		x.maxIndex = int(ln - 1)
		return nil
	}

	// 4. In the comparable format the length of the array is not known upfront.
	//    Array binary is terminated by a sequence of 0x02 and 0x01 bytes.
	//    At the beginning we need to read the raw bytes, and unescape all consecutive 0x00 0x03 bytes.
	//    Then we need to read elements of the array until we reach io.EOF.
	escape := bstio.ArrayEscapeAscending
	if x.opts.Descending {
		escape = bstio.ArrayEscapeDescending
	}

	// 5. Read the raw bytes of the array.
	data, n, err := bstio.ReadComparableBytesReader(x.r, x.opts.Descending, escape)
	if err != nil {
		return err
	}
	x.bytesRead += n

	// 6. Wrap the array bytes with a new reader.
	ar := iopool.GetReadSeeker(data)

	// 7. Find a number of elements in the array.
	//    NOTE: we don't know the length of the array, so we need to read the elements until we reach io.EOF.
	var ln int
	opts := bstio.ValueOptions{
		Descending:        x.opts.Descending,
		Comparable:        x.opts.Comparable,
		CompatibilityMode: x.opts.CompatibilityMode,
	}
	for {
		_, err = bstskip.SkipFuncOf(tt.Type)(ar, opts)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		ln++
	}

	// 8. Reset array reader to the beginning.
	_, err = ar.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// 9. Create a wrapped reader, which could be unwrapped at the end of the array extraction.
	//    NOTE: it is important to notice that comparable arrays need to be unwrapped.
	wr := iopool.WrapReader(ar)
	x.r = wr
	x.maxIndex = math.MaxInt

	return nil
}

func (x *Extractor) nextArrayElem() bool {
	// 1. If the previous elem was not done, then an error occurred.
	if !x.elemDone && x.index >= 0 {
		x.err = bsterr.Err(bsterr.CodeNotReadYet, "array element not extracted yet")
		return false
	}
	if x.baseDone {
		return false
	}

	// 2. Increase the index of the array elements.
	x.index++

	// 3. Check if all elements of the array were already extracted.
	if x.index > x.maxIndex {
		// 3.1. For comparable binaries, a reader was wrapped, thus we need to unwrap and set it back to extractor.
		if x.opts.Comparable {
			wr := x.r.(*iopool.SharedReadSeeker)
			x.r = wr.Root().(io.ReadSeeker)
			iopool.ReleaseReadSeeker(wr)
		}
		x.baseDone = true
		return false
	}

	// 4. Reset the done flag.
	x.elemDone = false
	return true
}

func (x *Extractor) finishArray() error {
	// 1. Check if the array is already done.
	if x.baseDone {
		return nil
	}

	// 2. Otherwise, try to get the valid type to skip the array elements.
	elem := x.opts.ExpectedType
	if !bsttype.TypesEqual(x.embed.elemType, elem) {
		elem = x.embedType
	}
	skipFn := bstskip.SkipFuncOf(elem)
	opts := bstio.ValueOptions{
		Descending:        x.opts.Descending,
		Comparable:        x.opts.Comparable,
		CompatibilityMode: x.opts.CompatibilityMode,
	}

	for x.index < x.maxIndex {
		// 3. Skip the array elements.
		n, err := skipFn(x.r, opts)
		if err != nil {
			return err
		}
		x.bytesRead += int(n)
		x.index++
	}

	// 3.1. For comparable binaries, a reader was wrapped, thus we need to unwrap and set it back to extractor.
	if x.opts.Comparable {
		wr := x.r.(*iopool.SharedReadSeeker)
		x.r = wr.Root().(io.ReadSeeker)
		iopool.ReleaseReadSeeker(wr)
	}

	x.elemDesc = true
	x.baseDone = true
	return nil
}

func (x *Extractor) finishArrayElem() {
	// Finishing the array element advances its index to the next one.
	x.elemDone = true

	if x.clearElemFn != nil {
		x.clearElemFn()
		x.clearElemFn = nil
	}
}

func (x *Extractor) previewPrevArrayElem() (bsttype.Type, bool) {
	if x.index == -1 {
		return nil, false
	}
	et := x.embedType.(*bsttype.Array)
	return et.Type, true
}
