package bstio

import (
	"io"
	"time"

	"github.com/devmodules/bst/bsterr"
)

// ReadDateTime reads binary encoded DateTime value from the given reader.
// The desc flag indicates if the value is encoded in descending order.
// Returns the number of bytes read and an error if any.
func ReadDateTime(r io.Reader, desc bool, loc *time.Location) (time.Time, int, error) {
	// 1. Read the Version Byte.
	ver, err := ReadByte(r)
	if err != nil {
		return time.Time{}, 0, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read DateTimeValue version")
	}
	bytesRead := 1

	if desc {
		ver = ^ver
	}

	// 2. Estimate the size of the binary value.
	wantLen := /*sec*/ 8 + /*nsec*/ 4 + /*zone offset*/ 2
	if ver == 2 {
		wantLen++
	}
	bin := make([]byte, wantLen+1)
	bin[0] = ver

	// 3. Read the binary value.
	n, err := r.Read(bin[1:])
	if err != nil {
		return time.Time{}, bytesRead, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read DateTimeValue")
	}
	bytesRead += n

	// 4. For descending values, we need to ReverseBytes the byte order.
	if desc {
		ReverseBytes(bin[1:])
	}

	var tm time.Time

	// 5. Decode the binary value.
	err = tm.UnmarshalBinary(bin)
	if err != nil {
		return time.Time{}, bytesRead, bsterr.ErrWrap(err, bsterr.CodeDecodingBinaryValue, "failed to read DateTimeValue")
	}

	// 6. If the timezone is defined in the type, apply it on the value.
	if loc != nil {
		tm = tm.In(loc)
	}
	return tm, bytesRead, nil
}

// WriteDateTime writes binary encoded DateTime value to the given writer.
// The desc flag indicates if the value is encoded in descending order.
// Returns the number of bytes written and an error if any.
func WriteDateTime(w io.Writer, tm time.Time, desc bool, loc *time.Location) (int, error) {
	// 1. If the timezone location is defined set it on the time value.
	if loc != nil {
		tm = tm.In(loc)
	}

	// 2. Encode the time into a binary format.
	bin, err := tm.MarshalBinary()
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to marshal DateTimeValue").
			WithDetails(bsterr.D("value", tm))
	}

	// 3. If the value is encoded in descending order, ReverseBytes the byte order.
	if desc {
		ReverseBytes(bin)
	}

	// 4. Write the binary value.
	var n int
	n, err = w.Write(bin)
	if err != nil {
		return n, bsterr.ErrWrap(err, bsterr.CodeWritingFailed, "failed to write DateTimeValue").
			WithDetail("value", tm)
	}
	return n, nil
}

// SkipDateTime skips binary encoded DateTime value from the given reader.
func SkipDateTime(rs io.ReadSeeker, desc bool) (int64, error) {
	// 1. Read the version byte.
	ver, err := ReadByte(rs)
	if err != nil {
		return 0, bsterr.ErrWrap(err, bsterr.CodeSkippingBinaryValue, "failed to skip DateTimeValue")
	}
	skipped := int64(1)

	if desc {
		ver = ^ver
	}
	wantLen := /*sec*/ 8 + /*nsec*/ 4 + /*zone offset*/ 2
	if ver == 2 {
		wantLen++
	}

	// 2. Skip the binary value.
	_, err = rs.Seek(int64(wantLen), io.SeekCurrent)
	if err != nil {
		return skipped, bsterr.ErrWrap(err, bsterr.CodeSkippingBinaryValue, "failed to skip DateTimeValue")
	}
	skipped += int64(wantLen)

	return skipped, nil
}
