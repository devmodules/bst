package bst

import "github.com/devmodules/bst/bsttype"

// ExpectedExtractor is a type that is used to extract the expected type from the binary data.
// This is used to extract the binary data with a type that could have little different data types
// then the embedded type.
// If the embed type or type provided by ExtractorType is different from the expected type then the extractor
// do whatever it can to skip or convert (cast) the extracted data to the expected ones.
// Example:
//   - If the expected type is a string and the embed type is a byte array, then the extractor will
//     convert the byte array to a string.
//   - If the expected type is a struct with lower fields number than the embed type, then the extractor
//     will skip the extra fields and will not fail.
type ExpectedExtractor struct {
	x            *Extractor
	expectedType bsttype.Type
	elemType     bsttype.Type
}
