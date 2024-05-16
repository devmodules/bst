package bsterr

import (
	"fmt"
	"strings"
)

// ErrCode is error code type wrapper.
type ErrCode uint16

const (
	// CodeMismatchingValueType is the error code for mismatching value type.
	CodeMismatchingValueType ErrCode = 1001
	// CodeMissingFixedSizeValues is the error code used when the values that should have a fixed size are missing.
	CodeMissingFixedSizeValues ErrCode = 1002
	// CodeInvalidIntegerBytesValue is the error code used when the enum bits value is invalid.
	CodeInvalidIntegerBytesValue ErrCode = 1003
	// CodeTypeConstraintViolation is the error code for type constraint violation.
	CodeTypeConstraintViolation ErrCode = 1004
	// CodeInvalidValue is the error code for invalid value.
	CodeInvalidValue ErrCode = 1005
	// CodeInvalidType is the error code for invalid type.
	CodeInvalidType ErrCode = 1006
	// CodeUndefinedType is the error code for undefined type.
	CodeUndefinedType ErrCode = 1007
	// CodeTypeNotMapped is the error code for type not mapped.
	CodeTypeNotMapped ErrCode = 1008

	// CodeDecodingBinaryValue is the error code for invalid binary value.
	// This is expected as internal error as the values stored in a binary format should not be
	// modified outside.
	CodeDecodingBinaryValue ErrCode = 2001
	// CodeEncodingBinaryValue is the error code for invalid binary value for encoding.
	CodeEncodingBinaryValue ErrCode = 2002
	// CodeUndefinedValue is the error code when one of the values are undefined.
	CodeUndefinedValue ErrCode = 2003
	// CodeSkippingBinaryValue is the error code for invalid binary value for skipping.
	CodeSkippingBinaryValue ErrCode = 2004
	// CodeConvertingIntoBinaryValue is the error code for invalid binary value for converting.
	CodeConvertingIntoBinaryValue ErrCode = 2005
	// CodeValueFieldMissing is the error code for missing value field.
	CodeValueFieldMissing ErrCode = 2006

	// CodeEncodingTypeUndefined is the error code for undefined value type for encoding.
	CodeEncodingTypeUndefined ErrCode = 3001
	// CodeDecodingBinaryType is the error code for invalid binary type.
	// This code occurs when something happened while decoding binary type.
	CodeDecodingBinaryType ErrCode = 3002
	// CodeEncodingBinaryType is the error code for invalid binary type for encoding.
	CodeEncodingBinaryType ErrCode = 3003
	// CodeSkippingBinaryType is the error code for invalid binary value for skipping.
	CodeSkippingBinaryType ErrCode = 3004

	// CodeReadingFailed is the error code when reading data failed.
	CodeReadingFailed ErrCode = 5001
	// CodeWritingFailed is the error code for writing failed.
	CodeWritingFailed ErrCode = 5002
	// CodeInternalSchemaStructureUndefined is the error code for schema structure undefined.
	CodeInternalSchemaStructureUndefined ErrCode = 5002

	// CodeAlreadyRead is an error code for situation where the element is already read.
	CodeAlreadyRead ErrCode = 6001
	// CodeNotReadYet is an error code for situation where the map key is not read yet.
	CodeNotReadYet ErrCode = 6002
	// CodeOutOfBounds is an error code for situation where the buffIndex is out of bounds.
	CodeOutOfBounds ErrCode = 6003
	// CodeAlreadyWritten is an error code for situation where the element is already written.
	CodeAlreadyWritten ErrCode = 6004
	// CodeNotWrittenYet is an error code for situation where the map key is not written yet.
	CodeNotWrittenYet ErrCode = 6005
	// CodeMalformedBinary is an error code for situation where the binary is malformed.
	CodeMalformedBinary ErrCode = 6006
	// CodeTypeAlreadyMapped is an error code for situation where the type is already mapped.
	CodeTypeAlreadyMapped ErrCode = 6007
	// CodeCyclicDependency is an error when the cyclic dependency is detected.
	CodeCyclicDependency ErrCode = 6008
	// CodeModulesUndefined is an error code for situation where the modules are undefined.
	CodeModulesUndefined ErrCode = 6009
)

var _ error = (*Error)(nil)

// Error is the error type used by the eserror package.
type Error struct {
	Code    ErrCode
	Msg     string
	Details []ErrorDetail
	Wrapped error
}

// Error implements the error interface.
func (e *Error) Error() string {
	str := fmt.Sprintf("%d: %s - %s", e.Code, e.Msg, Ds(e.Details))
	if e.Wrapped != nil {
		str += fmt.Sprintf("\n%s", e.Wrapped)
	}
	return str
}

// WithDetail sets the details of the error.
func (e *Error) WithDetail(key string, value interface{}) *Error {
	e.Details = append(e.Details, ErrorDetail{Key: key, Value: value})
	return e
}

// Is returns true if the given error is an equal to the given one.
func (e *Error) Is(cmp error) bool {
	if e == nil {
		return cmp == nil
	}

	xe, ok := cmp.(*Error)
	if !ok || xe == nil {
		return false
	}
	return e.Code == xe.Code && e.Msg == xe.Msg
}

// Unwrap returns the underlying error.
func (e Error) Unwrap() error {
	return e.Wrapped
}

// Wrap wraps the given error with the given code and message.
func (e *Error) Wrap(err error) *Error {
	e.Wrapped = err
	return e
}

// Ds is the wrapper for the array of details.
type Ds []ErrorDetail

func (d Ds) String() string {
	var sb strings.Builder
	for i, detail := range d {
		sb.WriteString(detail.String())
		if i < len(d)-1 {
			sb.WriteString(", ")
		}
	}
	return sb.String()
}

// D creates a detail with the given key and value.
func D(key string, value interface{}) ErrorDetail {
	return ErrorDetail{Key: key, Value: value}
}

// ErrorDetail is a simple key-value pair.
type ErrorDetail struct {
	Key   string
	Value interface{}
}

func (d ErrorDetail) String() string {
	return fmt.Sprintf("%s: %v", d.Key, d.Value)
}

// WithDetails sets the details of the error.
func (e *Error) WithDetails(details ...ErrorDetail) *Error {
	e.Details = append(e.Details, details...)
	return e
}

// Err creates an error with the given code and message.
func Err(code ErrCode, msg string) *Error {
	return &Error{Code: code, Msg: msg}
}

// Errf creates an error with the given code and formatted message.
func Errf(code ErrCode, format string, args ...interface{}) *Error {
	return &Error{Code: code, Msg: fmt.Sprintf(format, args...)}
}

// ErrWrap  wraps the given error with the given code and message.
func ErrWrap(err error, code ErrCode, msg string) *Error {
	return &Error{Code: code, Msg: msg, Wrapped: err}
}

// ErrWrapf wraps the given error with the given code and formatted message.
func ErrWrapf(err error, code ErrCode, format string, args ...interface{}) *Error {
	return &Error{Code: code, Msg: fmt.Sprintf(format, args...), Wrapped: err}
}
