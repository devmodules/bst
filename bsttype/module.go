package bsttype

import (
	"io"
	"sync"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
)

// Modules is a named list of Modules.
type Modules struct {
	List []*Module

	sharedDefs bool

	resolved bool
	checkSum int64
}

// Read decodes and reads binary encoded modules.
// SharedDefs option is used to enable/disable shared definitions.
// When shared definitions are enabled, Struct named references reuses the same definition.
// Thus, no allocation are required during decoding.
// However, once finished, the modules should be freed, which releases the memory used by the definitions.
// After being freed, Modules types should not be used anymore.
func (x *Modules) Read(r io.Reader, sharedDefs bool) (int, error) {
	// 1. Read the number of modules.
	ml, n, err := bstio.ReadUint(r, false)
	if err != nil {
		return n, err
	}
	bytesRead := n

	// 2. Check if the number of modules is greater than the allocated size of the Modules.
	if int(ml) > cap(x.List) {
		// 2.1. Allocate a new list of modules.
		x.List = make([]*Module, ml)
	} else {
		// 2.2. Reset number of modules up to needed size.
		x.List = x.List[:ml]
	}

	// 3. Read all the modules one by one.
	for i := uint(0); i < ml; i++ {
		var mod *Module
		if sharedDefs {
			mod = GetSharedModule()
		} else {
			mod = &Module{}
		}

		n, err = mod.Read(r)
		if err != nil {
			return bytesRead, err
		}
		bytesRead += n

		x.List[i] = mod
	}

	return bytesRead, nil
}

// Write encodes and writes binary encoded modules.
func (x *Modules) Write(w io.Writer) (int, error) {
	// 1. Write the number of modules.
	n, err := bstio.WriteUint(w, uint(len(x.List)), false)
	if err != nil {
		return n, err
	}
	bytesWritten := n

	// 2. Write all the modules one by one.
	for _, mod := range x.List {
		n, err = mod.Write(w)
		if err != nil {
			return bytesWritten, err
		}
		bytesWritten += n
	}

	return bytesWritten, nil
}

// Resolve all named references in each module.
// Returns an error if a reference cannot be resolved.
func (x *Modules) Resolve() error {
	// 1. Find duplicate definitions.
	if err := x.findDuplicates(); err != nil {
		return err
	}

	// 2. Find cycles.
	if err := x.DetectCycles(); err != nil {
		return err
	}

	// 2. Resolve references.
	if err := x.resolveReferences(); err != nil {
		return err
	}

	x.resolved = true
	return nil
}

// Free is used to free shared Modules, and all its elements.
// Be careful that this function should be called only once, after all Modules are finished, as it
// releases all the types and definitions used by this module.
func (x *Modules) Free() {
	for _, mod := range x.List {
		for _, def := range mod.Definitions {
			PutSharedType(def.Type)
		}
		if mod.sharedDefs {
			PutSharedModule(mod)
		}
	}

	if x.sharedDefs {
		PutSharedModules(x)
	}
}

func (x *Modules) findDuplicates() error {
	// 1. Iterate over all modules.
	for _, mod := range x.List {
		// 2. Iterate over all definitions.
		for j, def := range mod.Definitions {
			// 3. Try to find a matching definition.
			//    If a matching definition is found, it is an error.
			for k, secDef := range mod.Definitions {
				if j != k && secDef.Name == def.Name {
					return bsterr.Err(bsterr.CodeTypeAlreadyMapped, "type already mapped").
						WithDetails(
							bsterr.D("module", mod.Name),
							bsterr.D("name", def.Name),
						)
				}
			}
		}
	}
	return nil
}

// findNamedTypeDefinition finds a named type in the Modules.
func (x *Modules) findNamedTypeDefinition(module, name string) (Type, error) {
	// 1. Search all modules for matching named type.
	for _, mod := range x.List {
		if mod.Name == module {
			// 2. Find matching definition.
			for _, def := range mod.Definitions {
				if def.Name == name {
					return def.Type, nil
				}
			}
			// 3. If definition is not found, return an error - no type found within given module
			break
		}
	}
	return nil, bsterr.Errf(bsterr.CodeTypeNotMapped, "type %s.%s not found", module, name)
}

func (x *Modules) existsNamedType(module, name string) bool {
	for _, mod := range x.List {
		if mod.Name != module {
			continue
		}

		for _, def := range mod.Definitions {
			if def.Name == name {
				return true
			}
		}
		return false
	}
	return false
}

func (x *Modules) addNamedType(nt *Named, ifNotExists bool) error {
	// 1. Find the module.
	for i, mod := range x.List {
		if mod.Name == nt.Module {
			// 2. Find the definition.
			for _, def := range mod.Definitions {
				if def.Name == nt.Name {
					// 3. If the not exists option is enabled, return nil.
					if ifNotExists {
						return nil
					}
					// 3.2. Otherwise, return an error.
					return bsterr.Errf(bsterr.CodeTypeAlreadyMapped, "type %s.%s already mapped", nt.Module, nt.Name)
				}
			}
			// 4. Add the definition.
			x.List[i].Definitions = append(x.List[i].Definitions, ModuleDefinition{Name: nt.Name, Type: nt.Type})
			return nil
		}
	}
	// 5. Add the module.
	var mod *Module
	if x.sharedDefs {
		mod = GetSharedModule()
	} else {
		mod = &Module{}
	}
	mod.Name = nt.Module
	mod.Definitions = append(mod.Definitions, ModuleDefinition{Name: nt.Name, Type: nt.Type})
	x.List = append(x.List, mod)
	return nil
}

func (x *Modules) resolveReferences() error {
	x.resolved = false
	x.checkSum = 0
	// 1. Iterate over all modules and resolve all definition references.
	for _, mod := range x.List {
		// 2. Reset the checksum of the module.
		mod.checkSum = 0
		// 3. Replace all references with the shared types.
		for _, def := range mod.Definitions {
			// 4. A type could only be resolved if it contains or wraps other types.
			//    There is a special interface type which is used for the types to resolve its internal references.
			rr, ok := def.Type.(DependencyResolver)
			if !ok {
				continue
			}
			n, err := rr.ResolveDependencies(x)
			if err != nil {
				return err
			}

			// 5. Update the checksum of the module.
			mod.checkSum += n
		}
		// 6. Update modules checksum by this module checksum.
		x.checkSum += mod.checkSum
	}
	// 7. Mark the modules as resolved.
	x.resolved = true
	return nil
}

// IsResolved returns true if the Modules are resolved.
func (x *Modules) IsResolved() bool {
	// 1. Quickly check if the modules where resolved.
	if !x.resolved {
		return false
	}

	// 2. Get sure if the modules references did not change.
	//    NOTE: this is not a 100% accurate check, but it is fast and a good enough check.
	var checkSum int64
	for _, mod := range x.List {
		for _, def := range mod.Definitions {
			rc, ok := def.Type.(refCounter)
			if !ok {
				continue
			}
			checkSum += rc.countRefs()
		}
	}
	return checkSum == x.checkSum
}

// Merge merges input module 'm' into the module 'x'.
// All the definitions of 'm' are added to the module 'x'.
func (x *Modules) Merge(m *Modules) {
	for _, modEx := range m.List {
		for _, mod := range x.List {
			if mod.Name != modEx.Name {
				continue
			}

			for _, defEx := range modEx.Definitions {
				var found bool
				for _, def := range mod.Definitions {
					if def.Name == defEx.Name {
						found = true
						break
					}
				}

				if !found {
					cp := defEx.Type.(copier)
					mod.Definitions = append(mod.Definitions, ModuleDefinition{Name: defEx.Name, Type: cp.copy(x.sharedDefs)})
				}
			}
		}
	}
}

// Add adds a new module to the Modules list.
// If the module already exists, it tries to merge its definitions.
// If the definition already exists, it returns an error.
func (x *Modules) Add(m *Module) error {
	for _, mod := range x.List {
		if mod.Name == m.Name {
			for _, def := range m.Definitions {
				for _, defEx := range mod.Definitions {
					if def.Name == defEx.Name {
						return bsterr.Errf(bsterr.CodeTypeAlreadyMapped, "type %s.%s already mapped", m.Name, def.Name)
					}
				}
			}
			mod.Definitions = append(mod.Definitions, m.Definitions...)
			return nil
		}
	}
	x.List = append(x.List, m)
	return nil
}

// DetectCycles is a function that detects whether named definitions points in non-nullable
// directly to itself.
func (x *Modules) DetectCycles() error {
	for _, m := range x.List {
		for _, def := range m.Definitions {
			cd, ok := def.Type.(cycleDetector)
			if !ok {
				continue
			}
			if err := cd.detectCycles(m.Name, def.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

// Module is a BST package that contains multiple type definitions.
type Module struct {
	// Name is the name of the module.
	Name string
	// Definitions is a list of named module type definitions.
	Definitions []ModuleDefinition

	// sharedDefs is a flag that uses shared structure pointers for named definitions.
	// This way reading and writing could
	sharedDefs, resolved bool
	checkSum             int64
}

// Read the module from the input bytes reader.
func (x *Module) Read(r io.Reader) (int, error) {
	// 1. Read the name of the module.
	name, n, err := bstio.ReadStringNonComparable(r, false)
	if err != nil {
		return n, err
	}
	bytesRead := n

	x.Name = name

	// 2. Read the number of definitions.
	var numDefs uint
	numDefs, n, err = bstio.ReadUint(r, false)
	if err != nil {
		return bytesRead, err
	}
	bytesRead += n

	if int(numDefs) <= cap(x.Definitions) {
		x.Definitions = x.Definitions[:numDefs]
	} else {
		x.Definitions = make([]ModuleDefinition, numDefs)
	}

	// 3. Read the definitions.
	for i := uint(0); i < numDefs; i++ {
		// 3.1. Read the name of the definition.
		name, n, err = bstio.ReadStringNonComparable(r, false)
		if err != nil {
			return bytesRead, err
		}
		bytesRead += n

		// 3.2. Read the type of the definition.
		var t Type
		t, n, err = ReadType(r, x.sharedDefs)
		if err != nil {
			return bytesRead, err
		}
		bytesRead += n

		// 3.3. Add the definition.
		x.Definitions[i] = ModuleDefinition{Name: name, Type: t}
	}
	return bytesRead, nil
}

// Write the module binary content into the writer.
func (x Module) Write(w io.Writer) (int, error) {
	// 1. Write the name of the module.
	n, err := bstio.WriteStringNonComparable(w, x.Name, false)
	if err != nil {
		return n, err
	}
	bytesWritten := n

	// 2. Write the number of definitions.
	n, err = bstio.WriteUint(w, uint(len(x.Definitions)), false)
	if err != nil {
		return bytesWritten, err
	}
	bytesWritten += n

	// 3. Write the definitions.
	for _, def := range x.Definitions {
		// 3.1. Write the name of the definition.
		n, err = bstio.WriteStringNonComparable(w, def.Name, false)
		if err != nil {
			return bytesWritten, err
		}
		bytesWritten += n

		// 3.2. Write the type of the definition.
		n, err = WriteType(w, def.Type)
		if err != nil {
			return bytesWritten, err
		}
		bytesWritten += n
	}
	return bytesWritten, nil
}

// ModuleDefinition is a reference of the named type it needs to be resolved from the input modules.
// It is a temporary type used for the named type references while decoding the module types.
// NOTE: If the named type defines a structure, it uses a Struct pointer.
//
//	This implies, that if the user provided module definition with non pointer
//	named struct types, these definitions would have its definition replaced by the pointers.
type ModuleDefinition struct {
	// Name is the unique name of the named type definition.
	Name string
	// Type is the definition of the named type.
	Type Type
}

// DependencyResolver is an interface that is used to resolve references.
// The interface is used to resolve references in the type definitions.
type DependencyResolver interface {
	ResolveDependencies(m *Modules) (int64, error)
}

// refCounter is an interface that is used to count references.
type refCounter interface {
	countRefs() int64
}

// DependencyOperator is an interface that combines three dependency operations:
// - ComposeDependencies
// - NeedsDependencies
// - VerifyDependencies
type DependencyOperator interface {
	DependencyComposer
	DependencyNeeder
	DependencyVerifier
}

// DependencyNeedVerifier is an interface that combines two dependency operations:
// - NeedsDependencies
// - VerifyDependencies
type DependencyNeedVerifier interface {
	DependencyNeeder
	DependencyVerifier
}

// DependencyComposer is an interface that is used to compose dependencies.
type DependencyComposer interface {
	ComposeDependencies(m *Modules) error
}

// DependencyChecker is an interface that is used to check dependencies.
type DependencyChecker interface {
	CheckDependencies(m *Modules) (CheckDependenciesResult, error)
}

// CheckDependenciesResult is a result of the CheckDependencies operation.
type CheckDependenciesResult struct {
	ResolveRequired bool
	ComposeRequired bool
}

// DependencyNeeder is an interface that is used to determine if the 'modules' dependency are necessary to compose.
type DependencyNeeder interface {
	NeedsDependencies() bool
}

// DependencyVerifier is an interface that is used to verify if dependencies are well-defined.
type DependencyVerifier interface {
	VerifyDependencies() error
}

// cycleDetector is an interface that is used to detect cycles for given type.
type cycleDetector interface {
	detectCycles(mod, name string) error
}

const defaultModulesSize = 10

type sharedPool struct {
	pool        sync.Pool
	defaultSize int
}

func (s *sharedPool) put(v interface{}, curCap int) {
	// TODO: calibrate the size of the default size of the pool by math operations on the curCap.
	s.pool.Put(v)
}

var _modulesPool = &sharedPool{defaultSize: 10}

// GetSharedModules gets the shared Modules from the pool.
// NOTE: The caller is responsible for calling PutSharedModules when the Modules is no longer used.
func GetSharedModules() *Modules {
	v, ok := _modulesPool.pool.Get().(*Modules)
	if ok {
		return v
	}

	return &Modules{
		List:       make([]*Module, 0, _modulesPool.defaultSize),
		sharedDefs: true,
	}
}

// PutSharedModules puts the Modules back to the pool.
// NOTE: After calling this function, the Modules is no longer usable.
func PutSharedModules(m *Modules) {
	cp := cap(m.List)
	m.List = m.List[:0]

	_modulesPool.put(m, cp)
}

var _modulePool = &sharedPool{defaultSize: 10}

// GetSharedModule gets the shared module from the pool.
// NOTE: The caller is responsible for calling PutSharedModule when the Module is no longer used.
func GetSharedModule() *Module {
	v, ok := _modulePool.pool.Get().(*Module)
	if ok {
		return v
	}
	return &Module{
		Definitions: make([]ModuleDefinition, 0, _modulePool.defaultSize),
		sharedDefs:  true,
	}
}

// PutSharedModule puts the module back to the pool.
// NOTE: After calling this function, the Module is no longer usable.
func PutSharedModule(m *Module) {
	cp := len(m.Definitions)
	*m = Module{
		Definitions: m.Definitions[:0],
		sharedDefs:  true,
	}
	_modulePool.put(m, cp)
}
