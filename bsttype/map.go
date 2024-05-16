package bsttype

import (
	"io"
	"strings"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
)

// Compile time check to ensure that Map implements the Type interface.
var (
	_ Type         = (*Map)(nil)
	_ TypeReader   = (*Map)(nil)
	_ TypeWriter   = (*Map)(nil)
	_ TypeComparer = (*Map)(nil)
	_ TypeSkipper  = (*Map)(nil)
)

// Compile-time checks for Dependency interfaces.
var (
	_ DependencyOperator = (*Map)(nil)
	_ DependencyChecker  = (*Map)(nil)
	_ DependencyComposer = (*Map)(nil)
	_ DependencyNeeder   = (*Map)(nil)
	_ DependencyVerifier = (*Map)(nil)
	_ DependencyResolver = (*Map)(nil)
)

// Compile-time checks for internal interfaces
var (
	_ copier        = (*Map)(nil)
	_ cycleDetector = (*Map)(nil)
	_ refCounter    = (*Map)(nil)
)

type (
	// Map is the descriptor for a map type.
	// The binary representation of a map is the concatenation of the
	// type of the key and the type of the value.
	// Map binary representation looks like:
	// Size(bits) | Name         | Description
	//   1		  | KeyDesc      | Descending flag for the keys.
	//   2        | -            | Empty byte.
	//   5        | KeyType      | Type of the key.
	//   Variable | KeyContent   | Content of the key - optional (if the key type implements typeIsComplex interface).
	//   1		  | ValueDesc    | Descending flag for the values.
	//   2        | -            | Empty byte.
	//   5        | ValueType    | Type of the value.
	//   Variable | ValueContent | Content of the value - optional (if the value type implements typeIsComplex interface).
	Map struct {
		// Key is the key definition of the map.
		Key MapElement
		// Value is the value definition of the map.
		Value MapElement

		needsRelease bool
	}

	// MapElement is an element type of Map.
	MapElement struct {
		// Type determines a map element type.
		Type Type
		// Descending determines if the element is sorted in descending order.
		Descending bool
	}
)

// MapTypeOf creates a new map type for given key and value.
// A keyDesc determines if the key value is expected to be stored in descending order.
// A valueDesc determines if the value value is expected to be stored in descending order.
func MapTypeOf(key, value Type, keyDesc, valueDesc bool) *Map {
	return &Map{
		Key:   MapElement{Type: key, Descending: keyDesc},
		Value: MapElement{Type: value, Descending: valueDesc},
	}
}

// String returns a human-readable representation of the map value.
func (x MapElement) String() string {
	if x.Descending {
		return "-" + x.Type.String()
	}
	return x.Type.String()
}

// String returns a human-readable representation of the map type.
func (x *Map) String() string {
	var sb strings.Builder
	sb.WriteString("Map[")
	sb.WriteString(x.Key.String())
	sb.WriteRune(']')
	sb.WriteString(x.Value.String())
	return sb.String()
}

// Kind returns the basic kind of the value.
func (*Map) Kind() Kind {
	return KindMap
}

// CompareType returns true if the receiver and the argument have the same type.
// Implements the TypeComparer interface.
func (x *Map) CompareType(to TypeComparer) bool {
	tx, ok := to.(*Map)
	if !ok {
		return false
	}

	return x.Key.Descending == tx.Key.Descending && x.Value.Descending == tx.Value.Descending &&
		TypesEqual(x.Key.Type, tx.Key.Type) && TypesEqual(x.Value.Type, tx.Value.Type)
}

// SkipType skips the type in the byte slice.
func (x *Map) SkipType(rs io.ReadSeeker) (int64, error) {
	// 1. Read the key type.
	bt, err := bstio.ReadByte(rs)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read map key type")
	}
	bytesSkipped := int64(1)

	// 2. Drop the first bit of the key descending flag.
	bt = (bt << 1) >> 1

	// 3.  Get the key type.
	kt := emptyKindType(Kind(bt), false)

	// 4. If the key type implements TypeContent interface, then skip it's content.
	var skipped int64
	ktc, ok := kt.(TypeSkipper)
	if ok {
		skipped, err = ktc.SkipType(rs)
		if err != nil {
			return skipped + 1, err
		}
		bytesSkipped += skipped
	}

	// 5. Read the value type.
	bt, err = bstio.ReadByte(rs)
	if err != nil {
		return bytesSkipped, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read map value type")
	}
	bytesSkipped += 1

	// 6. Drop the first bit of the value descending flag.
	bt = (bt << 1) >> 1

	// 7.  Get the value type.
	vt := emptyKindType(Kind(bt), false)

	// 8. If the value type implements TypeContent interface, then skip it's content.
	vtc, ok := vt.(TypeSkipper)
	if ok {
		skipped, err = vtc.SkipType(rs)
		if err != nil {
			return skipped + 1, err
		}
		bytesSkipped += skipped
	}

	return bytesSkipped, nil
}

// ReadType reads the type from the byte reader.
// Implements the TypeReader interface.
func (x *Map) ReadType(r io.Reader) (int, error) {
	// 1. Read the key type.
	bt, err := bstio.ReadByte(r)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read map key type")
	}
	bytesRead := 1

	// 2. Read the key descending flag.
	x.Key.Descending = (bt & 0x80) != 0

	// 3. Drop the first bit of the key descending flag.
	bt = (bt << 1) >> 1

	// 4.  Get the key type.
	x.Key.Type = emptyKindType(Kind(bt), false)

	// 5. If the key type implements TypeContent interface, then read it's content.
	var read int
	tr, ok := x.Key.Type.(TypeReader)
	if ok {
		read, err = tr.ReadType(r)
		if err != nil {
			return bytesRead, err
		}
		bytesRead += read
	}
	// 6. Read the value type.
	bt, err = bstio.ReadByte(r)
	if err != nil {
		return bytesRead, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read map value type")
	}
	bytesRead += 1

	// 7. Read the value descending flag.
	x.Value.Descending = (bt & 0x80) != 0

	// 8. Drop the first bit of the value descending flag.
	bt = (bt << 1) >> 1

	// 9.  Get the value type.
	x.Value.Type = emptyKindType(Kind(bt), false)

	// 10. If the value type implements TypeContent interface, then read it's content.
	tr, ok = x.Value.Type.(TypeReader)
	if ok {
		read, err = tr.ReadType(r)
		if err != nil {
			return bytesRead, err
		}
		bytesRead += read
	}
	return bytesRead, nil
}

// WriteType writes the type to the writer.
// Implements the TypeWriter interface.
func (x *Map) WriteType(w io.Writer) (int, error) {
	// 1. Prepare the key type from the kind.
	bt := byte(x.Key.Type.Kind())

	// 2. Write the key descending flag.
	if x.Key.Descending {
		bt |= 0x80
	}

	// 3. Write the key type.
	bytesWritten, err := w.Write([]byte{bt})
	if err != nil {
		return bytesWritten, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write map key type")
	}

	// 4. If the key type implements TypeContent interface, then write it's content.
	var written int
	ktc, ok := x.Key.Type.(TypeWriter)
	if ok {
		written, err = ktc.WriteType(w)
		if err != nil {
			return written + 1, err
		}
		bytesWritten += written
	}

	// 5. Write the value type header.
	bt = byte(x.Value.Type.Kind())

	// 6. Write the value descending flag.
	if x.Value.Descending {
		bt |= 0x80
	}

	// 7. Write the value type.
	err = bstio.WriteByte(w, bt)
	if err != nil {
		return bytesWritten, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write map value type")
	}
	bytesWritten++

	// 8. If the value type implements TypeContent interface, then write it's content.
	vtc, ok := x.Value.Type.(TypeWriter)
	if ok {
		written, err = vtc.WriteType(w)
		if err != nil {
			return written + 1, err
		}
		bytesWritten += written
	}
	return bytesWritten, nil
}

//
// Dependencies
//

// CheckDependencies checks for any reference dependencies in its key and value Types.
// Implements DependencyChecker interface.
func (x *Map) CheckDependencies(m *Modules) (CheckDependenciesResult, error) {
	var res CheckDependenciesResult
	// 1. Check the key type.
	dm, ok := x.Key.Type.(DependencyChecker)
	if ok {
		partRes, err := dm.CheckDependencies(m)
		if err != nil {
			return CheckDependenciesResult{}, err
		}
		res.ResolveRequired = partRes.ResolveRequired
		res.ComposeRequired = partRes.ComposeRequired
	}

	// 2. Check the value type.
	dm, ok = x.Value.Type.(DependencyChecker)
	if ok {
		partRes, err := dm.CheckDependencies(m)
		if err != nil {
			return CheckDependenciesResult{}, err
		}
		res.ResolveRequired = res.ResolveRequired || partRes.ResolveRequired
		res.ComposeRequired = res.ComposeRequired || partRes.ComposeRequired
	}
	return res, nil
}

// ComposeDependencies checks for any reference dependencies in its key and value Types.
// Implements DependencyComposer interface.
func (x *Map) ComposeDependencies(m *Modules) error {
	kdm, ok := x.Key.Type.(DependencyComposer)
	if ok {
		if err := kdm.ComposeDependencies(m); err != nil {
			return err
		}
	}

	vdm, ok := x.Value.Type.(DependencyComposer)
	if ok {
		if err := vdm.ComposeDependencies(m); err != nil {
			return err
		}
	}

	return nil
}

// NeedsDependencies checks for any reference dependencies in its key and value Types.
// Implements DependencyNeeder interface.
func (x *Map) NeedsDependencies() bool {
	kdm, ok := x.Key.Type.(DependencyNeeder)
	if ok && kdm.NeedsDependencies() {
		return true
	}

	vdm, ok := x.Value.Type.(DependencyNeeder)
	if ok && vdm.NeedsDependencies() {
		return true
	}

	return false
}

// VerifyDependencies checks for any reference dependencies in its key and value Types.
// Implements DependencyVerifier interface.
func (x *Map) VerifyDependencies() error {
	vdk, ok := x.Key.Type.(DependencyVerifier)
	if ok {
		if err := vdk.VerifyDependencies(); err != nil {
			return err
		}
	}

	vdv, ok := x.Value.Type.(DependencyVerifier)
	if ok {
		if err := vdv.VerifyDependencies(); err != nil {
			return err
		}
	}
	return nil
}

// ResolveDependencies resolves the references in the type.
// Implements the DependencyResolver interface.
func (x *Map) ResolveDependencies(m *Modules) (int64, error) {
	kr, ok := x.Key.Type.(DependencyResolver)
	if ok {
		if _, err := kr.ResolveDependencies(m); err != nil {
			return 0, err
		}
	}

	vr, ok := x.Value.Type.(DependencyResolver)
	if ok {
		if _, err := vr.ResolveDependencies(m); err != nil {
			return 0, err
		}
	}
	return 0, nil
}

func (x *Map) detectCycles(mod, name string) error {
	if kn, ok := x.Key.Type.(*Named); ok {
		if kn.Module == mod && kn.Name == name {
			return bsterr.Err(bsterr.CodeCyclicDependency, "cycle detected in map key type").
				WithDetails(
					bsterr.D("module", mod),
					bsterr.D("name", name),
				)
		}
	}

	if vn, ok := x.Value.Type.(*Named); ok {
		if vn.Module == mod && vn.Name == name {
			return bsterr.Err(bsterr.CodeCyclicDependency, "cycle detected in map value type").
				WithDetails(
					bsterr.D("module", mod),
					bsterr.D("name", name),
				)
		}
	}
	return nil
}

func (x *Map) countRefs() int64 {
	var refs int64
	kr, ok := x.Key.Type.(refCounter)
	if ok {
		refs += kr.countRefs()
	}

	vr, ok := x.Value.Type.(refCounter)
	if ok {
		refs += vr.countRefs()
	}
	return refs
}

func (x *Map) copy(shared bool) Type {
	var cp *Map
	if shared {
		cp = getSharedMap()
	} else {
		cp = new(Map)
	}

	cp.Key = MapElement{
		Type:       cp.Key.Type.(copier).copy(shared),
		Descending: cp.Key.Descending,
	}
	cp.Value = MapElement{
		Type:       cp.Value.Type.(copier).copy(shared),
		Descending: cp.Value.Descending,
	}
	return cp
}

//
// Shared Pool
//

var _sharedMapPool = &sharedPool{}

func getSharedMap() *Map {
	v := _sharedMapPool.pool.Get()
	st, ok := v.(*Map)
	if ok {
		return st
	}
	return &Map{
		needsRelease: true,
	}
}

func putSharedMap(x *Map) {
	if !x.needsRelease {
		return
	}
	*x = Map{needsRelease: true}
	_sharedMapPool.pool.Put(x)
}
