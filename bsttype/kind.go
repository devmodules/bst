package bsttype

// _KindTypes is the map of standard types.
var _KindTypes = [...]func(bool) Type{
	KindUndefined: func(shared bool) Type { return getBasic(KindUndefined, shared) },
	KindBoolean:   func(shared bool) Type { return getBasic(KindBoolean, shared) },
	KindInt:       func(shared bool) Type { return getBasic(KindInt, shared) },
	KindInt8:      func(shared bool) Type { return getBasic(KindInt8, shared) },
	KindInt16:     func(shared bool) Type { return getBasic(KindInt16, shared) },
	KindInt32:     func(shared bool) Type { return getBasic(KindInt32, shared) },
	KindInt64:     func(shared bool) Type { return getBasic(KindInt64, shared) },
	KindUint:      func(shared bool) Type { return getBasic(KindUint, shared) },
	KindUint8:     func(shared bool) Type { return getBasic(KindUint8, shared) },
	KindUint16:    func(shared bool) Type { return getBasic(KindUint16, shared) },
	KindUint32:    func(shared bool) Type { return getBasic(KindUint32, shared) },
	KindUint64:    func(shared bool) Type { return getBasic(KindUint64, shared) },
	KindFloat32:   func(shared bool) Type { return getBasic(KindFloat32, shared) },
	KindFloat64:   func(shared bool) Type { return getBasic(KindFloat64, shared) },
	KindString:    func(shared bool) Type { return getBasic(KindString, shared) },
	KindTimestamp: func(shared bool) Type { return getBasic(KindTimestamp, shared) },
	KindDuration:  func(shared bool) Type { return getBasic(KindDuration, shared) },
	KindAny:       func(shared bool) Type { return getBasic(KindAny, shared) },
	KindNamed:     func(shared bool) Type { return getNamed(shared) },
	KindBytes:     func(shared bool) Type { return getBytes(shared) },
	KindStruct:    func(shared bool) Type { return getStruct(shared) },
	KindArray:     func(shared bool) Type { return getArray(shared) },
	KindMap:       func(shared bool) Type { return getMap(shared) },
	KindEnum:      func(shared bool) Type { return getEnum(shared) },
	KindDateTime:  func(shared bool) Type { return getDateTime(shared) },
	KindNullable:  func(shared bool) Type { return getNullable(shared) },
	KindOneOf:     func(shared bool) Type { return getOneOf(shared) },
}

func getBasic(k Kind, shared bool) *Basic {
	if shared {
		return getSharedBasic(k)
	}
	return &Basic{TypeKind: k}
}

func getNamed(shared bool) *Named {
	if shared {
		return getSharedNamed()
	}
	return &Named{}
}

func getBytes(shared bool) *Bytes {
	if shared {
		return getSharedBytes()
	}
	return &Bytes{}
}

func getStruct(shared bool) *Struct {
	if shared {
		return getSharedStruct()
	}
	return &Struct{}
}
func getArray(shared bool) *Array {
	if shared {
		return getSharedArray()
	}
	return &Array{}
}
func getMap(shared bool) *Map {
	if shared {
		return getSharedMap()
	}
	return &Map{}
}
func getEnum(shared bool) *Enum {
	if shared {
		return getSharedEnum()
	}
	return &Enum{}
}
func getDateTime(shared bool) *DateTime {
	if shared {
		return getSharedDateTime()
	}
	return &DateTime{}
}
func getNullable(shared bool) *Nullable {
	if shared {
		return getSharedNullable()
	}
	return &Nullable{}
}
func getOneOf(shared bool) *OneOf {
	if shared {
		return getSharedOneOf()
	}
	return &OneOf{}
}

// emptyKindType returns the standard types.
func emptyKindType(k Kind, shared bool) Type {
	return _KindTypes[k](shared)
}

//go:generate enumer -type=Kind -trimprefix=Kind -output=kind.gen.go

// Kind is a descriptor of the value kind.
type Kind uint8

const (
	// KindUndefined is the kind of the undefined type.
	KindUndefined Kind = iota
	// KindBoolean is the kind of bool values.
	KindBoolean
	// KindInt is the kind of int values.
	KindInt
	// KindInt8 is the kind of int8 values.
	KindInt8
	// KindInt16 is the kind of int16 values.
	KindInt16
	// KindInt32 is the kind of int32 values.
	KindInt32
	// KindInt64 is the kind of int64 values.
	KindInt64
	// KindUint is the kind of uint values.
	KindUint
	// KindUint8 is the kind of uint8 values.
	KindUint8
	// KindUint16 is the kind of uint16 values.
	KindUint16
	// KindUint32 is the kind of uint32 values.
	KindUint32
	// KindUint64 is the kind of uint64 values.
	KindUint64
	// KindFloat32 is the kind of float values.
	KindFloat32
	// KindFloat64 is the kind of double values.
	KindFloat64
	// KindString is the kind of string values.
	KindString
	// KindDuration is the kind of duration values.
	KindDuration
	// KindAny is the kind of the value that could take any of the provided values.
	KindAny
	// KindTimestamp is the kind of timestamp values.
	KindTimestamp
	// KindNamed is a type kind which is defined in a module schema and is identified by a name.
	KindNamed
	// KindBytes is the kind of byte values.
	KindBytes
	// KindStruct is the kind of struct values.
	KindStruct
	// KindArray is the kind of array values.
	KindArray
	// KindMap is the kind of map values.
	KindMap
	// KindEnum is the kind of enum values.
	KindEnum
	// KindDateTime is the kind of the date along with the timezone values.
	KindDateTime
	// KindNullable is the kind of nullable values.
	KindNullable
	// KindOneOf is the kind of the value that could take one of the provided values.
	KindOneOf
)

// IsBasic determines if the kind is basic or its type is composed of more variables.
func (i Kind) IsBasic() bool {
	return i < KindNamed
}
