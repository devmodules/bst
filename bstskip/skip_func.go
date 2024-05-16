package bstskip

import (
	"io"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bsttype"
)

// SkipFuncOf gets a skip function of given type.
func SkipFuncOf(t bsttype.Type) SkipFunc {
	return _SkipFuncs[t.Kind()](t)
}

// SkipFunc is a function that skips a value.
type SkipFunc func(br io.ReadSeeker, options bstio.ValueOptions) (int64, error)

var _SkipFuncs = [bsttype.KindOneOf + 1]func(bsttype.Type) SkipFunc{
	bsttype.KindUndefined: func(t bsttype.Type) SkipFunc { return undefinedSkipFunc },
	bsttype.KindBoolean:   func(t bsttype.Type) SkipFunc { return booleanSkipFunc },
	bsttype.KindInt:       func(t bsttype.Type) SkipFunc { return intSkipFunc },
	bsttype.KindInt8:      func(t bsttype.Type) SkipFunc { return int8SkipFunc },
	bsttype.KindInt16:     func(t bsttype.Type) SkipFunc { return int16SkipFunc },
	bsttype.KindInt32:     func(t bsttype.Type) SkipFunc { return int32SkipFunc },
	bsttype.KindInt64:     func(t bsttype.Type) SkipFunc { return int64SkipFunc },
	bsttype.KindUint:      func(t bsttype.Type) SkipFunc { return uintSkipFunc },
	bsttype.KindUint8:     func(t bsttype.Type) SkipFunc { return uint8SkipFunc },
	bsttype.KindUint16:    func(t bsttype.Type) SkipFunc { return uint16SkipFunc },
	bsttype.KindUint32:    func(t bsttype.Type) SkipFunc { return uint32SkipFunc },
	bsttype.KindUint64:    func(t bsttype.Type) SkipFunc { return uint64SkipFunc },
	bsttype.KindFloat32:   func(t bsttype.Type) SkipFunc { return float32SkipFunc },
	bsttype.KindFloat64:   func(t bsttype.Type) SkipFunc { return float64SkipFunc },
	bsttype.KindString:    func(t bsttype.Type) SkipFunc { return stringSkipFunc },
	bsttype.KindDuration:  func(t bsttype.Type) SkipFunc { return int64SkipFunc },
	bsttype.KindTimestamp: func(t bsttype.Type) SkipFunc { return int64SkipFunc },
	bsttype.KindBytes:     func(t bsttype.Type) SkipFunc { return bytesSkipFunc(t.(*bsttype.Bytes)) },
	bsttype.KindEnum:      func(t bsttype.Type) SkipFunc { return enumSkipFunc(t.(*bsttype.Enum)) },
}

func init() {
	_SkipFuncs[bsttype.KindNamed] = func(t bsttype.Type) SkipFunc { return namedSkipFunc(t.(*bsttype.Named)) }
	_SkipFuncs[bsttype.KindStruct] = func(t bsttype.Type) SkipFunc { return structSkipFunc(t.(*bsttype.Struct)) }
	_SkipFuncs[bsttype.KindArray] = func(t bsttype.Type) SkipFunc { return arraySkipFunc(t.(*bsttype.Array)) }
	_SkipFuncs[bsttype.KindMap] = func(t bsttype.Type) SkipFunc { return mapSkipFunc(t.(*bsttype.Map)) }
	_SkipFuncs[bsttype.KindNullable] = func(t bsttype.Type) SkipFunc { return nullableSkipFunc(t.(*bsttype.Nullable)) }
	_SkipFuncs[bsttype.KindOneOf] = func(t bsttype.Type) SkipFunc { return oneOfSkipFunc(t.(*bsttype.OneOf)) }
	_SkipFuncs[bsttype.KindAny] = func(t bsttype.Type) SkipFunc { return SkipAny }
}

func undefinedSkipFunc(_ io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return 0, bsterr.Err(bsterr.CodeUndefinedType, "undefined type cannot be skipped")
}

func booleanSkipFunc(br io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipBool(br)
}

func intSkipFunc(rs io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
	return bstio.SkipInt(rs, options.Descending, options.Comparable)
}
func int8SkipFunc(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipInt8(rs)
}
func int16SkipFunc(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipInt16(rs)
}
func int32SkipFunc(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipInt32(rs)
}
func int64SkipFunc(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipInt64(rs)
}
func uintSkipFunc(rs io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
	return bstio.SkipUint(rs, options.Descending)
}
func uint8SkipFunc(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	_, err := rs.Seek(1, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	return 1, nil
}
func uint16SkipFunc(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipUint16(rs)
}
func uint32SkipFunc(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipUint32(rs)
}
func uint64SkipFunc(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipUint64(rs)
}
func float32SkipFunc(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipFloat32(rs)
}
func float64SkipFunc(rs io.ReadSeeker, _ bstio.ValueOptions) (int64, error) {
	return bstio.SkipFloat64(rs)
}

func stringSkipFunc(rs io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
	return bstio.SkipString(rs, options.Descending, options.Comparable)
}

// SkipBytes skips the bsttype.Bytes value.
func SkipBytes(rs io.ReadSeeker, bt *bsttype.Bytes, options bstio.ValueOptions) (int64, error) {
	return bytesSkipFunc(bt)(rs, options)
}

func bytesSkipFunc(bt *bsttype.Bytes) SkipFunc {
	return func(rs io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
		return bstio.SkipBytes(rs, bt.FixedSize, options.Descending, options.Comparable)
	}
}

func enumSkipFunc(et *bsttype.Enum) SkipFunc {
	return func(rs io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
		return bstio.SkipEnumIndex(rs, et.ValueBytes, options.Descending)
	}
}

func nullableSkipFunc(nt *bsttype.Nullable) SkipFunc {
	return SkipFuncOf(nt.Type)
}

func namedSkipFunc(nt *bsttype.Named) SkipFunc {
	return SkipFuncOf(nt.Type)
}
