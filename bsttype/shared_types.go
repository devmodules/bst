package bsttype

import (
	"fmt"
)

// GetSharedType returns the shared type of the given type.
func GetSharedType(k Kind) Type {
	switch k {
	case KindStruct:
		return getSharedStruct()
	case KindOneOf:
		return getSharedOneOf()
	case KindMap:
		return getSharedMap()
	case KindArray:
		return getSharedArray()
	case KindNullable:
		return getSharedNullable()
	case KindDateTime:
		return getSharedDateTime()
	case KindBytes:
		return getSharedBytes()
	case KindEnum:
		return getSharedEnum()
	default:
		return getSharedBasic(k)
	}
}

// PutSharedType puts the shared type back into the cache.
func PutSharedType(t Type) {
	switch tp := t.(type) {
	case *Struct:
		for _, f := range tp.Fields {
			PutSharedType(f.Type)
		}
		putSharedStruct(tp)
	case *OneOf:
		for _, elem := range tp.Elements {
			PutSharedType(elem.Type)
		}
		putSharedOneOf(tp)
	case *Map:
		PutSharedType(tp.Key.Type)
		PutSharedType(tp.Value.Type)
		putSharedMap(tp)
	case *Array:
		PutSharedType(tp.Type)
		putSharedArray(tp)
	case *Nullable:
		PutSharedType(tp.Type)
		putSharedNullable(tp)
	case *DateTime:
		putSharedDateTime(tp)
	case *Bytes:
		putSharedBytes(tp)
	case *Enum:
		putSharedEnum(tp)
	case *Basic:
		putSharedBasic(tp)
	case *Named:
		putSharedNamed(tp)
	default:
		panic(fmt.Sprintf("unexpected type: %T", tp))
	}
}
