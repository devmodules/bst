package bsttype

// Compile-time checks for Basic interface implementations.
var (
	_ Type   = (*Basic)(nil)
	_ copier = (*Basic)(nil)
)

// Basic is a descriptor for the simple type definitions.
type Basic struct {
	TypeKind Kind

	isShared bool
}

// Kind returns a basic type kind.
func (b *Basic) Kind() Kind {
	return b.TypeKind
}

// String returns a human-readable string representation of the Basic.
func (b *Basic) String() string {
	return b.TypeKind.String()
}

// Reset resets the Basic to its default state.
func (b *Basic) Reset() {
	*b = Basic{}
}

// NeedsRelease returns true if the Basic needs to be released.
func (b *Basic) NeedsRelease() bool {
	return b.isShared
}

// Free is used to release a shared basic type.
func (b *Basic) Free() {
	if !b.isShared {
		return
	}
	putSharedBasic(b)
}

func (b *Basic) copy(shared bool) Type {
	if shared {
		return getSharedBasic(b.TypeKind)
	}
	return &Basic{TypeKind: b.TypeKind}
}

// Any gets the basic type that represents the Any type.
func Any() *Basic {
	return &Basic{TypeKind: KindAny}
}

// AnyShared gets the basic type that represents the Any type from the shared type pool.
// This type should be Freed after use.
func AnyShared() *Basic {
	return getSharedBasic(KindAny)
}

// Boolean gets the basic type that represents the Boolean type.
func Boolean() *Basic {
	return &Basic{TypeKind: KindBoolean}
}

// BooleanShared gets the basic type that represents the Boolean type from the shared type pool.
// This type should be Freed after use.
func BooleanShared() *Basic {
	return getSharedBasic(KindBoolean)
}

// Duration gets the basic type that represents the Duration type.
func Duration() *Basic {
	return &Basic{TypeKind: KindDuration}
}

// DurationShared gets the basic type that represents the Duration type from the shared type pool.
// This type should be Freed after use.
func DurationShared() *Basic {
	return getSharedBasic(KindDuration)
}

// Float32 gets the basic type that represents the Float32 type.
func Float32() *Basic {
	return &Basic{TypeKind: KindFloat32}
}

// Float32Shared gets a float32 type from the shared type pool.
// This type should be Freed after use.
func Float32Shared() *Basic {
	return getSharedBasic(KindFloat32)
}

// Float64 gets the basic type that represents the Float64 type.
func Float64() *Basic {
	return &Basic{TypeKind: KindFloat64}
}

// Float64Shared gets a float64 type from the shared type pool.
// This type should be Freed after use.
func Float64Shared() *Basic {
	return getSharedBasic(KindFloat64)
}

// String gets the string representation of the Basic.
func String() *Basic {
	return &Basic{TypeKind: KindString}
}

// StringShared gets a string type from the shared type pool.
// This type should be Freed after use.
func StringShared() *Basic {
	return getSharedBasic(KindString)
}

// Int8 gets a Int8 representation type.
func Int8() *Basic {
	return &Basic{TypeKind: KindInt8}
}

// Int8Shared gets Int8 type representation from the shared type pool.
// This type should be Freed after use.
func Int8Shared() *Basic {
	return getSharedBasic(KindInt8)
}

// Int16 gets a Int16 representation type.
func Int16() *Basic {
	return &Basic{TypeKind: KindInt16}
}

// Int16Shared gets Int16 type representation from the shared type pool.
// This type should be Freed after use.
func Int16Shared() *Basic {
	return getSharedBasic(KindInt16)
}

// Int32 gets a Int32 representation type.
func Int32() *Basic {
	return &Basic{TypeKind: KindInt32}
}

// Int32Shared gets Int32 type representation from the shared type pool.
// This type should be Freed after use.
func Int32Shared() *Basic {
	return getSharedBasic(KindInt32)
}

// Int64 gets a Int64 representation type.
func Int64() *Basic {
	return &Basic{TypeKind: KindInt64}
}

// Int64Shared gets Int64 type representation from the shared type pool.
// This type should be Freed after use.
func Int64Shared() *Basic {
	return getSharedBasic(KindInt64)
}

// Int gets a Int representation type.
func Int() *Basic {
	return &Basic{TypeKind: KindInt}
}

// IntShared gets Int type representation from the shared type pool.
// This type should be Freed after use.
func IntShared() *Basic {
	return getSharedBasic(KindInt)
}

// Uint8 gets Uint8 type representation.
func Uint8() *Basic {
	return &Basic{TypeKind: KindUint8}
}

// Uint8Shared gets Uint8 type representation from the shared type pool.
// This type should be Freed after use.
func Uint8Shared() *Basic {
	return &Basic{TypeKind: KindUint8}
}

// Uint16 gets Uint16 type representation.
func Uint16() *Basic {
	return &Basic{TypeKind: KindUint16}
}

// Uint16Shared gets Uint16 type representation from the shared type pool.
// This type should be Freed after use.
func Uint16Shared() *Basic {
	return &Basic{TypeKind: KindUint16}
}

// Uint32 gets Uint32 type representation.
func Uint32() *Basic {
	return &Basic{TypeKind: KindUint32}
}

// Uint32Shared gets Uint32 type representation from the shared type pool.
// This type should be Freed after use.
func Uint32Shared() *Basic {
	return &Basic{TypeKind: KindUint32}
}

// Uint64 gets Uint64 type representation.
func Uint64() *Basic {
	return &Basic{TypeKind: KindUint64}
}

// Uint64Shared gets Uint64 type representation from the shared type pool.
// This type should be Freed after use.
func Uint64Shared() *Basic {
	return &Basic{TypeKind: KindUint64}
}

// Uint gets Uint type representation.
func Uint() *Basic {
	return &Basic{TypeKind: KindUint}
}

// UintShared gets Uint type representation from the shared type pool.
// This type should be Freed after use.
func UintShared() *Basic {
	return &Basic{TypeKind: KindUint}
}

// Timestamp gets Timestamp type representation.
func Timestamp() *Basic {
	return &Basic{TypeKind: KindTimestamp}
}

// TimestampShared gets Timestamp type representation from the shared type pool.
// This type should be Freed after use.
func TimestampShared() *Basic {
	return getSharedBasic(KindTimestamp)
}

//
// Shared Pool
//

var basicPool = &sharedPool{}

func getSharedBasic(k Kind) *Basic {
	bt, ok := basicPool.pool.Get().(*Basic)
	if !ok {
		return &Basic{isShared: true, TypeKind: k}
	}
	bt.isShared = true
	bt.TypeKind = k
	return bt
}

func putSharedBasic(bt *Basic) {
	bt.Reset()
	basicPool.pool.Put(bt)
}
