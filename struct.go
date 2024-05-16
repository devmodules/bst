package bst

import (
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bstskip"
	"github.com/devmodules/bst/bsttype"
)

// WriteStruct creates a sub-composer for the struct type elements, and calls the given function.
// This function reuses current composer x so no reallocation
func (x *Composer) WriteStruct(fn func(c *Composer) error) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	if x.needWriteFieldHeader() {
		x.setFieldBuffer()
	}

	// 2. Create a savepoint and resetWithRoot given composer.
	sp := *x

	// 3. Verify if current element matches expected type.
	st, ok := x.elemType.(*bsttype.Struct)
	if !ok {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindStruct),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Reset current composer state.
	x.reset()

	// 5. Initialize the sub-composer.
	if err := x.initializeStructComposer(st, false); err != nil {
		return err
	}

	// 6. Call input function.
	if err := fn(x); err != nil {
		return err
	}

	// 7. Verify if writing was completed
	if x.index <= x.maxIndex {
		return bsterr.Err(bsterr.CodeWritingFailed, "sub-composer didn't write all elements")
	}

	// 8. Store number of bytes written to the struct composer.
	bw := x.bytesWritten

	// 9. Restore the savepoint.
	*x = sp

	// 10. Increase the number of bytes written by the struct composer.
	x.bytesWritten += bw

	// 11. Finish the element.
	if err := x.finishElem(); err != nil {
		return err
	}
	return nil
}

// ReadStruct reads the struct value from the extractor.
func (x *Extractor) ReadStruct(fn func(sx *Extractor) error) error {
	if x.err != nil {
		return x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	// 2. Create a snapshot of current element.
	sp := *x

	// 3. Ensure that the element is a structure.
	if x.elemType.Kind() != bsttype.KindStruct {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindStruct),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 4. Keep embedded and expected struct types.
	xt := x.elemType
	et := x.embed.elemType

	// 5. Reset the extractor.
	x.reset()

	// 6. Set up base types for the struct.
	x.opts.ExpectedType = xt
	x.embedType = et

	// 5. Initialize the base of the structure.
	if err := x.initStructBase(); err != nil {
		return err
	}

	// 6. Execute the extractor.
	if err := fn(x); err != nil {
		return err
	}

	// 7. Finish embedded element.
	if err := x.finishStruct(); err != nil {
		return err
	}

	// 8. Keep the number of bytes read.
	br := x.bytesRead

	// 9. Restore the extractor.
	*x = sp

	// 10. Update the number of bytes read.
	x.bytesRead += br

	// 11. Finish this function element.
	x.finishElem()

	return nil
}

// FieldName returns the name of the current field.
func (x *Extractor) FieldName() string {
	st, ok := x.embedType.(*bsttype.Struct)
	if !ok {
		panic("current extractor is not a struct")
	}
	if x.index >= len(st.Fields) {
		panic("extractor out of bounds")
	}
	return st.Fields[x.index].Name
}

func (x *Extractor) initStructBase() error {
	// 1. Initial index needs to be set to -1 as the Next function advances the index.
	x.index = -1

	// 2. If the compatibility mode is enabled, read the structure header
	if x.opts.CompatibilityMode {
		// 2.1. Read the structure header.
		h, err := x.readCompatibilityStructHeader()
		if err != nil {
			return err
		}

		// 2.2. Set embed type max index.
		x.embed.index = -1
		x.embed.maxIndex = h.maxIndex - 1

		// 2.3. If expected type is defined, then the maximum index is bound by the fields number.
		if x.opts.ExpectedType != nil {
			xt, ok := x.opts.ExpectedType.(*bsttype.Struct)
			if !ok {
				return bsterr.Errf(bsterr.CodeInvalidType, "expected type is not a struct: %v", x.opts.ExpectedType)
			}

			// 2.3.1. Set the maximum index to the number of fields decreased by 1 - we're starting the counter from 0.
			x.maxIndex = len(xt.Fields) - 1
		} else {
			// 2.3.2. The maximum index is equal to embed maximum index.
			x.maxIndex = x.embed.maxIndex
		}
		return nil
	}

	et, ok := x.embedType.(*bsttype.Struct)
	if !ok {
		return bsterr.Errf(bsterr.CodeInvalidType, "expected struct type, got %v", x.embedType)
	}

	// 3. If no expected type was provided, we base everything on the embedded one.
	if x.opts.ExpectedType == nil {
		// 3.1. Set the max index to the number of embedded fields.
		x.maxIndex = len(et.Fields) - 1
		x.embed.maxIndex = x.maxIndex
		return nil
	}

	xt, ok := x.opts.ExpectedType.(*bsttype.Struct)
	if !ok {
		return bsterr.Errf(bsterr.CodeInvalidType, "expected type is not a struct: %v", x.opts.ExpectedType)
	}

	// 4. If the expected type is the same as the embedded one, we base everything on the embedded one.
	if xt.CompareType(et) {
		// 4.1. Set the max index to the number of embedded fields.
		x.maxIndex = len(et.Fields) - 1
		return nil
	}

	// 5. If the expected type is a subset of the embedded one, we base everything on the expected one.
	x.maxIndex = len(xt.Fields) - 1
	x.embed.maxIndex = len(et.Fields) - 1

	return nil
}

// This function advances an index of the struct field and checks whether a new expected field is possible to extract.
func (x *Extractor) nextStructElem() bool {
	// 1. Check if the extractor is in compatibility mode.
	if x.opts.CompatibilityMode {
		hasNext, err := x.nextCompatibilityStructElem()
		if err != nil {
			x.err = err
			return false
		}
		return hasNext
	}

	// 2. Check if extractor embed type matches expected type or there is no expected type.
	if x.embedType == x.opts.ExpectedType || x.opts.ExpectedType == nil {
		return x.nextEmbedStructElem()
	}

	// 3. Otherwise, the next element needs to be taken from the expected type.
	hasNext, err := x.nextExpectedStructElem()
	if err != nil {
		x.err = err
		return false
	}
	return hasNext
}

func (x *Extractor) nextCompatibilityStructElem() (bool, error) {
	// 1. If there were no expected type defined, we need to base on the embedded type.
	if x.opts.ExpectedType == nil {
		return x.nextStructElemCompatibilityNoExpected()
	}

	// 2. If the expected type is the same as the embed type, it means that no embed type was defined in the binary
	//
	if x.embedType == x.opts.ExpectedType {
		return x.nextStructElemCompatibilityNoEmbed()
	}

	// 3. A third scenario is when the expected type is defined, as well as the embedded one, but they are not the same.
	return x.nextStructElemCompatibilityEmbedNotExpected()
}

func (x *Extractor) nextStructElemCompatibilityNoExpected() (bool, error) {
	// 1. If expected type is not defined, it ensures that the type was defined in the binary, and we can
	// 		safely try to read the next field.
	et := x.embedType.(*bsttype.Struct)

	x.index++
	x.elemDone = false
	if x.index > x.maxIndex {
		// 2. If the index is greater than the maximum index, we are done.
		x.baseDone = true
		return false, nil
	}

	// 3. Read the file header.
	fh, err := x.readCompatibleField()
	if err != nil {
		return false, err
	}

	// 4. Ensure that the field identifier is the expected one.
	//      This is kind of prevention for malformed binaries.
	if fh.index != x.index {
		return false, bsterr.Err(bsterr.CodeMalformedBinary, "expected embed field index doesn't match the one in the field header")
	}

	x.elemType = et.Fields[x.index].Type
	x.elemType, x.err = x.derefType(x.elemType)
	if x.err != nil {
		return false, x.err
	}

	x.embed.elemType = x.elemType
	x.elemDesc = et.Fields[x.index].Descending
	if x.opts.Descending {
		x.elemDesc = !x.elemDesc
	}
	return true, nil
}

func (x *Extractor) nextStructElemCompatibilityNoEmbed() (bool, error) {
	// If the expected type is defined and there were no embedded type in the binary,
	// We're using only expected type fields.
	// the compatibility mode, ensures that the number of fields are encoded in the binary.
	// Thus, we need to check if the next field is the expected one.
	xt, ok := x.opts.ExpectedType.(*bsttype.Struct)
	if !ok {
		panic("expected type is not a struct")
	}

	// 1. Advance the index of the expected type, to the next one.
	x.index++
	x.elemDone = false

	// 2. Check if we're expecting more fields.
	if x.index > x.maxIndex {
		// 2.1. If the is no more fields to expect, we are done.
		//
		// 2.2. Check if the embedded field had been used already.
		if x.embed.used {
			x.embed.index++
		}
		// 2.3. Embedded type index could still have more fields, thus we need to skip till the end of the struct.
		for x.embed.index <= x.embed.maxIndex {
			// 2.2. Read the file header.
			fh, err := x.readCompatibleField()
			if err != nil {
				return false, err
			}

			// 2.3. Skip all the bytes written for given field.
			_, err = x.r.Seek(int64(fh.length), io.SeekCurrent)
			if err != nil {
				return false, err
			}
			x.bytesRead += fh.length
			x.embed.index++
		}
		// 2.3. Now, we have read all the fields in the binary, and the expected type does not have more fields.
		//        We are done.
		x.baseDone = true
		x.embed.used = true
		return false, nil
	}
	exField := xt.Fields[x.index]

	// 3. If the field Index already matches the one read in previous calls, set up current element to the expected one.
	if exField.Index == uint(x.fieldHeader.index) {
		// 1.1. Set current struct element to the expected one.
		x.elemType, x.err = x.derefType(exField.Type)
		if x.err != nil {
			return false, x.err
		}
		x.elemDesc = exField.Descending
		if x.opts.Descending {
			x.elemDesc = !x.elemDesc
		}
		x.embed.elemType = x.elemType
		x.embed.used = true
		return true, nil
	}

	// 4. Check if current expected field is smaller than the one read from the file header.
	if exField.Index < uint(x.fieldHeader.index) {
		// 4.1. In this case, we cannot read the next field.
		return false, nil
	}

	// 5. We're in a case where the expected field is greater than the one read from the file header.
	//    exField.Index > fieldHeader.index
	//
	// 5.1. Skip the bytes from the previously read field header.
	if !x.embed.used {
		_, err := x.r.Seek(int64(x.fieldHeader.length), io.SeekCurrent)
		if err != nil {
			return false, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to seek to the next field")
		}
		x.bytesRead += x.fieldHeader.length
		x.embed.used = true
	}

	var err error
	// 5.2. The iteration of the embed type need to be over once the maximum is reached - a maximum is defined in the
	//      binary header.
	for x.embed.index <= x.embed.maxIndex {
		// 5.3. Read the field header.
		x.fieldHeader, err = x.readCompatibleField()
		if err != nil {
			return false, err
		}
		x.embed.index++
		x.embed.used = false

		// 5.4. Compare indexes of the field header with the one expected.
		if x.fieldHeader.index == int(exField.Index) {
			x.elemType, x.err = x.derefType(exField.Type)
			if x.err != nil {
				return false, x.err
			}
			x.elemDesc = exField.Descending
			if x.opts.Descending {
				x.elemDesc = !x.elemDesc
			}
			x.embed.elemType = x.elemType
			x.embed.used = true
			return true, nil
		}

		// 5.5. Check if the index in the field header is already greater than the one that we expect.
		//      This means that the binary type does not contain the expected field.
		if x.fieldHeader.index > int(exField.Index) {
			return false, nil
		}

		// 5.6. If the field header index is still smaller than the one we expect, we need to skip the bytes.
		_, err = x.r.Seek(int64(x.fieldHeader.length), io.SeekCurrent)
		if err != nil {
			return false, err
		}
		x.bytesRead += x.fieldHeader.length
	}

	// 6. This scenario occurs if there are no more fields in the embedded binary type to read.
	//    Expected field was not found, and no more fields from expected type are available.
	x.elemDone = true
	return false, nil
}

func (x *Extractor) nextStructElemCompatibilityEmbedNotExpected() (bool, error) {
	// In this scenario an embedded type was defined in the binary, as well as the expected one, but they are not the same.
	et := x.embedType.(*bsttype.Struct)
	xt := x.opts.ExpectedType.(*bsttype.Struct)

	// 1. Advance the index of the expected type, to the next one.
	x.index++
	x.elemDone = false

	// 2. Check if no more expected fields are available.
	var err error
	if x.index > x.maxIndex {
		if x.embed.used {
			x.embed.index++
			x.embed.used = false
		}
		// 2.1. If the is no more fields to expect, we are done.
		//        However, embedded type index could still have more fields, thus we may need to skip till the end of the struct.
		for x.embed.index <= x.embed.maxIndex {
			// 2.1.2. If the embed field was not used we need to skip it.
			if !x.embed.used {
				_, err = x.r.Seek(int64(x.fieldHeader.length), io.SeekCurrent)
				if err != nil {
					return false, err
				}
				x.bytesRead += x.fieldHeader.length
				x.embed.used = true
			}

			// 2.1.3. Read the next field header.
			x.fieldHeader, err = x.readCompatibleField()
			if err != nil {
				return false, err
			}
			x.embed.used = false
			x.embed.index++
		}
		// 2.2. Now, we have read all the fields in the binary, and the expected type does not have more fields.
		//        We are done.
		x.baseDone = true
		return false, nil
	}

	// 3. If the embed field was already used, we need to increase the index of the embed field.
	if x.embed.used || x.embed.index == -1 {
		x.embed.index++

		// 3.1. Check if the embed field index reached the maximum.
		if x.embed.index > x.embed.maxIndex {
			// 3.1.1. We're not setting the extractor as done, because we still have more fields to read in expected type.
			return false, nil
		}

		// 3.2. Read the next field header.
		x.fieldHeader, err = x.readCompatibleField()
		if err != nil {
			return false, err
		}
		x.embed.used = false
	}

	xField := xt.Fields[x.embed.index]
	etField := et.Fields[x.index]

	// 4. If the expected field is before the embedded field, then this field cannot be reached.
	//    This occurs when the embedded type definition is less specific than the expected type.
	//    In case when there are index gaps in the embedded type fields, which occurs in expected type
	//    a field could be defined in expected type, but not in the embedded type.
	//    In this case, we're not reading the field, but quickly returning false.
	if xField.Index < etField.Index {
		return false, nil
	}

	// 5. If the indexes are the same, then we can set up the next extractor element to be the expected one.
	if xField.Index == etField.Index {
		// 5.1. Set current struct element to the expected one.
		x.elemType, x.err = x.derefType(xField.Type)
		if x.err != nil {
			return false, x.err
		}
		// 5.2. Set the embed element type.
		x.embed.elemType, x.err = x.derefType(etField.Type)
		if x.err != nil {
			return false, x.err
		}
		x.embed.used = true

		// 5.3. If the expected field is descending, we need to invert the element descending flag.
		x.elemDesc = xField.Descending
		if x.opts.Descending {
			x.elemDesc = !x.elemDesc
		}
		return true, nil
	}

	// 6. The expected index is after the embedded index, so we need to skip the bytes of the embedded field.
	//    expectedField.Index > embeddedField.Index
	//
	//    Check if expected field in the embedded binary.
	for x.embed.index <= x.embed.maxIndex {
		if !x.embed.used {
			_, err = x.r.Seek(int64(x.fieldHeader.length), io.SeekCurrent)
			if err != nil {
				return false, err
			}

			x.bytesRead += x.fieldHeader.length
			x.embed.used = true
		}

		// 6.1. Read the field header.
		x.fieldHeader, err = x.readCompatibleField()
		if err != nil {
			return false, err
		}

		// 6.2. Advance the index of embedded type.
		x.embed.index++
		x.embed.used = false

		if xField.Index == uint(x.fieldHeader.index) {
			// 6.3. The expected field matches the embedded field, so we can set up the next extractor element to be the expected one.
			x.elemType, x.err = x.derefType(xField.Type)
			if x.err != nil {
				return false, x.err
			}

			// 6.4. Set up embed elem.
			x.embed.elemType, x.err = x.derefType(et.Fields[x.embed.index].Type)
			if x.err != nil {
				return false, x.err
			}
			x.embed.used = true

			// 6.5. Determine if the field value is descending.
			x.elemDesc = xField.Descending
			if x.opts.Descending {
				x.elemDesc = !x.elemDesc
			}
			return true, nil
		}

		// 6.6. If the index of the expected field is smaller than the index of the embedded field, then the
		//        expected field is not in the binary, and we return false.
		if xField.Index < uint(x.fieldHeader.index) {
			return false, nil
		}

		// 6.7. If the index of the embedded field is still less than the expected index, then we can try reaching
		//      the next embed field.
		//      xField.Index > x.fieldHeader.index
	}
	// 6.8. We have read all the fields in the embedded binary, and the expected field is not in the binary.
	//        We are done, however, we still have more fields to read in the expected type,
	//        and we're not marking the extractor as done.
	return false, nil
}

func (x *Extractor) nextEmbedStructElem() bool {
	// In this case, either there is no expected type or the expected type is the embedded one.
	// In both cases, we need to advance the index of the embedded type and map the field as next.
	x.index++
	x.elemDone = false

	// 1. If the index is greater than the max index, then we are done.
	//    No more fields are available in the embedded type.
	if x.index > x.maxIndex {
		x.baseDone = true
		return false
	}

	// 2. Get the field from the embedded type.
	et := x.embedType.(*bsttype.Struct)
	eField := et.Fields[x.index]

	// 3. Set the next extractor element to be the field from the embedded type.
	x.elemType, x.err = x.derefType(eField.Type)
	if x.err != nil {
		return false
	}
	x.embed.elemType, x.err = x.derefType(eField.Type)
	if x.err != nil {
		return false
	}
	x.embed.used = true
	x.elemDesc = eField.Descending

	// 4. If the descending flag is set, then we need to invert the descending flag.
	if x.opts.Descending {
		x.elemDesc = !x.elemDesc
	}

	return true
}

func (x *Extractor) nextExpectedStructElem() (bool, error) {
	// In this scenario, the expected type is provided, and it is not the embedded one.

	// 1. Advance the index of the expected type, to the next one.
	x.index++
	x.elemDone = false

	et := x.embedType.(*bsttype.Struct)
	xt := x.opts.ExpectedType.(*bsttype.Struct)
	// 2. Check if there are no more fields in the expected type.
	if x.index > x.maxIndex {
		if x.embed.used {
			x.embed.index++
			x.embed.used = false
		}
		// 2.1. It is possible that the embedded type has more fields, thus we may need to skip till the end of the struct.
		for x.embed.index <= x.embed.maxIndex {
			// 2.1.1. Get the field type from the embedded type.
			eField := et.Fields[x.embed.index]

			opts := bstio.ValueOptions{
				Descending:        x.opts.Descending,
				Comparable:        x.opts.Comparable,
				CompatibilityMode: x.opts.CompatibilityMode,
			}
			if eField.Descending {
				opts.Descending = !opts.Descending
			}

			// 2.1.2. Create a skipper for the field type, and skip the bytes.
			n, err := bstskip.SkipFuncOf(eField.Type)(x.r, opts)
			if err != nil {
				return false, err
			}
			x.bytesRead += int(n)

			// 2.1.3. Advance the index of the embedded type.
			x.embed.index++
		}
		// 2.2. We are done - no more fields in the expected type and no more fields in the embedded type.
		x.baseDone = true
		x.embed.used = true
		return false, nil
	}

	// 3. Check if the index of the expected field matches current embedded field.
	//    This could occur if a previous expected field was not found in the embedded type.
	xField := xt.Fields[x.index]

	if x.embed.used {
		x.embed.index++
		x.embed.used = false
	}

	// 4. Try to match the expected field with the embedded field.
	for x.embed.index <= x.embed.maxIndex {
		// 4.1. Get the field from the embedded type.
		eField := et.Fields[x.embed.index]

		// 4.2. Check if the indexes of the expected field and the embedded field match.
		if xField.Index == eField.Index {
			// 4.2.1. Now, the expected field is in the embedded type, so we can set up the next extractor element to be the expected one.
			x.elemType = xField.Type
			x.embed.elemType = eField.Type
			x.embed.used = true

			x.elemDesc = xField.Descending
			if x.opts.Descending {
				x.elemDesc = !x.elemDesc
			}
			return true, nil
		}

		// 4. If the expected field index is still before the embedded one,
		//    then current expected field is not in the embedded type.
		if xField.Index < eField.Index {
			return false, nil
		}

		opts := bstio.ValueOptions{
			Descending:        x.opts.Descending,
			Comparable:        x.opts.Comparable,
			CompatibilityMode: x.opts.CompatibilityMode,
		}
		if eField.Descending {
			opts.Descending = !opts.Descending
		}

		// 4.3. The expected field is after the embedded field, so we need to skip the bytes of the embedded field.
		//      expectedField.Index > embeddedField.Index
		n, err := bstskip.SkipFuncOf(eField.Type)(x.r, opts)
		if err != nil {
			return false, err
		}
		x.bytesRead += int(n)

		// 4.4.	Advance the index of the embedded type, and again try to match the expected field with the embedded field.
		x.embed.index++
		x.embed.used = true
	}

	// 5. We are done - no more fields in the embedded type.
	//    However, we're not marking the extractor as done, because we still have fields in the expected type.
	//    Once all fields got processed, we will mark the extractor as done (in the point 2. above).
	return false, nil
}

func (x *Extractor) finishStruct() error {
	if x.elemDone || x.baseDone {
		return nil
	}

	// 1. If the compatibility mode is on
	if x.opts.CompatibilityMode {
		if x.opts.ExpectedType == nil {
			// 1.1 When the expected type is not set, we use embedded indexes as default.
			x.embed.index = x.index
			x.embed.maxIndex = x.maxIndex
		}

		for x.embed.index <= x.embed.maxIndex {
			if !x.embed.used {
				_, err := x.r.Seek(int64(x.fieldHeader.length), io.SeekCurrent)
				if err != nil {
					return err
				}
				x.bytesRead += x.fieldHeader.length
				x.embed.used = true
			}
			fh, err := x.readCompatibleField()
			if err != nil {
				return err
			}

			x.bytesRead += fh.length
			x.embed.used = false
			x.embed.index++
		}
		x.elemDone = true
		return nil
	}

	// 2. Check if the embed type matches expected or there is no expected type.
	if x.embedType == x.opts.ExpectedType || x.opts.ExpectedType == nil {
		et, ok := x.embedType.(*bsttype.Struct)
		if !ok {
			return bsterr.Err(bsterr.CodeInvalidType, "embedded type is not a struct")
		}

		x.index++

		// 2.1. Iterate over the fields and skip one by one.
		for x.index <= x.maxIndex {
			eField := et.Fields[x.index]

			opts := bstio.ValueOptions{
				Descending:        x.opts.Descending,
				Comparable:        x.opts.Comparable,
				CompatibilityMode: x.opts.CompatibilityMode,
			}
			if eField.Descending {
				opts.Descending = !opts.Descending
			}
			n, err := bstskip.SkipFuncOf(eField.Type)(x.r, opts)
			if err != nil {
				return err
			}
			x.bytesRead += int(n)
			x.index++
		}
		x.elemDone = true
		return nil
	}

	// 3. In this case an embedded type is different from the expected one.
	et := x.embedType.(*bsttype.Struct)
	for x.embed.index <= x.embed.maxIndex {
		if !x.embed.used {
			etField := et.Fields[x.embed.index]

			opts := bstio.ValueOptions{
				Descending:        x.opts.Descending,
				Comparable:        x.opts.Comparable,
				CompatibilityMode: x.opts.CompatibilityMode,
			}
			if etField.Descending {
				opts.Descending = !opts.Descending
			}
			n, err := bstskip.SkipFuncOf(etField.Type)(x.r, opts)
			if err != nil {
				return err
			}

			x.bytesRead += int(n)
			x.embed.used = true
		}

		x.embed.index++
	}
	x.elemDone = true

	return nil
}

func (x *Extractor) finishStructElem() {
	x.elemDone = true
	if x.clearElemFn != nil {
		x.clearElemFn()
		x.clearElemFn = nil
	}
}

func (x *Extractor) previewPrevStructElem() (bsttype.Type, bool) {
	if x.index == 0 {
		return nil, false
	}
	et := x.embedType.(*bsttype.Struct)
	return et.Fields[x.index-1].Type, true
}

type fieldHeader struct {
	index  int
	length int
}

func (x *Extractor) readCompatibleField() (fieldHeader, error) {
	idx, n, err := bstio.ReadUint(x.r, false)
	if err != nil {
		return fieldHeader{}, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to read field index")
	}

	x.bytesRead += n

	length, n, err := bstio.ReadUint(x.r, false)
	if err != nil {
		return fieldHeader{}, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to read field length")
	}

	x.bytesRead += n

	return fieldHeader{int(idx), int(length)}, nil
}

type compatibilityStructHeader struct {
	maxIndex int
}

func (x *Extractor) readCompatibilityStructHeader() (compatibilityStructHeader, error) {
	maxIndex, n, err := bstio.ReadUint(x.r, false)
	if err != nil {
		return compatibilityStructHeader{}, err
	}
	x.bytesRead += n

	return compatibilityStructHeader{maxIndex: int(maxIndex)}, nil
}
