package bst

import (
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bstskip"
	"github.com/devmodules/bst/bsttype"
	"github.com/devmodules/bst/internal/iopool"
)

// ExtractorOptions is a set of options used for the extractor.
type ExtractorOptions struct {
	Headless          bool
	Descending        bool
	Comparable        bool
	CompatibilityMode bool
	ExpectedType      bsttype.Type
	Modules           *bsttype.Modules
}

// Extractor is binary serializable type extractor.
// Each element could be extracted by calling Next() and then Skip() or ReadXXX(), where XXX is the type
// of the value i.e. ReadInt64(), ReadString(), ReadBoolean(), etc.
//
// Nullable values needs to be either skipped or dereferenced by calling Deref().
// If the value is null, then nothing needs to be done afterwards. However, if the value is not null,
// it either needs to be read or skipped (after dereference).
//
// For element complex types (Struct, Array, Map) the method SubExtractor() can be used for
// extracting that element elements. This prevents for new allocation of the extractor, as the
// one on which the method is called is reused.
// In order to optimize the allocated memory, the extractor could be reused for the next type.
type Extractor struct {
	embedType, elemType                       bsttype.Type
	index, maxIndex                           int
	embed                                     extractorBaseStatus
	isKey, keyDone, elemDone, baseDone        bool
	r                                         io.ReadSeeker
	opts                                      ExtractorOptions
	boolBuf                                   byte
	boolBufPosition                           int
	headerRead, elemDesc                      bool
	bytesRead                                 int
	err                                       error
	fieldHeader                               fieldHeader
	clearElemFn                               func()
	clearModules, clearEmbedType, clearReader bool
}

type extractorBaseStatus struct {
	index    int
	maxIndex int

	elemType bsttype.Type
	used     bool
}

// NewExtractor creates a new value extractor read from the input reader, of the given type.
func NewExtractor(r io.Reader, opts ExtractorOptions) (*Extractor, error) {
	var (
		rs              io.ReadSeeker
		ok, clearReader bool
	)
	// 1. Check if the reader is not a read seeker and if so, wrap it in as a shared read seeker.
	if rs, ok = r.(io.ReadSeeker); !ok {
		rs = iopool.WrapReader(r)
		clearReader = true
	}

	// 2. Define the extractor.
	x := &Extractor{r: rs, clearReader: clearReader}

	// 3. Initialize the extractor with provided options.
	if err := x.init(opts); err != nil {
		return nil, err
	}
	return x, nil
}

// BytesRead returns the number of the bytes read during extraction.
func (x *Extractor) BytesRead() int {
	return x.bytesRead
}

// Comparable returns true if the data is encoded in an ordered (comparable) fashion.
func (x *Extractor) Comparable() bool {
	return x.opts.Comparable
}

// Close finishes up extraction of the binary values.
// This method should be called after the last call to Next().
// It releases all resources allocated by the extractor.
// This function could be called asynchronously once all extractions are done.
func (x *Extractor) Close() {
	// 1.  The close of the extractor should clear all the shared and releasable resources.
	//     At first check if the reader is shared and if so, release it.
	if x.clearReader {
		rs := x.r.(*iopool.SharedReadSeeker)
		iopool.ReleaseReadSeeker(rs)
	}

	// 2. Clear the modules if they were allocated as shared.
	if x.clearModules {
		x.opts.Modules.Free()
	}

	// 3. Clear the embed type if it was allocated as shared.
	if x.clearEmbedType {
		bsttype.PutSharedType(x.embedType)
	}
}

// EmbedType returns the type of the embedded value.
func (x *Extractor) EmbedType() bsttype.Type {
	return x.embedType
}

// Err returns the last error that occurred in the next.
func (x *Extractor) Err() error {
	return x.err
}

func (x *Extractor) readHeader() error {
	// 1. Check if the header was not already read.
	if x.headerRead {
		return bsterr.Err(bsterr.CodeAlreadyRead, "data header is already read")
	}

	// 2. The first byte of the input stream is expected to contain metadata about the value.
	bt, err := bstio.ReadByte(x.r)
	if err != nil {
		return bsterr.Err(bsterr.CodeReadingFailed, "failed to read data header")
	}
	x.bytesRead++

	// 3. The header bits are used in following way:
	//    - Bit 0: The data stream type is embedded.
	//    - Bit 1: Compatibility mode is used.
	//    - Bit 2: Value is stored in comparable fashion
	//    - Bit 3: Value is stored in descending order
	//    - Bit 4: Modules embed.
	var typeEmbed bool

	// 3.1. 0th bit is used to determine if the data is embedded.
	if bt&0x01 != 0 {
		typeEmbed = true
	}

	// 3.2. 1st bit - compatibility mode.
	if (bt>>1)&0x1 != 0 {
		x.opts.CompatibilityMode = true
	}

	// 3.3. 2nd bit - determines comparable format of values.
	if (bt>>2)&0x1 != 0 {
		x.opts.Comparable = true
	}

	// 3.3. 3rd bit - determines if the value is stored in descending order.
	if (bt>>3)&0x1 != 0 {
		x.opts.Descending = true
	}

	// 3.4. 4th bit - determines if modules are embedded.
	var modulesEmbed bool
	if (bt>>4)&0x01 != 0 {
		modulesEmbed = true
	}

	if modulesEmbed {
		// 4. Read, the modules embed in the header.
		m := bsttype.GetSharedModules()
		var n int
		n, err = m.Read(x.r, true)
		if err != nil {
			return err
		}
		x.bytesRead += n

		if x.opts.Modules == nil {
			// 4.1. If the modules are not defined yet, set them into the context of the extractor.
			x.opts.Modules = m
		} else {
			// 4.2. Otherwise, merge modules provided by the user into the modules read from the header.
			//      This way, user input modules are not changed.
			m.Merge(x.opts.Modules)
		}
		x.clearModules = true
	}

	// 5. If the type is not embed we can stop here.
	if typeEmbed {
		// 6. If the data stream type is embedded, try to read the type.
		var (
			et bsttype.Type
			n  int
		)
		et, n, err = bsttype.ReadType(x.r, true)
		if err != nil {
			return err
		}
		x.bytesRead += n

		x.embedType = et
		x.clearEmbedType = true
	}

	// 7. Set up embed type and mark the extractor header as read.
	x.headerRead = true
	return nil
}

// ResetTo reuses the extractor for the needs of the input type.
func (x *Extractor) ResetTo(r io.Reader, opts ExtractorOptions) error {
	var (
		rs              io.ReadSeeker
		ok, clearReader bool
	)
	// 1. Check if the reader is not a read seeker and if so, wrap it in as a shared read seeker.
	if rs, ok = r.(io.ReadSeeker); !ok {
		rs = iopool.WrapReader(r)
		clearReader = true
	}
	*x = Extractor{r: rs, clearReader: clearReader}

	// 2. Initialize it.
	if err := x.init(opts); err != nil {
		return err
	}
	return nil
}

// Next advances the extractor to the next field.
func (x *Extractor) Next() bool {
	// 1. Check if the error occurred in the previous step.
	if x.err != nil {
		return false
	}

	if x.baseDone {
		return false
	}

	// 2. Switch by the kind of embedded type.
	switch x.embedType.Kind() {
	case bsttype.KindArray:
		return x.nextArrayElem()
	case bsttype.KindMap:
		return x.nextMapElem()
	case bsttype.KindStruct:
		return x.nextStructElem()
	default:
		// This is about the basic type.
		return x.nextDefaultElem()
	}
}

// KeyDone marks the current key as done.
func (x *Extractor) KeyDone() bool {
	return x.keyDone
}

// ValueDone marks the current element as done.
func (x *Extractor) ValueDone() bool {
	return x.elemDone
}

// Index returns the buffIndex of the current field.
func (x *Extractor) Index() int {
	return x.index
}

// MaxIndex returns the maximum buffIndex of the current extractor.
func (x *Extractor) MaxIndex() int {
	return x.maxIndex
}

// Length returns the length of the current field.
func (x *Extractor) Length() int {
	if _, ok := x.embedType.(*bsttype.Struct); ok {
		panic("cannot get length of struct")
	}
	return x.maxIndex + 1
}

// Elem returns the type of the current field.
func (x *Extractor) Elem() bsttype.Type {
	return x.elemType
}

// Skip skips the field from the extractor.
// For map types this skips both the key and the value.
func (x *Extractor) Skip() (int64, error) {
	if x.elemDone {
		return 0, bsterr.Err(bsterr.CodeAlreadyRead, "data element was already read")
	}
	if x.index > x.maxIndex {
		return 0, bsterr.Err(bsterr.CodeOutOfBounds, "buffIndex out of bounds")
	}

	var skipped int64

	skipFunc := bstskip.SkipFuncOf(x.elemType)
	opts := bstio.ValueOptions{
		Comparable: x.opts.Comparable,
		Descending: x.opts.Descending,
	}
	if x.elemDesc {
		opts.Descending = !opts.Descending
	}
	n, err := skipFunc(x.r, opts)
	if err != nil {
		return 0, err
	}
	skipped += n
	x.bytesRead += int(n)
	x.finishElem()

	return skipped, nil
}

// reset current extractor to the initial state
func (x *Extractor) reset() {
	*x = Extractor{
		r:     x.r,
		opts:  x.opts,
		index: -1,
	}
}

func (x *Extractor) previewPrevElem() (bsttype.Type, bool) {
	switch x.embedType.Kind() {
	case bsttype.KindStruct:
		return x.previewPrevStructElem()
	case bsttype.KindArray:
		return x.previewPrevArrayElem()
	case bsttype.KindMap:
		return x.previewPrevMapType()
	default:
		return nil, false
	}
}

func (x *Extractor) init(options ExtractorOptions) error {
	// 1. Apply provided options.
	x.opts = options

	// 3. Verify if the extractor is formed in a valid way.
	if err := x.validate(); err != nil {
		return err
	}

	// 4. If the extractor is not headless, then read the header.
	if !x.opts.Headless {
		if err := x.readHeader(); err != nil {
			return err
		}
	}

	// 5. If the embed type is not provided then set it from the expected type.
	if x.embedType == nil {
		// 5.1. Check if the expected type was set up from the input options.
		if x.opts.ExpectedType == nil {
			return bsterr.Err(bsterr.CodeInvalidType, "no expected type provided for the extractor and no embed type encoded in the stream")
		}
		x.embedType = x.opts.ExpectedType
	}

	// 6. Initialize extractor for its type.
	switch x.embedType.Kind() {
	case bsttype.KindStruct:
		return x.initStructBase()
	case bsttype.KindArray:
		return x.initializeArray()
	case bsttype.KindMap:
		return x.initializeMap()
	case bsttype.KindNamed:
		return x.initializeNamed()
	default:
		return x.initializeDefault()
	}
}

type wrappedReader struct {
	*iopool.SharedReadSeeker
	root io.ReadSeeker
}

func (x *Extractor) finishElem() {
	switch x.embedType.Kind() {
	case bsttype.KindStruct:
		x.finishStructElem()
	case bsttype.KindArray:
		x.finishArrayElem()
	case bsttype.KindMap:
		x.finishMapElem()
	default:
		x.finishDefaultElem()
	}
}

func (x *Extractor) validate() error {
	if x.opts.Headless && x.embedType == nil && x.opts.ExpectedType == nil {
		return bsterr.Err(bsterr.CodeInvalidType, "no base type provided for headless data extractor")
	}
	return nil
}

//
// Named type
//

func (x *Extractor) initializeNamed() error {
	// 1. Dereference the type to extract the underlying type.
	nt := x.embedType.(*bsttype.Named)
	if nt.Type != nil {
		x.elemType = nt.Type
		return nil
	}

	// 2. If the named type is not resolved, check if the modules are defined.
	if x.opts.Modules == nil {
		x.err = bsterr.Err(bsterr.CodeInvalidType, "no modules provided for named type")
		return x.err
	}

	// 3. Check if the modules are resolved, and if not resolve them.
	if !x.opts.Modules.IsResolved() {
		if err := x.opts.Modules.Resolve(); err != nil {
			x.err = err
			return x.err
		}
	}

	// 3. Try to resolve the named type from the modules.
	if _, err := nt.ResolveDependencies(x.opts.Modules); err != nil {
		x.err = err
		return err
	}
	x.elemType = nt.Type
	return nil
}

//
// Default type
//

func (x *Extractor) initializeDefault() error {
	x.elemType = x.embedType
	return nil
}

func (x *Extractor) nextDefaultElem() bool {
	return !x.elemDone
}

func (x *Extractor) finishDefaultElem() {
	x.elemDone = true
	x.baseDone = true
}

func (x *Extractor) derefType(et bsttype.Type) (bsttype.Type, error) {
	nt, ok := et.(*bsttype.Named)
	if !ok {
		return et, nil
	}
	var named *bsttype.Named
	for {
		named, ok = nt.Type.(*bsttype.Named)
		if !ok {
			return nt.Type, nil
		}
		nt = named
	}
}
