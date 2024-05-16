package bst

import (
	"fmt"
	"io"
	"math"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
	"github.com/devmodules/bst/internal/iopool"
)

// ComposerOptions is the options for the composer.
type ComposerOptions struct {
	Descending        bool
	Comparable        bool
	CompatibilityMode bool
	EmbedType         bool
	Modules           *bsttype.Modules
	Length            int
}

// Composer is the composer for the binary serialization of the BST.
type Composer struct {
	baseType        bsttype.Type
	index, maxIndex int
	elemType        bsttype.Type
	opts            ComposerOptions
	w               io.Writer
	boolBuf         byte
	boolBufPos      byte
	elemDesc, isKey, lengthWritten, definedLength,
	done, fhWritten, bufWrites bool
	bytesWritten    int
	modules         *bsttype.Modules
	externalModules bool
}

// NewComposer creates a new binary value composer.
func NewComposer(w io.Writer, baseType bsttype.Type, opts ComposerOptions) (*Composer, error) {
	// 1. Create the composer.
	c := &Composer{w: w}

	// 2. Apply the options.
	if err := c.applyOptions(opts); err != nil {
		return nil, err
	}

	// 2. Initialize it.
	if err := c.initializeComposer(baseType, true); err != nil {
		return nil, err
	}

	// 3. Return the composer.
	return c, nil
}

// Close the composer, finishing any pending writes.
func (x *Composer) Close() error {
	if !x.externalModules && x.modules != nil {
		defer x.modules.Free()
	}

	switch bt := x.baseType.(type) {
	case *bsttype.Struct:
		return x.closeStruct(bt)
	case *bsttype.Array:
		return x.closeArray(bt)
	case *bsttype.Map:
		return x.closeMap()
	default:
		return nil
	}
}

// IsDone returns true if the composer has finished writing the current element.
func (x *Composer) IsDone() bool {
	return x.done
}

// ResetOn resetWithRoot the state, writer and the base type of the composer, so it could be used again
// without unnecessary allocations.
func (x *Composer) ResetOn(w io.Writer, baseType bsttype.Type, opts ComposerOptions) error {
	// 1. Reset the composer to the initial state.
	*x = Composer{w: w}

	if err := x.applyOptions(opts); err != nil {
		return err
	}

	// 2. Initialize the composer.
	if err := x.initializeComposer(baseType, true); err != nil {
		return err
	}
	return nil
}

// Reset the state of the composer to its initial state.
func (x *Composer) Reset(opts ComposerOptions) error {
	x.boolBuf = 0x00
	x.boolBufPos = 0
	x.bytesWritten = 0
	x.isKey = false
	x.index = 0
	x.maxIndex = 0
	x.done = false
	x.bufWrites = false
	x.definedLength = false

	if err := x.applyOptions(opts); err != nil {
		return err
	}

	switch t := x.baseType.(type) {
	case *bsttype.Struct:
		return x.initializeStructComposer(t, true)
	case *bsttype.Array:
		return x.initializeArrayComposer(t, true)
	case *bsttype.Map:
		return x.initializeMapComposer(t, true)
	case *bsttype.Named:
		return x.initializeNamedComposer(t, true)
	default:
		return x.initializeBasicComposer(t, true)
	}
}

// SkipField skips writing selected field in the structure.
func (x *Composer) SkipField() error {
	// 1. Verify if the composer is based on the structure.
	st, ok := x.baseType.(*bsttype.Struct)
	if !ok {
		return bsterr.Err(bsterr.CodeInvalidType, "cannot skip field in non-struct type")
	}

	// 2. Finish current field of a structure.
	if err := x.finishStructElem(st); err != nil {
		return err
	}
	return nil
}

func (x *Composer) initializeComposer(baseType bsttype.Type, header bool) (err error) {
	// 1. Switch the composer based on the base type.
	switch bt := baseType.(type) {
	case *bsttype.Struct:
		return x.initializeStructComposer(bt, header)
	case *bsttype.Array:
		return x.initializeArrayComposer(bt, header)
	case *bsttype.Map:
		return x.initializeMapComposer(bt, header)
	case *bsttype.Named:
		return x.initializeNamedComposer(bt, header)
	default:
		return x.initializeBasicComposer(bt, header)
	}
}

func (x *Composer) initializeNamedComposer(bt *bsttype.Named, header bool) error {
	// 1. Set up the base type.
	x.baseType = bt

	// 2. If the header option is true, resolve all dependencies if needed and write the header.
	if header {
		// 3. Prepare type dependencies
		if err := x.prepareTypeDependencies(); err != nil {
			return err
		}

		// 4. Write the header.
		if err := x.writeHeader(); err != nil {
			return err
		}
	}

	// 5. Recursively initialize the composer.
	return x.initializeComposer(bt.Type, false)
}

func (x *Composer) initializeArrayComposer(bt *bsttype.Array, header bool) error {
	// 1. Set up composer base type.
	x.baseType = bt

	// 2. Check if the header needs to be written.
	if header {
		// 2.1. Prepare type dependencies.
		if err := x.prepareTypeDependencies(); err != nil {
			return err
		}

		// 2.2. Write the header.
		if err := x.writeHeader(); err != nil {
			return err
		}
	}

	// 3. Initialize the composer for an array.
	x.initializeArray(bt)

	// 4. verify if the composer is valid.
	if err := x.verifyArrayBase(); err != nil {
		return err
	}

	// 5. If the array length is defined and the array is not of fixed size
	// 	then write the length.
	if !bt.HasFixedSize() && x.definedLength && !x.opts.Comparable {
		if err := x.writeArrayLength(bt); err != nil {
			return err
		}
	}
	return nil
}

func (x *Composer) initializeMapComposer(bt *bsttype.Map, header bool) error {
	// 1. Set up composer base type.
	x.baseType = bt

	// 2. Check if the header needs to be written.
	if header {
		// 2.1. Prepare type dependencies.
		if err := x.prepareTypeDependencies(); err != nil {
			return err
		}

		// 2.2. Write the header.
		if err := x.writeHeader(); err != nil {
			return err
		}
	}

	// 3. Initialize the composer for a map.
	x.initializeMap(bt)

	// 4. Verify if the composer is valid.
	if err := x.verifyMapBase(); err != nil {
		return err
	}

	// 5. If the length was predefined, write it to the writer.
	if x.definedLength {
		if err := x.writeMapLength(); err != nil {
			return err
		}
	}
	return nil
}

func (x *Composer) initializeStructComposer(st *bsttype.Struct, header bool) error {
	// 1. Set the base type for the structure.
	x.baseType = st

	// 2. Check if the header needs to be written.
	if header {
		// 3.1. Prepare type dependencies.
		if err := x.prepareTypeDependencies(); err != nil {
			return err
		}

		// 3.2. Write the header.
		if err := x.writeHeader(); err != nil {
			return err
		}
	}

	// 4. Set up maximum index as the number of fields in the structure.
	x.maxIndex = len(st.Fields) - 1
	if x.maxIndex >= 0 {
		// 4.1. If the structure has fields, set the first element to the 0th field index.
		x.elemType = st.Fields[0].Type
		x.elemDesc = st.Fields[0].Descending
	}

	// 5. Estimate initial field order.
	if x.opts.Descending {
		x.elemDesc = !x.elemDesc
	}

	// 6. verify if the composer is valid.
	if err := x.verifyStructBase(); err != nil {
		return err
	}

	// 7. If the type is in compatibility mode write the struct header to the writer.
	if x.opts.CompatibilityMode {
		if err := x.writeStructHeader(); err != nil {
			return err
		}
	}
	return nil
}

func (x *Composer) prepareTypeDependencies() error {
	// 1. If the type needs to be embedded, and no modules are provided,
	//    check if we need to compose the modules for the type.
	if x.opts.EmbedType && !x.externalModules && x.modules == nil {
		// 1.1. If the type implements DependencyOperator determine if we need to prepare modules.
		do, ok := x.baseType.(bsttype.DependencyOperator)
		if ok && do.NeedsDependencies() {
			// 1.2. Verify named dependencies.
			if err := do.VerifyDependencies(); err != nil {
				return err
			}

			// 1.3. Modules are not defined, thus we need to get a shared instance of the modules.
			//    NOTE: Close function needs to free the modules.
			x.modules = bsttype.GetSharedModules()

			// 1.4. Compose the modules.
			if err := do.ComposeDependencies(x.modules); err != nil {
				return err
			}
		}
		// 1.5. Embedded type are now prepared.
		return nil
	}

	// 2. If the external modules where provided, verify if they're resolved and verify its dependencies.
	if x.externalModules {
		// 3. Check if the modules are resolved already, if not, resolve them.
		if !x.modules.IsResolved() {
			if err := x.modules.Resolve(); err != nil {
				return err
			}
		}
	}

	// 4. If the input type needs dependencies, verify if all dependencies named types contains the referenced Type.
	dnv, ok := x.baseType.(bsttype.DependencyNeeder)
	if ok && !dnv.NeedsDependencies() {
		return nil
	}

	// 5. Check if all the necessary module dependencies are valid.
	dc, ok := x.baseType.(bsttype.DependencyChecker)
	if !ok {
		return nil
	}

	// 5.1. The dependencies might need to be checked if they need to be resolved.
	res, err := dc.CheckDependencies(x.modules)
	if err != nil {
		return err
	}
	if !res.ResolveRequired && !res.ComposeRequired {
		// 5.2. Otherwise we're done here.
		return nil
	}

	// 6. If the dependencies needs to be resolved we need to have modules.
	if res.ComposeRequired && x.modules == nil {
		// NOTE: Close function needs to free the modules.
		x.modules = bsttype.GetSharedModules()

		// 7. In order to resolve the dependencies, we need to compose the modules.
		//    NOTE: If the base type implements dependency checker it also implements DependencyComposer.
		if err = x.baseType.(bsttype.DependencyComposer).ComposeDependencies(x.modules); err != nil {
			return err
		}

		// 8. Resolve the modules so that all the types are resolved.
		if err = x.modules.Resolve(); err != nil {
			return err
		}
	}

	if res.ResolveRequired {
		// 9. Resolve the dependencies.
		if !x.modules.IsResolved() {
			if err = x.modules.Resolve(); err != nil {
				return err
			}
		}

		// 10. Resolve the dependencies.
		_, err = x.baseType.(bsttype.DependencyResolver).ResolveDependencies(x.modules)
		if err != nil {
			return err
		}
	}

	return nil
}

func (x *Composer) finishElem() error {
	switch et := x.baseType.(type) {
	case *bsttype.Struct:
		return x.finishStructElem(et)
	case *bsttype.Array:
		x.finishArrayElem(et)
	case *bsttype.Map:
		x.finishMapElem(et)
	}
	return nil
}

func (x *Composer) finishStructElem(et *bsttype.Struct) error {
	// 1. Check if the element was written in the buffer.
	if fb, ok := x.w.(*iopool.SharedBuffer); ok && x.opts.CompatibilityMode && x.bufWrites {
		// 1.1. Retrieve the root writer.
		root := fb.Root

		// 1.2. Write the field header.
		n, err := x.writeFieldHeader(root, x.fieldIndex(), uint(fb.Len()))
		if err != nil {
			return err
		}

		x.bytesWritten += n

		// 1.3. Write buffered field value.
		_, err = fb.WriteTo(root)
		if err != nil {
			return err
		}

		// 1.4 Reset root writer.
		x.w = root
		x.bufWrites = false

		// 1.5. Release the field buffer.
		iopool.ReleaseBuffer(fb)
	}

	return x.incrementStructElem(et)
}

func (x *Composer) incrementStructElem(et *bsttype.Struct) error {
	// 2. Increment current struct field buffIndex.
	x.index++
	x.fhWritten = false

	// 3. Check if the struct already reached its max index.
	if x.index > x.maxIndex {
		// 2. If the buffIndex reached the end of the struct, mark the composer as done.
		x.done = true
		return nil
	}
	// 4. If the buffIndex is not the end of the struct, set the current element to the next field.
	x.elemType = et.Fields[x.index].Type

	// 5. If the field is a named type dereference it.
	for {
		if nt, ok := x.elemType.(*bsttype.Named); ok {
			x.elemType = nt.Type
		} else {
			break
		}
	}
	// 6. Set up the encoding order for the next field.
	x.elemDesc = et.Fields[x.index].Descending
	if x.opts.Descending {
		x.elemDesc = !x.elemDesc
	}
	return nil
}

func (x *Composer) fieldIndex() uint {
	// 1. Only the struct base composer has a field index.
	st, ok := x.baseType.(*bsttype.Struct)
	if !ok {
		panic(fmt.Sprintf("invalid type to get a field index: %T", x.baseType))
	}
	return st.Fields[x.index].Index
}

func (x *Composer) finishArrayElem(et *bsttype.Array) {
	// 1. Increment current array buffIndex.
	x.index++

	// 2. If the buffIndex reached the end of the array, mark the composer as done.
	if x.index > x.maxIndex {
		x.done = true
		return
	}

	// 3. Reset the current element to the next element.
	//    This may be required if the element is some composite type - i.e. Nullable.
	x.elemType = et.Type

	// 4. If the field is a named type dereference it.
	for {
		if nt, ok := x.elemType.(*bsttype.Named); ok {
			x.elemType = nt.Type
		} else {
			break
		}
	}

	// 5. Set up the encoding order for the next element.
	x.elemDesc = x.opts.Descending
}

func (x *Composer) finishMapElem(et *bsttype.Map) {
	// 1. Check if current element written was a key or value.
	if x.isKey {
		// 2. If the current element was a key, set the current element to the value.
		x.isKey = false
		x.elemType = et.Value.Type

		// 2.1. Dereference possible named type.
		for {
			if nt, ok := x.elemType.(*bsttype.Named); ok {
				x.elemType = nt.Type
			} else {
				break
			}
		}

		x.elemDesc = et.Value.Descending
		if x.opts.Descending {
			x.elemDesc = !x.elemDesc
		}
		return
	}
	// 3. Otherwise, mark current element as done and increment current buffIndex.
	//    And set current element type to the key.
	x.index++

	// 4. If the index reached maximum, mark the composer as done.
	if x.index > x.maxIndex {
		x.done = true
		return
	}

	// 5. Reset the pointer to the key, with its type and descending flag.
	x.isKey = true
	x.elemType = et.Key.Type

	// 6. Dereference possible named type.
	for {
		if nt, ok := x.elemType.(*bsttype.Named); ok {
			x.elemType = nt.Type
		} else {
			break
		}
	}

	x.elemDesc = et.Key.Descending
	if x.opts.Descending {
		x.elemDesc = !x.elemDesc
	}
}

func (x *Composer) previewNextElem() (bsttype.Type, bool) {
	switch et := x.baseType.(type) {
	case *bsttype.Array:
		if x.index+1 > x.maxIndex {
			return nil, false
		}
		return et.Type, true
	case *bsttype.Struct:
		if x.index+1 > x.maxIndex {
			return nil, false
		}
		return et.Fields[x.index+1].Type, true
	case *bsttype.Map:
		if x.isKey {
			return et.Value.Type, true
		}
		if x.index+1 > x.maxIndex {
			return nil, false
		}
		return et.Key.Type, true
	default:
		return nil, false
	}
}

func (x *Composer) initializeStruct(st *bsttype.Struct) {
	// 1. Set up the base type.
	x.baseType = st

	// 2. Maximum index is the maximum number of fields.
	//    This field is only for the StructType fields slice index,
	//	  and not for the maximum value of the struct fields Index field.
	x.maxIndex = len(st.Fields) - 1
	if x.maxIndex >= 0 {
		// 3. If the struct has fields, set the current element to the first field.
		x.elemType = st.Fields[0].Type
		x.elemDesc = st.Fields[0].Descending
	}

	// 3. Estimate initial field order.
	if x.opts.Descending {
		x.elemDesc = !x.elemDesc
	}
}

func (x *Composer) initializeArray(at *bsttype.Array) {
	// 1. Set up the base type to the array.
	x.baseType = at

	// 2. Set up an element type of the array.
	x.elemType = at.Elem()
	x.elemDesc = x.opts.Descending

	// 3. If the array has fixed size, set the maximum index to the array size.
	if at.HasFixedSize() {
		x.maxIndex = int(at.FixedSize) - 1
		return
	}

	// 4. If the length of the array was not specified, set the maximum index to MaxInt.
	//    The composer needs to be closed for undefined length arrays.
	if at.FixedSize == 0 && (!x.definedLength || x.opts.Comparable) {
		x.maxIndex = math.MaxInt
		x.w = iopool.GetBuffer(x.w)
	}
}

func (x *Composer) initializeMap(st *bsttype.Map) {
	// 1. Set up the base type to the map.
	x.baseType = st

	// 2. Set up current pointer to the Key of the map.
	x.elemType = st.Key.Type
	x.elemDesc = st.Key.Descending
	if x.opts.Descending {
		x.elemDesc = !x.elemDesc
	}
	x.isKey = true

	// 3. If the map size was not specified, set the maximum index to MaxInt and wrap the writer with a buffer.
	//    The composer needs to be closed for undefined size maps.
	if !x.definedLength || x.opts.Comparable {
		x.maxIndex = math.MaxInt
		x.w = iopool.GetBuffer(x.w)
	}
}

func (x *Composer) verifyMapBase() error {
	if x.maxIndex < 0 {
		return bsterr.Err(bsterr.CodeInvalidValue, "undefined map size")
	}
	return nil
}

func (x *Composer) verifyArrayBase() error {
	if x.maxIndex < 0 {
		return bsterr.Err(bsterr.CodeInvalidValue, "undefined array size")
	}

	at := x.baseType.(*bsttype.Array)
	if at.HasFixedSize() && x.maxIndex != int(at.FixedSize)-1 {
		return bsterr.Err(bsterr.CodeInvalidValue, "array fixed size mismatch")
	}
	return nil
}

func (x *Composer) verifyStructBase() error {
	st := x.baseType.(*bsttype.Struct)
	if x.maxIndex != len(st.Fields)-1 {
		return bsterr.Err(bsterr.CodeInvalidValue, "invalid struct size")
	}

	return nil
}

func (x *Composer) verifyDefaultBase() error {
	return nil
}

func (x *Composer) writeMapLength() error {
	// 1. Write the length of the map.
	n, err := bstio.WriteUint(x.w, uint(x.maxIndex+1), x.opts.Descending)
	if err != nil {
		return err
	}

	x.bytesWritten += n
	return nil
}

func (x *Composer) writeArrayLength(bt *bsttype.Array) error {
	// 1. Check if the array has fixed size.
	if bt.HasFixedSize() {
		return nil
	}

	// 2. Write the length of the array.
	n, err := bstio.WriteUint(x.w, uint(x.maxIndex+1), x.opts.Descending)
	if err != nil {
		return err
	}

	x.bytesWritten += n
	return nil
}

func (x *Composer) writeFieldHeader(w io.Writer, index uint, size uint) (int, error) {
	// 1. Write field buffIndex.
	n, err := bstio.WriteUint(w, index, false)
	if err != nil {
		return 0, err
	}
	bytesWritten := n
	// 2. Write field size.
	n, err = bstio.WriteUint(w, size, false)
	if err != nil {
		return bytesWritten, err
	}
	bytesWritten += n

	// 3. Mark the field header as written.
	x.fhWritten = true
	return bytesWritten, nil
}

func (x *Composer) setFieldBuffer() {
	buf := iopool.GetBuffer(x.w)
	x.w = buf
	x.bufWrites = true
}

func (x *Composer) closeStruct(et *bsttype.Struct) error {
	if !x.externalModules && x.modules != nil {
		defer x.modules.Free()
	}
	if !x.opts.CompatibilityMode {
		return nil
	}

	if x.fhWritten {
		return nil
	}
	if err := x.finishStructElem(et); err != nil {
		return err
	}
	return nil
}

func (x *Composer) closeArray(bt *bsttype.Array) error {
	// 1. Nothing needs to be done for the fixed size arrays.
	if bt.HasFixedSize() || (x.definedLength && !x.opts.Comparable) {
		// 1.1 Mark the array composer as done.
		x.done = true
		return nil
	}

	// 2. Check if last value that was written was a boolean.
	if x.boolBufPos > 0 {
		if err := bstio.WriteByte(x.w, x.boolBuf); err != nil {
			return err
		}
		x.boolBufPos = 0
		x.boolBuf = 0x00
		x.bytesWritten++
	}

	// 3. Variable size array was written to the buffer, and its length
	//     was not written.
	sb, ok := x.w.(*iopool.SharedBuffer)
	if !ok {
		return bsterr.Err(bsterr.CodeWritingFailed, "")
	}

	root := sb.Root

	if !x.opts.Comparable {
		// 4.1. If the value is non-comparable an array length is written.
		n, err := bstio.WriteUint(root, uint(x.index), x.opts.Descending)
		if err != nil {
			return err
		}

		x.bytesWritten += n

		// 4.2. Write the array to the buffer.
		_, err = sb.WriteTo(root)
		if err != nil {
			return err
		}
	} else {
		// 5. For comparable arrays, shared buffer data is stored as comparable bytes.
		n, err := bstio.WriteBufferedBytesInternalComparable(root, sb, bstio.ArrayEscape, x.elemDesc)
		if err != nil {
			return err
		}
		// 5.1. The number of bytes written is a difference between the number of bytes written
		//      by above function and the number of bytes written by the shared buffer.
		//      This is because a shared buffer bytes were already counted on each element.
		x.bytesWritten += n - len(sb.Bytes)
	}

	// 6. Reset the buffer.
	x.w = root

	// 7. Release the buffer.
	iopool.ReleaseBuffer(sb)

	// 8. Mark the array composer as done.
	x.done = true

	return nil
}

func (x *Composer) closeMap() error {
	// 1. Verify if both the map key and value were written.
	if !x.isKey {
		return bsterr.Err(bsterr.CodeWritingFailed, "cannot close the composer with written key without value pair")
	}

	// 2. Check if last value that was written was a boolean.
	if x.boolBufPos > 0 {
		if err := bstio.WriteByte(x.w, x.boolBuf); err != nil {
			return err
		}
		x.boolBufPos = 0
		x.boolBuf = 0x00
		x.bytesWritten++
	}

	// 2. If the length was already defined, nothing needs to be done.
	if x.definedLength && !x.opts.Comparable {
		// 2.1. Mark the map composer as done.
		x.done = true
		return nil
	}

	// 2. Variable size array was written to the buffer, and its length
	//     was not written.
	sb, ok := x.w.(*iopool.SharedBuffer)
	if !ok {
		return bsterr.Err(bsterr.CodeWritingFailed, "")
	}
	root := sb.Root

	if !x.opts.Comparable {
		// 3.1. If the value is non-comparable a map length is written.
		n, err := bstio.WriteUint(root, uint(x.index), x.opts.Descending)
		if err != nil {
			return err
		}

		x.bytesWritten += n

		// 3.2. Write the map to the buffer.
		_, err = sb.WriteTo(root)
		if err != nil {
			return err
		}
	} else {
		// 4.1. For comparable maps, shared buffer data is stored as comparable bytes.
		n, err := bstio.WriteBufferedBytesInternalComparable(root, sb, bstio.MapEscape, x.elemDesc)
		if err != nil {
			return err
		}
		// 4.2. The number of bytes written is a difference between the number of bytes written
		//      by above function and the number of bytes written by the shared buffer.
		//      This is because a shared buffer bytes were already counted on each element.
		x.bytesWritten += n - len(sb.Bytes)
	}

	// 5. Reset the buffer.
	x.w = root

	// 6. Release the buffer.
	iopool.ReleaseBuffer(sb)

	// 7. Mark the map composer as done.
	x.done = true

	return nil
}

func (x *Composer) needWriteFieldHeader() bool {
	if x.baseType == nil {
		return false
	}
	return x.baseType.Kind() == bsttype.KindStruct && x.opts.CompatibilityMode && !x.bufWrites
}

func (x *Composer) writeHeader() error {
	// 1. The composer header is a byte that contains following flags.
	var h byte

	// 2. 0th bit - if the type is embedded.
	if x.opts.EmbedType {
		h |= 1 << 0
	}

	// 3. 1st bit - if the compatibility mode is enabled.
	if x.opts.CompatibilityMode {
		h |= 1 << 1
	}

	// 4. 2nd bit - if the value is stored in comparable format.
	if x.opts.Comparable {
		h |= 1 << 2
	}

	// 5. 3rd bit - if the value is stored in descending order.
	if x.opts.Descending {
		h |= 1 << 3
	}

	// 6. 4th bit - modules are embed.
	if x.opts.EmbedType && x.modules != nil {
		h |= 1 << 4
	}

	// 6. Write the header.
	if err := bstio.WriteByte(x.w, h); err != nil {
		return err
	}
	x.bytesWritten++

	// 7. If the type is embedded, write the type binary just after the header.
	if x.opts.EmbedType {
		// 7.1. Write modules binary.
		if x.modules != nil {
			n, err := x.modules.Write(x.w)
			if err != nil {
				return err
			}
			x.bytesWritten += n
		}

		// 7.2. Write the binary of the type that will be encoded.
		n, err := bsttype.WriteType(x.w, x.baseType)
		if err != nil {
			return err
		}

		x.bytesWritten += n
	}

	return nil
}

func (x *Composer) initializeBasicComposer(bt bsttype.Type, header bool) error {
	// 1. Initialize the composer for basic types.
	x.baseType = bt
	x.elemType = bt

	var err error
	if header {
		if err = x.writeHeader(); err != nil {
			return err
		}
	}

	// 5.3. verify if the composer is valid.
	if err = x.verifyDefaultBase(); err != nil {
		return err
	}
	return nil
}

func (x *Composer) writeStructHeader() error {
	// 1. Write the max index of the struct.
	n, err := bstio.WriteUint(x.w, uint(x.maxIndex), x.opts.Descending)
	if err != nil {
		return bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "writing struct header failed")
	}
	x.bytesWritten += n
	return nil
}

func (x *Composer) applyOptions(opts ComposerOptions) error {
	x.opts = opts
	if opts.Modules != nil {
		x.modules = opts.Modules
		x.externalModules = true
	}
	x.opts.EmbedType = opts.EmbedType
	if opts.Length != 0 {
		x.definedLength = true
		x.maxIndex = opts.Length - 1
	}
	return nil
}
