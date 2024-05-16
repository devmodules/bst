package bst

import (
	"errors"
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bstskip"
	"github.com/devmodules/bst/bsttype"
	"github.com/devmodules/bst/internal/iopool"
)

// WriteMap writes a map value to the composer. It creates a sub-composer which would be used
// as an argument to the input function.
func (x *Composer) WriteMap(fn func(c *Composer) error, optLength int) error {
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
	mt, ok := x.elemType.(*bsttype.Map)
	if !ok {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindMap),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 5. resetWithRoot current composer state.
	x.reset()

	// 6. Set up optional length if defined.
	if optLength > 0 {
		x.maxIndex = optLength - 1
		x.definedLength = true
	}

	// 7. Initialize map composer.
	if err := x.initializeMapComposer(mt, false); err != nil {
		return err
	}

	// 8. Call input function.
	if err := fn(x); err != nil {
		return err
	}

	// 9. Verify if writing was completed
	if x.index <= x.maxIndex && !x.definedLength {
		return bsterr.Err(bsterr.CodeWritingFailed, "not all expected elements in the map were written")
	}

	// 10. Close the map composer.
	if err := x.closeMap(); err != nil {
		return err
	}

	// 11. Store the number of bytes written to the map composer.
	bw := x.bytesWritten

	// 12. Restore the savepoint.
	*x = sp

	// 13. Increase the number of bytes written by the map composer.
	x.bytesWritten += bw

	// 14. Finish the element.
	if err := x.finishElem(); err != nil {
		return err
	}
	return nil
}

// ReadMap extracts and reads the map value. The input function
// determines how the map should be extracted.
func (x *Extractor) ReadMap(fn func(x *Extractor) error) error {
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

	// 4. Ensure that the element is of a map type.
	if x.elemType.Kind() != bsttype.KindMap {
		x.err = bsterr.Err(bsterr.CodeInvalidType, "invalid type to read").
			WithDetails(
				bsterr.D("expected", bsttype.KindMap),
				bsterr.D("actual", x.elemType.Kind()),
			)
		return x.err
	}

	// 5. Keep expected and embedded map types.
	xt := x.elemType
	et := x.embed.elemType

	// 6. Reset the state of the extractor.
	x.reset()

	// 7. Set up embedded and expected types.
	x.opts.ExpectedType = xt
	x.embedType = et

	// 8. Initialize the extractor for the map.
	if err := x.initializeMap(); err != nil {
		x.err = err
		return err
	}

	// 8. Execute the extraction function.
	if err := fn(x); err != nil {
		x.err = err
		return err
	}

	// 9. Check if the map was fully extracted.
	if err := x.finishMap(); err != nil {
		return err
	}

	// 10. Keep the number of bytes read.
	br := x.bytesRead

	// 11. Restore an extractor from the snapshot.
	*x = sp

	// 12. Update the number of bytes read.
	x.bytesRead += br

	// 13. Finish this element.
	x.finishElem()

	return nil
}

func (x *Extractor) initializeMap() error {
	// 1. Set up common map fields.
	bt := x.embedType.(*bsttype.Map)
	x.index = -1
	x.isKey = true
	switch {
	case x.opts.ExpectedType == nil || x.opts.ExpectedType == x.embedType:
		x.elemType, x.err = x.derefType(bt.Key.Type)
		if x.err != nil {
			return x.err
		}
		x.embed.elemType = x.elemType
	case x.opts.ExpectedType != x.embedType:
		mp, ok := x.opts.ExpectedType.(*bsttype.Map)
		if !ok {
			return bsterr.Err(bsterr.CodeInvalidType, "invalid type to read").
				WithDetails(
					bsterr.D("expected", bsttype.KindMap),
					bsterr.D("actual", x.opts.ExpectedType.Kind()),
				)
		}
		x.elemType, x.err = x.derefType(mp.Key.Type)
		if x.err != nil {
			return x.err
		}

		x.embed.elemType, x.err = x.derefType(bt.Key.Type)
		if x.err != nil {
			return x.err
		}
	}

	x.elemDesc = bt.Key.Descending

	// 2. If the extractor is not in comparable format, we need to read the length of the map.
	if !x.opts.Comparable {
		// 2.1. Read the length of the map.
		ln, n, err := bstio.ReadUint(x.r, x.opts.Descending)
		if err != nil {
			return err
		}
		x.bytesRead += n

		// 2.2. Set the maximum index of the map.
		x.maxIndex = int(ln - 1)
		return nil
	}

	// 3. In the comparable format the length of the map is not known upfront.
	//    Map binary is terminated by a sequence of 0x03 and 0x01 bytes.
	//    At the beginning we need to read the raw bytes, and unescape all consecutive 0x00 0x03 bytes.
	//    Then we need to read elements of the map until we reach io.EOF.
	escape := bstio.MapEscapeAscending
	if x.opts.Descending {
		escape = bstio.MapEscapeDescending
	}

	// 4. Read the raw bytes of the map.
	data, n, err := bstio.ReadComparableBytesReader(x.r, x.opts.Descending, escape)
	if err != nil {
		return err
	}

	x.bytesRead += n

	// 5. Compute the number of elements in the map.
	//    Note: this function could be optimized by reading the values until io.EOF is reached.
	rs := iopool.GetReadSeeker(data)
	sk, sv := bstskip.SkipFuncOf(bt.Key.Type), bstskip.SkipFuncOf(bt.Value.Type)
	kOpts := bstio.ValueOptions{
		Descending:        x.opts.Descending,
		Comparable:        x.opts.Comparable,
		CompatibilityMode: x.opts.CompatibilityMode,
	}
	if bt.Key.Descending {
		kOpts.Descending = true
	}
	vOpts := bstio.ValueOptions{
		Descending:        x.opts.Descending,
		Comparable:        x.opts.Comparable,
		CompatibilityMode: x.opts.CompatibilityMode,
	}
	if bt.Value.Descending {
		vOpts.Descending = true
	}
	for {
		// 5.1. Skip the element key.
		if _, err = sk(rs, kOpts); err != nil {
			// 5.1.1. If the error is EOF, no more elements are in the map.
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		// 5.2. Skip the element value.
		if _, err = sv(rs, vOpts); err != nil {
			// 5.2.1. If the error is EOF, no more elements are in the map.
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		x.maxIndex++
	}

	// 6. Seek back map reader to the beginning.
	_, err = rs.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// 7. Wrap the map reader and set it as a default reader.
	x.r = iopool.WrapReader(rs)

	return nil
}

func (x *Extractor) nextMapElem() bool {
	if x.baseDone {
		return false
	}
	mt := x.embedType.(*bsttype.Map)
	if x.keyDone && !x.elemDone {
		return true
	}

	if !x.elemDone && x.index >= 0 {
		x.err = bsterr.Err(bsterr.CodeNotReadYet, "map extractor key, value not read yet")
		return false
	}

	// 1. Increase the index of the map k-v pairs.
	x.index++

	// 2. Check if all elements of the array were already extracted.
	if x.index > x.maxIndex {
		// 2.1. For comparable binaries, a reader was wrapped, thus we need to unwrap and set it back to extractor.
		if x.opts.Comparable {
			wr := x.r.(*iopool.SharedReadSeeker)
			x.r = wr.Root().(io.ReadSeeker)
			iopool.ReleaseReadSeeker(wr)
		}
		x.baseDone = true
		return false
	}

	switch {
	case x.opts.ExpectedType != nil || x.opts.ExpectedType == x.embedType:
		// 3.1. Set up the next extractor element to be the key.
		x.elemType, x.err = x.derefType(mt.Key.Type)
		if x.err != nil {
			return false
		}
		x.elemDesc = mt.Key.Descending
		x.embed.elemType = x.elemType
	case x.opts.ExpectedType != x.embedType:
		// 3.2. Set up the next extractor element to be the key.
		x.embed.elemType, x.err = x.derefType(mt.Key.Type)
		if x.err != nil {
			return false
		}
		x.elemDesc = mt.Key.Descending

		xmt, ok := x.opts.ExpectedType.(*bsttype.Map)
		if !ok {
			x.err = bsterr.Err(bsterr.CodeInvalidType, "expected type is not a map")
			return false
		}
		x.elemType, x.err = x.derefType(xmt.Key.Type)
		if x.err != nil {
			return false
		}
	}

	// 4. If the descending flag is set, then we need to invert the descending flag.
	if x.opts.Descending {
		x.elemDesc = !x.elemDesc
	}

	// 4. Reset the done flag.
	x.elemDone = false
	return true
}

func (x *Extractor) finishMapElem() {
	// 1. If the extractor is of type Map or Array and the value is not null,
	if !x.isKey {
		x.elemDone = true
		if x.clearElemFn != nil {
			x.clearElemFn()
			x.clearElemFn = nil
		}
		return
	}
	x.keyDone = true
	x.isKey = false

	if x.clearElemFn != nil {
		x.clearElemFn()
		x.clearElemFn = nil
	}

	et := x.embedType.(*bsttype.Map)
	switch {
	case x.opts.ExpectedType != nil || x.opts.ExpectedType == x.embedType:
		x.elemType, x.err = x.derefType(et.Value.Type)
		if x.err != nil {
			return
		}
		x.embed.elemType = x.elemType
	case x.opts.ExpectedType != x.embedType:
		x.embed.elemType, x.err = x.derefType(et.Value.Type)
		if x.err != nil {
			return
		}
		xmt, ok := x.opts.ExpectedType.(*bsttype.Map)
		if !ok {
			x.err = bsterr.Err(bsterr.CodeInvalidType, "expected type is not a map").
				WithDetails(
					bsterr.D("expectedType", x.opts.ExpectedType),
				)
			return
		}
		x.elemType, x.err = x.derefType(xmt.Value.Type)
		if x.err != nil {
			return
		}
	}

	x.elemDesc = et.Value.Descending
	if x.opts.Descending {
		x.elemDesc = !x.elemDesc
	}
}

func (x *Extractor) previewPrevMapType() (bsttype.Type, bool) {
	if x.index == -1 && x.isKey {
		return nil, false
	}
	et := x.embedType.(*bsttype.Map)
	if x.isKey {
		return et.Value.Type, true
	}
	return et.Key.Type, true
}

func (x *Extractor) finishMap() error {
	// 1. Check if the map is already done.
	if x.baseDone {
		return nil
	}
	// 2. Otherwise, the binary data for the remaining map values needs to be skipped.
	//    Check if the key was read but the value wasn't.
	if x.keyDone && !x.elemDone {
		_, err := x.Skip()
		if err != nil {
			x.err = err
			return err
		}
		x.index++
	}

	// 3. Prepare the skipper for the key and value types.
	mt := x.embedType.(*bsttype.Map)
	ks := bstskip.SkipFuncOf(mt.Key.Type)
	vs := bstskip.SkipFuncOf(mt.Value.Type)

	kOpts := bstio.ValueOptions{
		Descending:        x.opts.Descending,
		Comparable:        x.opts.Comparable,
		CompatibilityMode: x.opts.CompatibilityMode,
	}
	if mt.Key.Descending {
		kOpts.Descending = !kOpts.Descending
	}
	vOpts := bstio.ValueOptions{
		Descending:        x.opts.Descending,
		Comparable:        x.opts.Comparable,
		CompatibilityMode: x.opts.CompatibilityMode,
	}
	if mt.Value.Descending {
		vOpts.Descending = !vOpts.Descending
	}

	// 4. Skip all the keys and values.
	for x.index < x.maxIndex {
		//  4.1. Skip the key.
		n, err := ks(x.r, kOpts)
		if err != nil {
			x.err = err
			return err
		}
		x.bytesRead += int(n)

		// 4.2. Skip the value.
		n, err = vs(x.r, vOpts)
		if err != nil {
			x.err = err
			return err
		}
		x.bytesRead += int(n)
		x.index++
	}
	x.baseDone = true
	x.elemDone = true
	return nil
}
