//go:build 386 || amd64
// +build 386 amd64

package bstio

import "unsafe"

// The idea taken from the fastXOR in crypto.
const wordSize = int(unsafe.Sizeof(uintptr(0)))

// ReverseBytes reverses the bytes of the given slice.
func ReverseBytes(b []byte) {
	n := len(b)
	w := n / wordSize
	if w > 0 {
		bw := *(*[]uintptr)(unsafe.Pointer(&b))
		for i := 0; i < w; i++ {
			bw[i] = ^bw[i]
		}
	}

	for i := w * wordSize; i < n; i++ {
		b[i] = ^b[i]
	}
}
