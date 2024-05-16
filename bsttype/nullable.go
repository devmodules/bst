package bsttype

import (
	"fmt"
	"io"

	"github.com/devmodules/bst/bsterr"
)

// Compile-time check that Nullable implements Type interface.
var (
	_ Type         = (*Nullable)(nil)
	_ TypeReader   = (*Nullable)(nil)
	_ TypeWriter   = (*Nullable)(nil)
	_ TypeSkipper  = (*Nullable)(nil)
	_ TypeComparer = (*Nullable)(nil)
)

// Compile-time check if Nullable implements Dependency interfaces.
var (
	_ DependencyOperator = (*Nullable)(nil)
	_ DependencyChecker  = (*Nullable)(nil)
	_ DependencyComposer = (*Nullable)(nil)
	_ DependencyNeeder   = (*Nullable)(nil)
	_ DependencyVerifier = (*Nullable)(nil)
	_ DependencyResolver = (*Nullable)(nil)
)

// Compile-time check for internal interfaces.
var (
	_ copier        = (*Nullable)(nil)
	_ cycleDetector = (*Nullable)(nil)
	_ refCounter    = (*Nullable)(nil)
)

// Nullable is the type that wraps another type and provides a Null value functionality.
// The binary representation of the nullable type is determined by the wrapped type itself.
type Nullable struct {
	Type Type

	needsRelease bool
}

// NullableOf returns the nullable type that wraps input type.
func NullableOf(t Type) *Nullable {
	if t == nil {
		t = &Basic{}
	}
	return &Nullable{Type: t}
}

// NullableOfShared returns the nullable type that wraps input type.
// The returned type is taken out from the shared pool and should be put back to the pool after use.
func NullableOfShared(t Type) *Nullable {
	n := getSharedNullable()
	n.Type = t
	return n
}

// Type returns the type of the value.
func (x *Nullable) String() string {
	return fmt.Sprintf("Nullable(%s)", x.Type)
}

// Kind returns the kind of the value.
// Implements the Type interface.
func (*Nullable) Kind() Kind {
	return KindNullable
}

// Elem dereferences the pointer wrapped Type.
func (x *Nullable) Elem() Type {
	return x.Type
}

// CompareType returns true if the types are equal.
func (x *Nullable) CompareType(to TypeComparer) bool {
	nt, ok := to.(*Nullable)
	if !ok {
		return false
	}

	return TypesEqual(nt.Type, x.Type)
}

// SkipType skips the content of the embedded nullable type.
func (x *Nullable) SkipType(rs io.ReadSeeker) (int64, error) {
	return SkipType(rs)
}

// ReadType reads the type from the byte reader.
// Implements the TypeReader interface.
func (x *Nullable) ReadType(r io.Reader) (int, error) {
	tp, n, err := ReadType(r, false)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read nullable type")
	}
	x.Type = tp
	return n, nil
}

// WriteType writes the type to the byte writer.
// Implements the TypeWriter interface.
func (x *Nullable) WriteType(w io.Writer) (int, error) {
	return WriteType(w, x.Type)
}

//
// Dependencies
//

// CheckDependencies checks if the type that is Nullable has any Named references
func (x *Nullable) CheckDependencies(m *Modules) (CheckDependenciesResult, error) {
	dt, ok := x.Type.(DependencyChecker)
	if !ok {
		return CheckDependenciesResult{}, nil
	}
	return dt.CheckDependencies(m)
}

// ComposeDependencies checks if the type that is Nullable could have any Named references
// and composes a Modules out of it.
// Implements DependencyComposer interface.
func (x *Nullable) ComposeDependencies(m *Modules) error {
	dt, ok := x.Type.(DependencyComposer)
	if !ok {
		return nil
	}
	return dt.ComposeDependencies(m)
}

// NeedsDependencies checks if the type that is Nullable needs any Named references
// Implements DependencyNeeder interface.
func (x *Nullable) NeedsDependencies() bool {
	dt, ok := x.Type.(DependencyNeeder)
	if !ok {
		return false
	}
	return dt.NeedsDependencies()
}

// VerifyDependencies verifies if the type that is Nullable has any Named references
// Implements DependencyVerifier interface.
func (x *Nullable) VerifyDependencies() error {
	dt, ok := x.Type.(DependencyVerifier)
	if !ok {
		return nil
	}
	return dt.VerifyDependencies()
}

// ResolveDependencies resolves the references in the nullable type.
// Implements DependencyResolver interface.
func (x *Nullable) ResolveDependencies(m *Modules) (int64, error) {
	mr, ok := x.Type.(DependencyResolver)
	if !ok {
		return 0, nil
	}
	return mr.ResolveDependencies(m)
}

func (x *Nullable) detectCycles(mod, name string) error {
	nt, ok := x.Type.(*Named)
	if !ok {
		return nil
	}
	if nt.Module == mod && nt.Name == name {
		return bsterr.Err(bsterr.CodeCyclicDependency, "cyclic dependency detected").
			WithDetails(
				bsterr.D("module", mod),
				bsterr.D("name", name),
			)
	}
	return nil
}

// countRefs counts the references in the nullable type.
// Implements refCounter interface.
func (x *Nullable) countRefs() int64 {
	cr, ok := x.Type.(refCounter)
	if !ok {
		return 0
	}
	return cr.countRefs()
}

func (x *Nullable) copy(shared bool) Type {
	var cp *Nullable
	if shared {
		cp = getSharedNullable()
	} else {
		cp = new(Nullable)
	}
	cp.Type = x.Type.(copier).copy(shared)
	return cp
}

//
// Shared Pool
//

var _sharedNullablesPool = &sharedPool{}

func getSharedNullable() *Nullable {
	v := _sharedNullablesPool.pool.Get()
	st, ok := v.(*Nullable)
	if ok {
		return st
	}
	return &Nullable{}
}

func putSharedNullable(x *Nullable) {
	if !x.needsRelease {
		return
	}
	*x = Nullable{}
	_sharedNullablesPool.put(x, 0)
}
