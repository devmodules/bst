package bst

import (
	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// WriteOneOfByIndex writes oneof header that matches element with an buffIndex.
func (x *Composer) WriteOneOfByIndex(index uint) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	ot, ok := x.elemType.(*bsttype.OneOf)
	if !ok {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindOneOf),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. Find the oneof matching the buffIndex.
	var elem bsttype.Type
	for i := range ot.Elements {
		if ot.Elements[i].Index == index {
			elem = ot.Elements[i].Type
			break
		}
	}

	// 4. Write the oneof buffIndex header.
	if err := x.writeOneOfIndex(index, elem, ot.IndexBytes); err != nil {
		return err
	}
	return nil
}

// WriteOneOfByName writes oneof header that matches element with a name.
func (x *Composer) WriteOneOfByName(name string) error {
	// 1. Check if the element was already written.
	if x.done {
		return bsterr.Err(bsterr.CodeAlreadyWritten, "element already written")
	}

	// 2. Verify if current element matches expected type.
	ot, ok := x.elemType.(*bsttype.OneOf)
	if !ok {
		return bsterr.Err(bsterr.CodeInvalidType, "invalid type to write").
			WithDetails(
				bsterr.D("expected", bsttype.KindOneOf),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. Find the oneof matching the name.
	var (
		elem  bsttype.Type
		index uint
	)
	for i := range ot.Elements {
		if ot.Elements[i].Name == name {
			elem = ot.Elements[i].Type
			index = ot.Elements[i].Index
			break
		}
	}

	// 4. Write the oneof buffIndex header.
	if err := x.writeOneOfIndex(index, elem, ot.IndexBytes); err != nil {
		return err
	}

	return nil
}

func (x *Composer) writeOneOfIndex(index uint, elem bsttype.Type, indexBytes uint8) error {
	// 1. Return an error if the buffIndex is not found.
	if elem == nil {
		return bsterr.Err(bsterr.CodeInvalidValue, "invalid oneof value").
			WithDetails(
				bsterr.D("value", index),
			)
	}

	// 2. If the base is a struct, check if the field header needs to be written.
	if x.needWriteFieldHeader() {
		x.setFieldBuffer()
	}

	// 3. Marshal oneof buffIndex.
	n, err := bstio.WriteOneOfIndex(x.w, index, indexBytes, x.elemDesc)
	if err != nil {
		return err
	}

	x.bytesWritten += n

	// 4. Dereference the oneof element.
	x.elemType = elem
	return nil
}

func (x *Composer) reset() {
	*x = Composer{w: x.w, opts: x.opts, modules: x.modules}
}

// OneOfHeader is the header of the OneOf Value.
type OneOfHeader struct {
	Index uint
	Type  bsttype.Type
}

// ReadOneOfHeader reads the header of the OneOf Value.
func (x *Extractor) ReadOneOfHeader() (OneOfHeader, error) {
	if x.err != nil {
		return OneOfHeader{}, x.err
	}
	// 1. Check if reading element value is already finished.
	if x.elemDone {
		return OneOfHeader{}, bsterr.Err(bsterr.CodeAlreadyRead, "elem already done")
	}

	// 2. Verify if current element matches the expected type.
	ot, ok := x.elemType.(*bsttype.OneOf)
	if !ok {
		return OneOfHeader{}, bsterr.Err(bsterr.CodeInvalidType, "invalid type element type").
			WithDetails(
				bsterr.D("expected", bsttype.KindOneOf),
				bsterr.D("actual", x.elemType.Kind()),
			)
	}

	// 3. Read the oneOfIndex value.
	idx, n, err := bstio.ReadOneOfIndex(x.r, ot.IndexBytes, x.elemDesc)
	if err != nil {
		return OneOfHeader{}, err
	}
	x.bytesRead += n

	var t bsttype.Type
	for _, elem := range ot.Elements {
		if elem.Index == idx {
			t = elem.Type
			break
		}
	}

	if t == nil {
		return OneOfHeader{}, bsterr.Err(bsterr.CodeInvalidValue, "no matching oneof buffIndex value")
	}

	t, x.err = x.derefType(t)
	if x.err != nil {
		return OneOfHeader{}, x.err
	}
	x.elemType = t
	return OneOfHeader{Index: idx, Type: t}, nil
}
