package bst

import (
	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// WriteAnyType writes an any type value to the composer.
func (x *Composer) WriteAnyType(v bsttype.Type) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	if x.elemType.Kind() != bsttype.KindAny {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindAny),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. Verify if input Type is defined.
	if v == nil || v.Kind() == bsttype.KindUndefined {
		return bsterr.Err(bsterr.CodeInvalidValue, "undefined type to write")
	}

	// 4. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		x.setFieldBuffer()
	}

	// 5. Check if the type requires modules to be written along with the type.
	var (
		header byte
		m      *bsttype.Modules
	)

	dn, ok := v.(bsttype.DependencyOperator)
	if ok && dn.NeedsDependencies() {
		// 5.1. Verify if the modules were embedded and the modules already contains the dependencies of the input type.
		var alreadyContains bool
		if x.opts.EmbedType && x.modules != nil {
			dc, ok := v.(bsttype.DependencyChecker)
			if ok {
				res, err := dc.CheckDependencies(x.modules)
				if err != nil {
					return err
				}
				if !res.ComposeRequired {
					alreadyContains = true
				}
			}
		}

		if !alreadyContains {
			m = bsttype.GetSharedModules()
			if err := dn.ComposeDependencies(m); err != nil {
				return err
			}
			header |= 1 << 4
		}
	}

	// 6. Write the 'Any' value header.
	err := bstio.WriteByte(x.w, header)
	if err != nil {
		return err
	}
	x.bytesWritten++

	var n int

	// 7. If the modules are required decode it.
	if m != nil {
		n, err = m.Write(x.w)
		if err != nil {
			return err
		}
		x.bytesWritten += n
	}

	// 5. Write the type.
	n, err = bsttype.WriteType(x.w, v)
	if err != nil {
		return bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write type")
	}
	x.bytesWritten += n

	// 6. Set element type to the input type.
	x.elemType = v

	return nil
}

// ReadAnyType reads the type of the 'AnyType' value and dereferences extractor element.
func (x *Extractor) ReadAnyType() (bsttype.Type, error) {
	if x.err != nil {
		return nil, x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return nil, bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	// 2. Verify if current element matches the expected type.
	if x.elemType.Kind() != bsttype.KindAny {
		return nil, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindAny),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}
	// 3. Read the 'Any' value header.
	header, err := bstio.ReadByte(x.r)
	if err != nil {
		return nil, err
	}

	// 4. Check if the type defined in the Any value requires modules.
	var (
		m             *bsttype.Modules
		sharedModules bool
		n             int
	)
	if (header>>4)&0x1 != 0 {
		m = bsttype.GetSharedModules()
		sharedModules = true
		n, err = m.Read(x.r, true)
		if err != nil {
			return nil, err
		}
		x.bytesRead += n
	}

	// If the modules were not defined in the Any value, get the default extractor modules.
	if m == nil {
		m = x.opts.Modules
	}

	// 4. Read the type of the 'AnyType' value.
	t, n, err := bsttype.ReadType(x.r, true)
	if err != nil {
		return nil, err
	}
	x.bytesRead += n

	// 5. Check if the type needs to be resolved out of the modules.
	dr, ok := t.(bsttype.DependencyResolver)
	if ok {
		// 5.1. Resolve the type.
		if _, err = dr.ResolveDependencies(m); err != nil {
			x.err = err
			return nil, err
		}
	}

	// 6. Prepare clear function while finishing the element.
	x.clearElemFn = func() {
		if sharedModules {
			m.Free()
		}
		bsttype.PutSharedType(t)
	}

	// 7. Dereference the 'AnyType' value elem type.
	x.elemType, x.err = x.derefType(t)
	if x.err != nil {
		return nil, x.err
	}

	return t, nil
}
