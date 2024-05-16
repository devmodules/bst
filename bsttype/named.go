package bsttype

import (
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
)

// Compile-time checks for the Type interfaces.
var (
	_ Type         = (*Named)(nil)
	_ TypeReader   = (*Named)(nil)
	_ TypeWriter   = (*Named)(nil)
	_ TypeSkipper  = (*Named)(nil)
	_ TypeComparer = (*Named)(nil)
)

// Compile-time checks for Dependency interfaces.
var (
	_ DependencyOperator = (*Named)(nil)
	_ DependencyChecker  = (*Named)(nil)
	_ DependencyComposer = (*Named)(nil)
	_ DependencyNeeder   = (*Named)(nil)
	_ DependencyVerifier = (*Named)(nil)
	_ DependencyResolver = (*Named)(nil)
)

// Compile-time checks for internal interfaces
var (
	_ copier     = (*Named)(nil)
	_ refCounter = (*Named)(nil)
)

// Named is the implementation of the Type that wraps some repeatable Type
// and assigns a name in specific module.
// It needs to be resolved out of the well known modules during the decoding.
// The binary format of the type does not compose wrapped type.
type Named struct {
	// Module defines a module to which a named type is assigned.
	Module string
	// Name defines a unique name for given module.
	Name string
	// Type determines a wrapped type for the named one.
	// It is only available after being resolved out of the module.
	Type Type

	resolved, needsRelease bool
}

// Kind returns the kind of the type.
func (x *Named) Kind() Kind {
	return KindNamed
}

// String returns the string representation of the type.
func (x *Named) String() string {
	return x.Module + "." + x.Name
}

// Compile-time checks if Named implements TypeSkipper interface.
var _ TypeSkipper = (*Named)(nil)

// SkipType skips the binary data that represents the type.
// It returns the number of bytes skipped and an error if any.
// Implements the Skipper interface.
func (x *Named) SkipType(br io.ReadSeeker) (int64, error) {
	// 1. Skip the module name.
	n, err := bstio.SkipNonComparableString(br, false)
	if err != nil {
		return n, err
	}
	bytesSkipped := n

	// 2. Skip the name.
	n, err = bstio.SkipNonComparableString(br, false)
	if err != nil {
		return n, err
	}
	bytesSkipped += n

	return bytesSkipped, nil
}

// ReadType reads the binary data that represents the type.
// The Named binary does not contain the type itself.
// It needs to be resolved out of the module definition.
// Implements the Reader interface.
func (x *Named) ReadType(r io.Reader) (bytesRead int, err error) {
	// 1. Read the module name.
	var n int
	x.Module, n, err = bstio.ReadString(r, false, false)
	if err != nil {
		return n, err
	}
	bytesRead = n

	// 2. Read the name.
	x.Name, n, err = bstio.ReadString(r, false, false)
	if err != nil {
		return bytesRead + n, err
	}
	bytesRead += n

	return bytesRead, nil
}

// WriteType writes the binary data that represents the type.
// It returns the number of bytes written and an error if any.
// The Named binary does not contain the type itself, thus
// no type is written.
// Implements the Writer interface.
func (x *Named) WriteType(w io.Writer) (int, error) {
	// 1. Write the module name.
	n, err := bstio.WriteString(w, x.Module, false, false)
	if err != nil {
		return n, err
	}
	bytesWritten := n

	// 2. Write the name.
	n, err = bstio.WriteString(w, x.Name, false, false)
	if err != nil {
		return bytesWritten + n, err
	}

	return bytesWritten + n, nil
}

// CompareType checks if the input type 'tc' is equal to current type.
// Implements the TypeComparer interface.
func (x *Named) CompareType(to TypeComparer) bool {
	named, ok := to.(*Named)
	if !ok {
		return false
	}
	return x.Module == named.Module && x.Name == named.Name
}

//
// Dependencies
//

// CheckDependencies checks the dependencies of the named type.
// It verifies if the modules contains that given named type within its registry.
// Implements DependencyChecker interface.
func (x *Named) CheckDependencies(m *Modules) (res CheckDependenciesResult, err error) {
	// 1. Check if the Named is resolved and the type is defined.
	if x.Type != nil {
		res.ComposeRequired = m == nil
		if m != nil {
			res.ResolveRequired = !m.existsNamedType(x.Module, x.Name)
		}
		return res, nil
	}

	// 2. Check if the modules are defined.
	if m == nil {
		return CheckDependenciesResult{}, bsterr.Err(bsterr.CodeTypeNotMapped, "named type not resolved and no modules defined").
			WithDetails(
				bsterr.D("name", x.Name),
				bsterr.D("module", x.Module),
			)
	}

	// 3. Check if the type definition is defined in the modules.
	if !m.existsNamedType(x.Module, x.Name) {
		return CheckDependenciesResult{}, bsterr.Err(bsterr.CodeTypeNotMapped, "named type undefined and not defined in the modules").
			WithDetails(
				bsterr.D("name", x.Name),
				bsterr.D("module", x.Module),
			)
	}

	// 4. The type is undefined but exists in the modules. It needs to be resolved.
	res.ResolveRequired = true
	return res, nil
}

// ComposeDependencies composes the dependencies within provided modules.
// Implements DependencyComposer interface.
func (x *Named) ComposeDependencies(m *Modules) error {
	return m.addNamedType(x, true)
}

// VerifyDependencies verifies the dependencies of the named type.
// Implements DependencyVerifier interface.
func (x *Named) VerifyDependencies() error {
	if x.Name == "" {
		return bsterr.Err(bsterr.CodeTypeNotMapped, "named type name is undefined").
			WithDetails(
				bsterr.D("module", x.Module),
			)
	}

	if x.Type == nil {
		return bsterr.Err(bsterr.CodeTypeNotMapped, "named type not defined").
			WithDetails(
				bsterr.D("name", x.Name),
				bsterr.D("module", x.Module),
			)
	}

	return nil
}

// NeedsDependencies returns always true for the named type.
// Implements DependencyNeeder interface.
func (x *Named) NeedsDependencies() bool {
	return true
}

// ResolveDependencies resolves the references of the named type.
// It tries to add the named type to the modules if it still not exists there.
func (x *Named) ResolveDependencies(m *Modules) (int64, error) {
	if x.resolved {
		return 1, nil
	}
	def, err := m.findNamedTypeDefinition(x.Module, x.Name)
	if err != nil {
		return 0, err
	}
	x.Type = def
	x.resolved = true
	return 1, nil
}

// countRefs counts the references of the named type.
func (x *Named) countRefs() int64 {
	return 1
}

func (x *Named) copy(shared bool) Type {
	var cp *Named
	if shared {
		cp = getSharedNamed()
	} else {
		cp = new(Named)
	}
	cp.Name = x.Name
	cp.Module = x.Module
	cp.resolved = x.resolved

	if x.Type != nil {
		cp.Type = x.Type.(copier).copy(shared)
	}
	return cp
}

//
// Shared Pool
//

var _sharedNamedPool = &sharedPool{}

func getSharedNamed() *Named {
	v := _sharedNamedPool.pool.Get()
	st, ok := v.(*Named)
	if ok {
		return st
	}
	return &Named{needsRelease: true}
}

func putSharedNamed(x *Named) {
	if !x.needsRelease {
		return
	}
	*x = Named{needsRelease: true}
	_sharedNamedPool.pool.Put(x)
}
