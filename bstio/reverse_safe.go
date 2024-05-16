//go:build !386 && !amd64
// +build !386,!amd64

package bstio

// ReverseBytes reverses the bytes of the given slice.
func ReverseBytes(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = ^b[i]
	}
}
