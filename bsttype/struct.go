package bsttype

import (
	"io"
	"strconv"
	"strings"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
)

// Compile time checks for the Type interfaces.
var (
	_ Type         = (*Struct)(nil)
	_ TypeSkipper  = (*Struct)(nil)
	_ TypeWriter   = (*Struct)(nil)
	_ TypeReader   = (*Struct)(nil)
	_ TypeComparer = (*Struct)(nil)
)

// Compile-time checks for Dependency interfaces.
var (
	_ DependencyOperator = (*Struct)(nil)
	_ DependencyChecker  = (*Struct)(nil)
	_ DependencyComposer = (*Struct)(nil)
	_ DependencyNeeder   = (*Struct)(nil)
	_ DependencyVerifier = (*Struct)(nil)
	_ DependencyResolver = (*Struct)(nil)
)

// Compile-time checks for internal interfaces
var (
	_ copier        = (*Struct)(nil)
	_ refCounter    = (*Struct)(nil)
	_ cycleDetector = (*Struct)(nil)
)

type (
	// Struct is the Type implementation for the struct value.
	// It is used to describe the structure of a document with a set of fields.
	// Binary representation looks like:
	// Size(bits)   | Name               | Description
	// -------------+--------------------+------------
	//    8		    | Field Count Size   | Header with the size of the field count.
	//    8-64      | Field Count        | The number of fields in the struct.
	//    N * Count | Fields elements    | Binary representation of the fields.
	Struct struct {
		Fields []StructField

		needsRelease bool
	}

	// StructField is a single field in a struct.
	// Each struct field binary representation is:
	// Size(bits) | Name			   | Description
	// -----------+--------------------+------------
	//    8       | Index binary size  | Header with the size of the index binary size.
	//    0-64    | Index binary       | The index of the field in the struct.
	//    8       | Name Length Size   | Header with the size of the name length.
	//    0 - 64  | Name Length        | The length of the string (0 if the length size was marked as 0x00).
	//    0 - N   | Name               | The name of the field (0 if the field is undefined)
	//    1       | Descending flag    | The flag to indicate if the field is descending.
	//    2       | -				   | Padding to align field type within a single byte.
	//    5       | Type               | The type of the field.
	//    0 - N   | Type Content       | The content of the type - optional if Type is not basic.
	StructField struct {
		// Index is the identifier of the struct field.
		Index uint
		// Name is the name of the field.
		Name string
		// Descending is a flag that determines if given field is encoded in descending order.
		Descending bool
		// Type is the type of the field.
		Type Type
	}
)

// Kind returns the basic kind of the value.
func (*Struct) Kind() Kind {
	return KindStruct
}

// String
func (x *Struct) String() string {
	sb := strings.Builder{}

	sb.WriteString("struct {")
	for i, f := range x.Fields {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(strconv.FormatInt(int64(f.Index), 10))
		sb.WriteRune(' ')
		sb.WriteString(f.Name)
		sb.WriteString(": ")
		sb.WriteString(f.Type.String())
	}
	sb.WriteRune('}')
	return sb.String()
}

// SkipType skips the bytes in the reader to the next value.
// Implements the TypeSkipper interface.
func (x *Struct) SkipType(rs io.ReadSeeker) (int64, error) {
	// 1. Read the number of fields.
	length, bl, err := bstio.ReadUint(rs, false)
	if err != nil {
		return int64(bl), bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read struct type field length")
	}
	bytesSkipped := int64(bl)

	// 2. Skip the fields.
	var n int64
	for i := uint(0); i < length; i++ {
		// 2.1. Skip the index.
		n, err = bstio.SkipUint(rs, false)
		if err != nil {
			return n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip struct field index")
		}
		bytesSkipped += n

		// 2.2. Skip the name of the field.
		n, err = bstio.SkipNonComparableString(rs, false)
		if err != nil {
			return bytesSkipped, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip struct field name")
		}
		bytesSkipped += n

		// 2.3. Skip the type of the field.
		n, err = SkipType(rs)
		if err != nil {
			return bytesSkipped, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip struct field type")
		}
		bytesSkipped += n
	}
	return bytesSkipped, nil
}

// ReadType reads the value from the byte slice.
// Implements the TypeReader interface.
func (x *Struct) ReadType(r io.Reader) (int, error) {
	// 1. Read the number of fields.
	fl, bl, err := bstio.ReadUint(r, false)
	if err != nil {
		return bl, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read struct type field length")
	}
	bytesRead := bl

	if fl == 0 {
		return bytesRead, nil
	}

	// 2. Initialize the fields.
	x.Fields = make([]StructField, fl)

	// 3. Read the fields.
	var (
		tp         Type
		descending bool
		index      uint
		n          int
		name       string
	)
	for i := uint(0); i < fl; i++ {
		// 3.1. Read the field index.
		index, n, err = bstio.ReadUint(r, false)
		if err != nil {
			return bytesRead, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read struct field index")
		}
		bytesRead += n

		// 3.2. Read the name of the field.
		name, n, err = bstio.ReadStringNonComparable(r, false)
		if err != nil {
			return n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read struct field name")
		}
		bytesRead += n

		// 3.3. Read the byte for the type of the field along with the descending flag.
		tp, descending, n, err = readFieldType(r)
		if err != nil {
			return n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read struct field type")
		}
		bytesRead += n

		x.Fields[i] = StructField{
			Index:      index,
			Name:       name,
			Type:       tp,
			Descending: descending,
		}
	}
	return bytesRead, nil
}

func readFieldType(r io.Reader) (Type, bool, int, error) {
	// 1. Read the header byte.
	bt, err := bstio.ReadByte(r)
	if err != nil {
		return nil, false, 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read struct field type")
	}
	total := 1

	// 2. First bit specifies if the field is descending.
	descending := bt&0x80 != 0

	// 3. Trim the first bit and initialize an empty type.
	bt &^= 0x80
	et := emptyKindType(Kind(bt), false)

	// 4. Check if the type ha a ReadType function.
	tr, ok := et.(TypeReader)
	if !ok {
		return et, descending, total, nil
	}
	n, err := tr.ReadType(r)
	if err != nil {
		return nil, false, 0, err
	}
	return et, descending, total + n, nil
}

// WriteType writes the value to the byte slice.
func (x *Struct) WriteType(w io.Writer) (int, error) {
	// 1. Write the number of fields.
	n, err := bstio.WriteUint(w, uint(len(x.Fields)), false)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write struct type field length")
	}
	bytesWritten := n

	// 2. Write the fields.
	for _, f := range x.Fields {
		// 2.1. Write the index.
		n, err = bstio.WriteUint(w, f.Index, false)
		if err != nil {
			return bytesWritten, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write struct field index")
		}
		bytesWritten += n

		// 2.2. Write the name of the field.
		n, err = bstio.WriteString(w, f.Name, false, false)
		if err != nil {
			return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write struct field name")
		}
		bytesWritten += n

		// 2.3. Write the type of the field.
		n, err = writeFieldType(w, f.Type, f.Descending)
		if err != nil {
			return n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write struct field type")
		}
		bytesWritten += n
	}

	return bytesWritten, nil
}

func writeFieldType(w io.Writer, vt Type, desc bool) (int, error) {
	// 1. Convert the type kind to the byte.
	fk := byte(vt.Kind())

	// 2. If the type is descending, set the descending flag for the first MSB.
	if desc {
		fk |= 0x80
	}

	// 3. Write the type byte.
	if err := bstio.WriteByte(w, fk); err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write type").
			WithDetail("type", vt.Kind())
	}
	total := 1

	// 4.  If the type implements TypeContent interface, write the content.
	tc, ok := vt.(TypeWriter)
	if !ok {
		return total, nil
	}

	bw, err := tc.WriteType(w)
	if err != nil {
		return total + bw, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write type content").
			WithDetail("type", vt.Kind())
	}
	return total + bw, nil
}

// CompareType returns true if the types are equal.
// Implements the TypeComparer interface.
func (x *Struct) CompareType(to TypeComparer) bool {
	xto, ok := to.(*Struct)
	if !ok {
		return false
	}

	if len(x.Fields) != len(xto.Fields) {
		return false
	}

	for i := range x.Fields {
		if x.Fields[i].Name != xto.Fields[i].Name {
			return false
		}

		if !TypesEqual(x.Fields[i].Type, xto.Fields[i].Type) {
			return false
		}
	}
	return true
}

// PreviewPrevElemType returns the type of the previous element.
func (x *Struct) PreviewPrevElemType(i int) (Type, bool) {
	if i < 0 || i >= len(x.Fields) {
		return nil, false
	}
	return x.Fields[i].Type, true
}

// CheckDependencies iterates over all fields and tries to check all named type dependency within given modules.
// Implements DependencyChecker interface.
func (x *Struct) CheckDependencies(m *Modules) (CheckDependenciesResult, error) {
	var res CheckDependenciesResult
	for _, f := range x.Fields {
		dc, ok := f.Type.(DependencyChecker)
		if !ok {
			continue
		}

		fieldRes, err := dc.CheckDependencies(m)
		if err != nil {
			return CheckDependenciesResult{}, err
		}
		res.ResolveRequired = res.ResolveRequired || fieldRes.ResolveRequired
		res.ComposeRequired = res.ComposeRequired || fieldRes.ComposeRequired
	}
	return res, nil
}

// ComposeDependencies iterates over all fields and tries to compose all named type dependency within given modules.
// Implements DependencyComposer interface.
func (x *Struct) ComposeDependencies(m *Modules) error {
	for _, f := range x.Fields {
		dc, ok := f.Type.(DependencyComposer)
		if !ok {
			continue
		}
		if err := dc.ComposeDependencies(m); err != nil {
			return err
		}
	}
	return nil
}

// NeedsDependencies returns true if any of the struct fields requires modules dependencies.
// Implements DependencyNeeder interface.
func (x *Struct) NeedsDependencies() bool {
	for _, f := range x.Fields {
		dc, ok := f.Type.(DependencyNeeder)
		if !ok {
			continue
		}

		if dc.NeedsDependencies() {
			return true
		}
	}
	return false
}

// VerifyDependencies iterates over all fields and tries to verify all named type dependencies.
// Implements DependencyVerifier interface.
func (x *Struct) VerifyDependencies() error {
	for _, f := range x.Fields {
		dv, ok := f.Type.(DependencyVerifier)
		if !ok {
			continue
		}

		if err := dv.VerifyDependencies(); err != nil {
			return err
		}
	}
	return nil
}

var _ DependencyResolver = (*Struct)(nil)

// ResolveDependencies resolve references within given struct type, taken out of the Modules.
func (x *Struct) ResolveDependencies(m *Modules) (int64, error) {
	// 1. Iterate over all types that could be referenced.
	for _, f := range x.Fields {
		mr, ok := f.Type.(DependencyResolver)
		if !ok {
			continue
		}
		if _, err := mr.ResolveDependencies(m); err != nil {
			return 0, err
		}
	}
	return 0, nil
}

var _ refCounter = (*Struct)(nil)

func (x *Struct) countRefs() int64 {
	var refs int64
	for _, f := range x.Fields {
		cf, ok := f.Type.(refCounter)
		if !ok {
			continue
		}
		refs += cf.countRefs()
	}
	return refs
}

// detectCycles iterates over all fields and tries to detect cycles within given modules.
func (x *Struct) detectCycles(mod, name string) error {
	// 1. Iterate over all structure fields and check for cycles.
	for _, f := range x.Fields {
		// 2. A cycle in the struct could only be found on named type fields.
		nt, ok := f.Type.(*Named)
		if !ok {
			continue
		}

		// 3. Check if this named type points directly to the one provided in the input.
		if nt.Module == mod && nt.Name == name {
			return bsterr.Err(bsterr.CodeCyclicDependency, "cyclic dependency detected").
				WithDetails(
					bsterr.D("module", mod),
					bsterr.D("name", name),
				)
		}

		// 4. Otherwise, check if this named type points to a struct or oneOf, and if so, check for cycles.
		switch tp := nt.Type.(type) {
		case *Struct:
			if err := tp.detectCycles(mod, name); err != nil {
				// TODO: add thee field name as the root o the error to have a traceability of the cycles.
				//       This should little more complex as the error might not be directly extracted in the below type.
				return err
			}
		case *OneOf:
			if err := tp.detectCycles(mod, name); err != nil {
				// TODO: add thee field name as the root o the error to have a traceability of the cycles.
				//       This should little more complex as the error might not be directly extracted in the below type.
				return err
			}
		}
	}
	return nil
}

func (x *Struct) copy(shared bool) Type {
	var cp *Struct
	if shared {
		cp = getSharedStruct()
	} else {
		cp = new(Struct)
	}
	if cap(cp.Fields) < len(x.Fields) {
		cp.Fields = make([]StructField, len(x.Fields))
	} else {
		cp.Fields = cp.Fields[:len(x.Fields)]
	}

	for i, f := range x.Fields {
		cp.Fields[i] = StructField{
			Index:      f.Index,
			Name:       f.Name,
			Descending: f.Descending,
			Type:       f.Type.(copier).copy(shared),
		}
	}
	return cp
}

// Shared pool
var _sharedStructsPool = &sharedPool{defaultSize: 10}

func getSharedStruct() *Struct {
	v := _sharedStructsPool.pool.Get()
	st, ok := v.(*Struct)
	if ok {
		return st
	}
	return &Struct{
		Fields:       make([]StructField, 0, _sharedStructsPool.defaultSize),
		needsRelease: true,
	}
}

func putSharedStruct(x *Struct) {
	if !x.needsRelease {
		return
	}
	length := cap(x.Fields)
	*x = Struct{needsRelease: true, Fields: x.Fields[:0]}
	_sharedStructsPool.put(x, length)
}
