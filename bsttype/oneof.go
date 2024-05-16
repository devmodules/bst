package bsttype

import (
	"fmt"
	"io"
	"strings"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
)

// Compile time check if the OneOf implements Type interface.
var (
	_ Type         = (*OneOf)(nil)
	_ TypeSkipper  = (*OneOf)(nil)
	_ TypeReader   = (*OneOf)(nil)
	_ TypeWriter   = (*OneOf)(nil)
	_ TypeComparer = (*OneOf)(nil)
)

// Compile-time checks for Dependency interfaces.
var (
	_ DependencyOperator = (*OneOf)(nil)
	_ DependencyChecker  = (*OneOf)(nil)
	_ DependencyComposer = (*OneOf)(nil)
	_ DependencyNeeder   = (*OneOf)(nil)
	_ DependencyVerifier = (*OneOf)(nil)
	_ DependencyResolver = (*OneOf)(nil)
)

// Compile-time checks for internal interfaces
var (
	_ copier        = (*OneOf)(nil)
	_ cycleDetector = (*OneOf)(nil)
	_ refCounter    = (*OneOf)(nil)
)

// DefaultOneOfIndexBits is the default number of bits used to store the oneOf buffIndex.
const DefaultOneOfIndexBits = bstio.BinarySizeUint8

type (
	// OneOf is a type that can be marshaled into one of the types in the list.
	// Binary representation looks as:
	//    Size (bits)    |      Name        |    Description
	//    -------------- | ---------------- | ----------------
	//        8          | IndexBytes       | Number of bytes used to encode the index of the elements.
	//     N * Elements  | Elements         | List of elements.
	OneOf struct {
		// IndexBytes is the number of bits used to encode the buffIndex of the type element.
		IndexBytes uint8
		// Elements is a list of the types to which a value can be marshaled.
		Elements []OneOfElement

		needsRelease bool
	}

	// OneOfElement is a type that can be marshaled into one of the types in the list.
	// Binary representation looks as:
	//   Size (bits)    |      Name        |    Description
	//   -------------- | ---------------- | ----------------
	//   IndexBytes     | Index            | The index of the type element.
	//   8              | Name Length Size | The number of bytes used to encode the name length.
	//   0 - 64         | Name Length      | The length of the name.
	//   0 - N	        | Name             | The name of the type element.
	//   8              | Type Kind        | The kind of the type.
	//   0 - N          | Type Content     | (Optional) content of the one of type.
	OneOfElement struct {
		// Index is the unique integer for the OneOf that matches given element Type.
		// It is expected to work like an enum for the types.
		Index uint
		// Name is the human-readable name of the one-of element type.
		Name string
		// Type is the type of the element.
		Type Type
	}
)

// Kind returns the kind of the type.
func (o *OneOf) Kind() Kind {
	return KindOneOf
}

// String returns a human-readable string representation of the OneOf.
// Example: OneOfType.Name(Elements: [{Index: 0, Name: "ElementZero", Type: Type {Kind: KindString, Name: "string"}},{Index: 1, Name: "ElementOne", Type: Type {Kind: KindInt, Name: "string"}}], IndexBytes: 1)
func (o *OneOf) String() string {
	sb := strings.Builder{}
	sb.WriteString("OneOfType")
	sb.WriteString("(Elements: [")
	for i, e := range o.Elements {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("{Index: %d, Name: %s, Type: %s}", e.Index, e.Name, e.Type))
	}
	sb.WriteString("], IndexBytes: ")
	sb.WriteString(fmt.Sprintf("%d", o.IndexBytes))
	sb.WriteString(")")
	return sb.String()
}

// String returns a human-readable string representation of the OneOfElement.
func (x OneOfElement) String() string {
	return fmt.Sprintf("OneOfTypeElement {Index: %d, Name: %s, Type: %s}", x.Index, x.Name, x.Type)
}

// SkipType skips the type in the reader.
// This is used to skip the type in the reader when the type is not needed.
// Implements TypeSkipper interface.
func (o *OneOf) SkipType(rs io.ReadSeeker) (int64, error) {
	// 1. Read the number of bytes used to encode the buffIndex.
	indexBytes, err := bstio.ReadByte(rs)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to read oneOf type buffIndex bytes")
	}
	bytesSkipped := int64(1)

	// 2. Read the number of elements.
	var (
		temp uint
		n    int
	)
	temp, n, err = bstio.ReadUint(rs, false)
	if err != nil {
		return bytesSkipped + int64(n), bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to read oneOf length")
	}
	bytesSkipped += int64(n)

	// 4. Skip the elements.
	var n64 int64
	for i := uint(0); i < temp; i++ {
		// 4.1. Skip the buffIndex.
		switch indexBytes {
		case bstio.BinarySizeZero:
			n64, err = bstio.SkipUint(rs, false)
		case bstio.BinarySizeUint8:
			n64, err = bstio.SkipUint8Value(rs)
		case bstio.BinarySizeUint16:
			n64, err = bstio.SkipUint16(rs)
		case bstio.BinarySizeUint32:
			n64, err = bstio.SkipUint32(rs)
		case bstio.BinarySizeUint64:
			n64, err = bstio.SkipUint64(rs)
		default:
			return bytesSkipped, bsterr.Err(bsterr.CodeUndefinedType, "invalid oneOf buffIndex bytes").
				WithDetails(
					bsterr.D("oneOf", o),
					bsterr.D("indexBytes", indexBytes),
				)
		}
		if err != nil {
			return bytesSkipped + n64, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to skip oneOf buffIndex")
		}
		bytesSkipped += n64

		// 4.2. Skip the string name.
		n64, err = bstio.SkipNonComparableString(rs, false)
		if err != nil {
			return bytesSkipped + n64, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to skip oneOf name")
		}
		bytesSkipped += n64

		// 4.3. Skip the type.
		n64, err = SkipType(rs)
		if err != nil {
			return bytesSkipped + n64, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to skip oneOf type")
		}
		bytesSkipped += n64
	}
	return bytesSkipped, nil
}

// ReadType reads the type from the reader.
// Implements TypeReader interface.
func (o *OneOf) ReadType(r io.Reader) (int, error) {
	// 1. Read the number of bytes used to encode the buffIndex.
	bt, err := bstio.ReadByte(r)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to read oneOf type buffIndex bytes")
	}
	bytesRead := 1

	// 2. Set the number of bytes used to encode the buffIndex.
	o.IndexBytes = bt

	// 3. Read the number of elements.
	temp, ni, err := bstio.ReadUintValue(r, bt, false)
	if err != nil {
		return bytesRead + ni, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to read oneOf length").
			WithDetails(bsterr.D("oneOf", o))
	}
	bytesRead += ni

	// 4. Read the elements.
	var n int
	for i := uint(0); i < temp; i++ {
		// 4.1. Read the element buffIndex.
		var index uint
		switch o.IndexBytes {
		case bstio.BinarySizeZero:
			index, n, err = bstio.ReadUint(r, false)
		case bstio.BinarySizeUint8:
			var tu uint8
			tu, n, err = bstio.ReadUint8(r, false)
			index = uint(tu)
		case bstio.BinarySizeUint16:
			var tu uint16
			tu, n, err = bstio.ReadUint16(r, false)
			index = uint(tu)
		case bstio.BinarySizeUint32:
			var tu uint32
			tu, n, err = bstio.ReadUint32(r, false)
			index = uint(tu)
		case bstio.BinarySizeUint64:
			var tu uint64
			tu, n, err = bstio.ReadUint64(r, false)
			index = uint(tu)
		default:
			return bytesRead, bsterr.Err(bsterr.CodeUndefinedType, "invalid oneOf buffIndex bytes").
				WithDetails(
					bsterr.D("oneOf", o),
					bsterr.D("indexBytes", o.IndexBytes),
				)
		}
		if err != nil {
			return bytesRead + n, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to read oneOf buffIndex").
				WithDetails(
					bsterr.D("oneOf", o),
					bsterr.D("buffIndex", index),
				)
		}
		bytesRead += n

		// 4.2. Read the element name.
		var name string
		name, n, err = bstio.ReadStringNonComparable(r, false)
		if err != nil {
			return bytesRead + n, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to read oneOf name").
				WithDetails(
					bsterr.D("oneOf", o),
					bsterr.D("buffIndex", index),
				)
		}
		bytesRead += n

		// 4.3. Read the element type.
		var elem Type
		elem, n, err = ReadType(r, false)
		if err != nil {
			return bytesRead + n, bsterr.ErrWrap(err, bsterr.CodeReadingFailed, "failed to read oneOf element type").
				WithDetails(
					bsterr.D("oneOf", o),
					bsterr.D("buffIndex", index),
					bsterr.D("name", name),
				)
		}
		bytesRead += n

		o.Elements = append(o.Elements, OneOfElement{
			Index: index,
			Name:  name,
			Type:  elem,
		})
	}
	return bytesRead, nil
}

// WriteType writes the type to the writer.
// Implements TypeWriter interface.
func (o *OneOf) WriteType(w io.Writer) (int, error) {
	// 1. Write the number of bytes used to encode the buffIndex.
	if err := bstio.WriteByte(w, o.IndexBytes); err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write oneOf type buffIndex bytes size")
	}
	bytesWritten := 1

	// 3. Write the number of elements.
	n, err := bstio.WriteUintValue(w, uint(len(o.Elements)), o.IndexBytes, false)
	if err != nil {
		return bytesWritten + n, err
	}
	bytesWritten += n

	// 4. Write the elements.
	for i := range o.Elements {
		// 4.1. Write the element buffIndex.
		switch o.IndexBytes {
		case bstio.BinarySizeUint8:
			n, err = bstio.WriteUint8(w, uint8(o.Elements[i].Index), false)
		case bstio.BinarySizeUint16:
			n, err = bstio.WriteUint16(w, uint16(o.Elements[i].Index), false)
		case bstio.BinarySizeUint32:
			n, err = bstio.WriteUint32(w, uint32(o.Elements[i].Index), false)
		case bstio.BinarySizeUint64:
			n, err = bstio.WriteUint64(w, uint64(o.Elements[i].Index), false)
		case bstio.BinarySizeZero:
			n, err = bstio.WriteUint(w, o.Elements[i].Index, false)
		default:
			return bytesWritten, bsterr.Err(bsterr.CodeUndefinedType, "invalid oneOf buffIndex bytes").
				WithDetails(
					bsterr.D("oneOf", o),
					bsterr.D("indexBytes", o.IndexBytes),
				)
		}
		if err != nil {
			return bytesWritten + n, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write oneOf buffIndex").
				WithDetails(
					bsterr.D("oneOf", o),
					bsterr.D("buffIndex", o.Elements[i].Index),
				)
		}
		bytesWritten += n

		// 4.2. Write the element name.
		n, err = bstio.WriteString(w, o.Elements[i].Name, false, false)
		if err != nil {
			return bytesWritten + n, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write oneOf name").
				WithDetails(
					bsterr.D("oneOf", o),
					bsterr.D("buffIndex", o.Elements[i].Index),
					bsterr.D("name", o.Elements[i].Name),
				)
		}
		bytesWritten += n

		// 4.3. Write the element type.
		n, err = WriteType(w, o.Elements[i].Type)
		if err != nil {
			return bytesWritten + n, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write oneOf type").
				WithDetails(
					bsterr.D("oneOf", o),
					bsterr.D("buffIndex", o.Elements[i].Index),
					bsterr.D("name", o.Elements[i].Name),
					bsterr.D("type", o.Elements[i].Type),
				)
		}
		bytesWritten += n
	}
	return bytesWritten, nil
}

// CompareType returns true if the other type is equal to this type.
// Implements TypeComparer interface.
func (o *OneOf) CompareType(to TypeComparer) bool {
	// 1. Prevent panic if the other type is nil.
	if to == nil {
		return false
	}

	// 2. If the other type is not a OneOfType, return false.
	toO, ok := to.(*OneOf)
	if !ok {
		return false
	}

	// 3. If the number of bytes that is used to encode the buffIndex is not equal, return false.
	if o.IndexBytes != toO.IndexBytes {
		return false
	}

	// 4. If the number of elements is not equal, return false.
	if len(o.Elements) != len(toO.Elements) {
		return false
	}

	// 5. If the elements are not equal, return false.
	for i := range o.Elements {
		if o.Elements[i].Index != toO.Elements[i].Index {
			return false
		}

		if o.Elements[i].Name != toO.Elements[i].Name {
			return false
		}
		if !TypesEqual(o.Elements[i].Type, toO.Elements[i].Type) {
			return false
		}
	}

	// 6. All tests passed, return true.
	return true
}

//
// Dependencies
//

// CheckDependencies iterates over all the elements of the OneOf type and checks if any could be missing in the modules.
// Implements DependencyChecker interface.
func (o *OneOf) CheckDependencies(m *Modules) (CheckDependenciesResult, error) {
	var res CheckDependenciesResult
	for _, elem := range o.Elements {
		dc, ok := elem.Type.(DependencyChecker)
		if !ok {
			continue
		}
		elemRes, err := dc.CheckDependencies(m)
		if err != nil {
			return CheckDependenciesResult{}, err
		}

		res.ResolveRequired = res.ResolveRequired || elemRes.ResolveRequired
		res.ComposeRequired = res.ComposeRequired || elemRes.ComposeRequired
	}
	return res, nil
}

// ComposeDependencies iterates over all the elements of the OneOf type and checks if any could compose dependencies in the modules.
// Implements DependencyComposer interface.
func (o *OneOf) ComposeDependencies(m *Modules) error {
	for _, elem := range o.Elements {
		dc, ok := elem.Type.(DependencyComposer)
		if !ok {
			continue
		}

		if err := dc.ComposeDependencies(m); err != nil {
			return err
		}
	}
	return nil
}

// NeedsDependencies returns true if the type needs dependencies.
// Implements DependencyNeeder interface.
func (o *OneOf) NeedsDependencies() bool {
	for _, elem := range o.Elements {
		nd, ok := elem.Type.(DependencyNeeder)
		if !ok {
			continue
		}
		if nd.NeedsDependencies() {
			return true
		}
	}
	return false
}

// VerifyDependencies iterates over all the elements of the OneOf type and checks if any of it not well-defined.
// Implements DependencyVerifier interface.
func (o *OneOf) VerifyDependencies() error {
	for _, elem := range o.Elements {
		dv, ok := elem.Type.(DependencyVerifier)
		if !ok {
			continue
		}
		if err := dv.VerifyDependencies(); err != nil {
			return err
		}
	}
	return nil
}

// ResolveDependencies resolves possible references in the OneOf elements.
// Implements DependencyResolver interface.
func (o *OneOf) ResolveDependencies(m *Modules) (int64, error) {
	for _, elem := range o.Elements {
		mr, ok := elem.Type.(DependencyResolver)
		if !ok {
			continue
		}
		if _, err := mr.ResolveDependencies(m); err != nil {
			return 0, err
		}
	}
	return 0, nil
}

func (o *OneOf) detectCycles(mod, name string) error {
	// 1. Iterate over all elements of the OneOf definition.
	for _, elem := range o.Elements {
		// 2. A cycle could occur if the element is a named type.
		nt, ok := elem.Type.(*Named)
		if !ok {
			continue
		}

		// 3. If the named.Module and Name are equal to the current module and name, the cycle is detected.
		if nt.Name == name && nt.Module == mod {
			return bsterr.Err(bsterr.CodeCyclicDependency, "cycle detected in oneOf element").
				WithDetails(
					bsterr.D("mod", mod),
					bsterr.D("name", elem.Name),
					bsterr.D("oneOfIndex", elem.Index),
				)
		}

		// 4. Otherwise, check if this named type points to a struct, and if so, check for indirect cycles.
		st, ok := nt.Type.(*Struct)
		if !ok {
			continue
		}
		if err := st.detectCycles(mod, name); err != nil {
			// TODO: add thee field name as the root o the error to have a traceability of the cycles.
			//       This should little more complex as the error might not be directly extracted in the below type.
			return err
		}
	}
	return nil
}

func (o *OneOf) countRefs() int64 {
	var refs int64
	for _, elem := range o.Elements {
		mr, ok := elem.Type.(refCounter)
		if !ok {
			continue
		}
		refs += mr.countRefs()
	}
	return refs
}

func (o *OneOf) copy(shared bool) Type {
	var cp *OneOf
	if shared {
		cp = getSharedOneOf()
	} else {
		cp = new(OneOf)
	}
	cp.IndexBytes = o.IndexBytes
	if cap(cp.Elements) < len(o.Elements) {
		cp.Elements = make([]OneOfElement, len(o.Elements))
	} else {
		cp.Elements = cp.Elements[:len(o.Elements)]
	}
	for i, elem := range o.Elements {
		cp.Elements[i] = OneOfElement{
			Index: elem.Index,
			Name:  elem.Name,
			Type:  elem.Type.(copier).copy(true),
		}
	}
	return cp
}

//
// Shared pool
//

var _sharedOneOfPool = &sharedPool{defaultSize: 10}

func getSharedOneOf() *OneOf {
	v := _sharedOneOfPool.pool.Get()
	st, ok := v.(*OneOf)
	if ok {
		return st
	}
	return &OneOf{
		Elements:     make([]OneOfElement, 0, _sharedOneOfPool.defaultSize),
		needsRelease: true,
	}
}

func putSharedOneOf(x *OneOf) {
	if !x.needsRelease {
		return
	}
	length := cap(x.Elements)
	*x = OneOf{needsRelease: true, Elements: x.Elements[:0]}
	_sharedOneOfPool.put(x, length)
}
