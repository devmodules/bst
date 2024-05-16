package bsttype

import (
	"fmt"
	"io"
	"time"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
)

var (
	_ Type         = (*DateTime)(nil)
	_ TypeReader   = (*DateTime)(nil)
	_ TypeWriter   = (*DateTime)(nil)
	_ TypeSkipper  = (*DateTime)(nil)
	_ TypeComparer = (*DateTime)(nil)
)

// Compile-time checks for internal interfaces
var (
	_ copier = (*DateTime)(nil)
)

type (
	// DateTime is the Type implementation for the time.Time along with its timezone.
	// Binary representation of the DateTime is:
	// Size(bits)   | Name                        | Description
	// -------------+-----------------------------+------------
	//     8		| FixedZone                   | Null flag of the FixedZone field.
	//     8        | FixedZone Name Length Size  | Header with the size of the FixedZone Name field length.
	//     0-64     | FixedZone Name Length       | The length of the FixedZone Name field.
	//     0-N	    | FixedZone Name              | The string Name of the FixedZone.
	//     32       | FixedZone Offset            | The Offset of the FixedZone.
	DateTime struct {
		HasFixedZone bool
		FixedZone    DateTimeFixedZone

		needsRelease bool
	}
	// DateTimeFixedZone is the timezone for the DateTime.
	DateTimeFixedZone struct {
		Name   string
		Offset int
	}
)

// Kind returns the basic kind of the value.
func (*DateTime) Kind() Kind {
	return KindDateTime
}

// String returns a human-readable representation of the DateTime.
func (x *DateTime) String() string {
	if !x.HasFixedZone {
		return "DateTime"
	}
	return fmt.Sprintf("DateTime(%s: %d)", x.FixedZone.Name, x.FixedZone.Offset)
}

// Location returns the timezone of the DateTime.
func (x *DateTime) Location() *time.Location {
	if !x.HasFixedZone {
		return nil
	}

	return time.FixedZone(x.FixedZone.Name, x.FixedZone.Offset)
}

// SkipType the bytes in the reader to the next value.
// Implements the TypeSkipper interface.
func (x *DateTime) SkipType(rs io.ReadSeeker) (int64, error) {
	// 1. Read nullable flag. If null, return.
	bt, err := bstio.ReadByte(rs)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read DateTimeType nullable flag")
	}
	bytesSkipped := int64(1)

	// 2. Check if the value is null.
	switch bt {
	case bstio.NullableIsNull:
		return bytesSkipped, nil
	case bstio.NullableIsNotNull:
	default:
		return 0, bsterr.Err(bsterr.CodeDecodingBinaryValue, "invalid DateTimeType nullable flag")
	}

	// 3. Skip the FixedZone.Name field.
	n, err := bstio.SkipNonComparableString(rs, false)
	if err != nil {
		return bytesSkipped, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip DateTimeType FixedZone.Name")
	}
	bytesSkipped += n

	// 4. Skip the FixedZone.Offset field.
	n, err = bstio.SkipInt32(rs)
	if err != nil {
		return bytesSkipped, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to skip DateTimeType FixedZone.Offset")
	}
	bytesSkipped += n

	return bytesSkipped, nil
}

// ReadType reads the value from the byte slice.
// Implements TypeReader interface.
func (x *DateTime) ReadType(r io.Reader) (int, error) {
	// 1. Read the DateTimeTypeDefinition
	nf, err := bstio.ReadNullableFlag(r, false)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read DateTimeType nullable flag")
	}

	// 2. Check if the value is null.
	switch nf {
	case bstio.NullableIsNull:
		x.HasFixedZone = false
		x.FixedZone = DateTimeFixedZone{}
		return 1, nil
	case bstio.NullableIsNotNull:
		x.HasFixedZone = true
	default:
		return 0, bsterr.Err(bsterr.CodeDecodingBinaryValue, "invalid DateTimeType nullable flag")
	}

	// 3. Read the FixedZone.Name field.
	name, n, err := bstio.ReadStringNonComparable(r, false)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read DateTimeType FixedZone.Name")
	}
	bytesRead := n + 1

	// 4. Read the FixedZone.Offset field.
	offset, n, err := bstio.ReadInt32(r, false)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read DateTimeType FixedZone.Offset")
	}
	bytesRead += n

	// 5. Create the FixedZone.
	x.FixedZone = DateTimeFixedZone{
		Name:   name,
		Offset: int(offset),
	}
	return bytesRead, nil
}

// WriteType writes the value to the writer.
// Implements TypeWriter interface.
func (x *DateTime) WriteType(w io.Writer) (int, error) {
	// 1. Write the nullable flag.
	flag := bstio.NullableIsNotNull
	if !x.HasFixedZone {
		flag = bstio.NullableIsNull
	}
	err := bstio.WriteByte(w, flag)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write DateTimeType nullable flag")
	}

	if flag == bstio.NullableIsNull {
		return 1, nil
	}
	bytesRead := 1

	// 2. Write the FixedZone.Name field.
	var n int
	n, err = bstio.WriteString(w, x.FixedZone.Name, false, false)
	if err != nil {
		return bytesRead, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write DateTimeType FixedZone.Name")
	}
	bytesRead += n

	// 3. Write the FixedZone.Offset field.
	n, err = bstio.WriteInt32(w, int32(x.FixedZone.Offset), false)
	if err != nil {
		return bytesRead, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write DateTimeType FixedZone.Offset")
	}
	bytesRead += n

	return bytesRead, nil
}

// CompareType returns true if the two types are equal.
func (x *DateTime) CompareType(to TypeComparer) bool {
	dtTo, ok := to.(*DateTime)
	if !ok {
		return false
	}

	if (x.HasFixedZone && !dtTo.HasFixedZone) || (!x.HasFixedZone && dtTo.HasFixedZone) {
		return false
	}

	if !x.HasFixedZone && !dtTo.HasFixedZone {
		return true
	}

	return x.FixedZone.Name == dtTo.FixedZone.Name && x.FixedZone.Offset == dtTo.FixedZone.Offset
}

func (x *DateTime) copy(shared bool) Type {
	var dt *DateTime
	if !shared {
		dt = &DateTime{}
	} else {
		dt = getSharedDateTime()
	}
	*dt = *x
	return dt
}

//
// Shared Pool
//

var _sharedDateTimePool = &sharedPool{defaultSize: 10}

func getSharedDateTime() *DateTime {
	v := _sharedDateTimePool.pool.Get()
	st, ok := v.(*DateTime)
	if ok {
		return st
	}
	return &DateTime{
		needsRelease: true,
	}
}

func putSharedDateTime(x *DateTime) {
	if !x.needsRelease {
		return
	}
	*x = DateTime{needsRelease: true}
	_sharedDateTimePool.pool.Put(x)
}
