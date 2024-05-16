package bsttype

import (
	"fmt"
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
)

// Compile-time check if Array implements interfaces.
var (
	_ Type         = (*Array)(nil)
	_ TypeComparer = (*Array)(nil)
	_ TypeSkipper  = (*Array)(nil)
	_ TypeReader   = (*Array)(nil)
	_ TypeWriter   = (*Array)(nil)
)

// Compile-time checks for Dependency interfaces.
var (
	_ DependencyOperator = (*Array)(nil)
	_ DependencyChecker  = (*Array)(nil)
	_ DependencyComposer = (*Array)(nil)
	_ DependencyNeeder   = (*Array)(nil)
	_ DependencyVerifier = (*Array)(nil)
	_ DependencyResolver = (*Array)(nil)
)

// Compile-time checks for internal interfaces
var (
	_ copier        = (*Array)(nil)
	_ cycleDetector = (*Array)(nil)
	_ refCounter    = (*Array)(nil)
)

// ArrayOf returns the array type of the given element type.
// If the element type is nil, the function panics.
func ArrayOf(t Type) *Array {
	if t == nil {
		panic("array element type is nil")
	}
	return &Array{Type: t}
}

// ArrayOfShared returns the array type of the given element type.
// If the element type is nil, the function panics.
// The Array type is taken from the shared pool and should be returned after usage.
func ArrayOfShared(t Type) *Array {
	if t == nil {
		panic("array element type is nil")
	}
	a := getSharedArray()
	a.Type = t
	return a
}

// FixedSizeArrayOf returns the array type of the given element type
// with the given fixed size.
// If the element type is nil, the function panics.
func FixedSizeArrayOf(t Type, fixedSize uint) *Array {
	if t == nil {
		panic("array element type is nil")
	}
	return &Array{
		Type:      t,
		FixedSize: fixedSize,
	}
}

// FixedSizeSharedArrayOf returns the array type of the given element type
// with the given fixed size.
// If the element type is nil, the function panics.
// The Array type is taken from the shared pool and should be returned after usage.
func FixedSizeSharedArrayOf(t Type, fixedSize uint) *Array {
	if t == nil {
		panic("array element type is nil")
	}
	a := getSharedArray()
	a.Type = t
	a.FixedSize = fixedSize
	return a
}

// Array is a descriptor of the array type.
// The array type binary is composed as follows:
//   - The first byte is the type header which is in fact the Kind of the array type.
//   - If the base type of the array is a complex, it is followed by its content.
//   - The next byte is the array size header.
//     If the array has fixed size, the most significant bit is set to 1 and the remaining 7 bits
//     are used to encode the binary size of the fixed size integer.
//   - If the array has fixed size, after the array size header, the fixed size integer is encoded.
type Array struct {
	Type      Type
	FixedSize uint
	isShared  bool
}

// String returns the string representation of the type.
func (x *Array) String() string {
	if x.Type == nil {
		return "UndefinedArray"
	}
	if x.HasFixedSize() {
		return fmt.Sprintf("Array[%d](%s)", x.FixedSize, x.Type.String())
	}
	return fmt.Sprintf("Array(%s)", x.Type.String())
}

// Kind returns the kind of the value.
func (*Array) Kind() Kind {
	return KindArray
}

// Elem dereferences the array type wrapper and returns the wrapped type.
func (x *Array) Elem() Type {
	return x.Type
}

// HasFixedSize returns true if the array has fixed size.
func (x *Array) HasFixedSize() bool {
	return x.FixedSize > 0
}

// CompareType returns true if the two types are equal.
// Implements the TypeComparer interface.
func (x *Array) CompareType(to TypeComparer) bool {
	ot, ok := to.(*Array)
	if !ok {
		return false
	}

	if x.FixedSize != ot.FixedSize {
		return false
	}
	return TypesEqual(x.Type, ot.Type)
}

// SkipType skips the type of the value.
// Implements the TypeSkipper interface.
func (x *Array) SkipType(rs io.ReadSeeker) (int64, error) {
	// 1. Skip the type.
	bytesSkipped, err := SkipType(rs)
	if err != nil {
		return bytesSkipped, err
	}

	// 2. Read the array size header byte.
	bt, err := bstio.ReadByte(rs)
	if err != nil {
		return bytesSkipped, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryType, "failed to read array size header")
	}
	bytesSkipped++

	// 3. Check if the array has fixed size.
	// If the array has fixed size, the binary size is encoded in the header byte.
	if bt == 0 {
		return bytesSkipped, nil
	}

	// 4. Check if the header most significant bit is set to 1.
	if bt>>7 != 1 {
		return bytesSkipped, bsterr.Err(bsterr.CodeDecodingBinaryType, "invalid array size header").
			WithDetails(bsterr.D("header", bt))
	}

	// 5. Clear the most significant bit and check how many bytes are used to encode the array size.
	var toSkip int64
	switch (bt << 1) >> 1 {
	case bstio.BinarySizeZero:
		// Rarely used to have a fixed size of 0.
		return 1, nil
	case bstio.BinarySizeUint8:
		toSkip = 1
	case bstio.BinarySizeUint16:
		toSkip = 2
	case bstio.BinarySizeUint32:
		toSkip = 4
	case bstio.BinarySizeUint64:
		toSkip = 8
	default:
		return 1, bsterr.Errf(bsterr.CodeDecodingBinaryType, "invalid array type header byte").
			WithDetails(
				bsterr.D("detail", "header byte should contain binary size of the fixed size integer"),
				bsterr.D("header", bt),
			)
	}

	// 6. Skip the fixed size integer bytes.
	n, err := rs.Seek(toSkip, io.SeekCurrent)
	if err != nil {
		return bytesSkipped + n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip fixed size integer")
	}
	bytesSkipped += n

	return bytesSkipped, nil
}

// ReadType reads the type from the reader.
// Implements the TypeReader interface.
func (x *Array) ReadType(r io.Reader) (int, error) {
	// 1. Read the type.
	tp, bytesRead, err := ReadType(r, x.isShared)
	if err != nil {
		return bytesRead, err
	}
	x.Type = tp

	// 5. Read the array size header byte.
	bt, err := bstio.ReadByte(r)
	if err != nil {
		return bytesRead, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryType, "failed to read array size header")
	}
	bytesRead++

	// 6. Check if the array has fixed size.
	// If the array has fixed size, the binary size is encoded in the header byte.
	if bt == 0 {
		x.FixedSize = 0
		return bytesRead, nil
	}

	// 7. Check if the header most significant bit is set to 1.
	if bt>>7 != 1 {
		return bytesRead, bsterr.Err(bsterr.CodeDecodingBinaryValue, "invalid array size header").
			WithDetails(bsterr.D("header", bt))
	}

	// 8. Clear the most significant bit and ch
	// eck how many bytes are used to encode the array size.
	bt = (bt << 1) >> 1

	fixedSize, n, err := bstio.ReadUintValue(r, bt, false)
	if err != nil {
		return bytesRead + n, err
	}
	x.FixedSize = fixedSize

	return bytesRead + n, nil
}

// WriteType writes the type to the writer.
// Implements the TypeWriter interface.
func (x *Array) WriteType(w io.Writer) (int, error) {
	// 1. Write the base type header.
	bytesWritten, err := WriteType(w, x.Type)
	if err != nil {
		return bytesWritten, err
	}

	// 2. Write the array size header.
	if !x.HasFixedSize() {
		if err = bstio.WriteByte(w, 0x00); err != nil {
			return bytesWritten, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write array size header")
		}
		bytesWritten += 1
		return bytesWritten, nil
	}
	// 3. Write the array size header for fixed size array.
	//    10000000 (0x80) - most significant bit is set to 1.
	bth := byte(0x80)

	size := bstio.UintSizeHeader(x.FixedSize, false)
	bth |= size
	if err = bstio.WriteByte(w, bth); err != nil {
		return bytesWritten, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write array size header")
	}
	bytesWritten++

	// 4. Encode the array size.
	n, err := bstio.WriteUintValue(w, x.FixedSize, size, false)
	if err != nil {
		return bytesWritten + n, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write array size header")
	}

	return bytesWritten + n, nil
}

// CheckDependencies checks if the dependencies are valid.
// Implements the DependencyChecker interface.
func (x *Array) CheckDependencies(m *Modules) (CheckDependenciesResult, error) {
	dm, ok := x.Type.(DependencyChecker)
	if !ok {
		return CheckDependenciesResult{}, nil
	}
	return dm.CheckDependencies(m)
}

// ComposeDependencies if necessary composes named references in the Modules.
// Implements the DependencyComposer interface.
func (x *Array) ComposeDependencies(m *Modules) error {
	dm, ok := x.Type.(DependencyComposer)
	if !ok {
		return nil
	}
	return dm.ComposeDependencies(m)
}

// VerifyDependencies verifies if the dependencies are valid.
// Implements the DependencyVerifier interface.
func (x *Array) VerifyDependencies() error {
	dm, ok := x.Type.(DependencyVerifier)
	if !ok {
		return nil
	}
	return dm.VerifyDependencies()
}

// NeedsDependencies returns whether the type needs dependencies.
// Implements the DependencyNeeder interface.
func (x *Array) NeedsDependencies() bool {
	dm, ok := x.Type.(DependencyNeeder)
	if !ok {
		return false
	}
	return dm.NeedsDependencies()
}

// ResolveDependencies allows to resolve references defined by the Array.
// An array cannot reference the same named array one after another.
func (x *Array) ResolveDependencies(m *Modules) (int64, error) {
	mr, ok := x.Type.(DependencyResolver)
	if !ok {
		return 0, nil
	}
	return mr.ResolveDependencies(m)
}

func (x *Array) detectCycles(mod, name string) error {
	nt, ok := x.Type.(*Named)
	if !ok {
		return nil
	}

	if nt.Module == mod && nt.Name == name {
		return bsterr.Err(bsterr.CodeCyclicDependency, "cyclic dependency detected").
			WithDetails(bsterr.D("module", mod), bsterr.D("name", name))
	}
	return nil
}

func (x *Array) countRefs() int64 {
	rc, ok := x.Type.(refCounter)
	if !ok {
		return 0
	}
	return rc.countRefs()
}

func (x *Array) copy(shared bool) Type {
	var a *Array
	if shared {
		a = getSharedArray()
	} else {
		a = &Array{}
	}
	*a = *x
	a.Type = x.Type.(copier).copy(shared)
	return a
}

//
// Shared Pool
//

var _sharedArrayPool = &sharedPool{defaultSize: 10}

func getSharedArray() *Array {
	v := _sharedArrayPool.pool.Get()
	st, ok := v.(*Array)
	if ok {
		return st
	}
	return &Array{}
}

func putSharedArray(x *Array) {
	if !x.isShared {
		return
	}
	*x = Array{}
	_sharedArrayPool.pool.Put(x)
}
