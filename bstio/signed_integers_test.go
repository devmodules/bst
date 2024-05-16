package bstio

import (
	"math"
	"testing"
)

type dummyWriter struct{}

func (dummyWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (dummyWriter) WriteByte(byte) error {
	return nil
}

func BenchmarkWriteByteWriter(b *testing.B) {
	dw := dummyWriter{}
	for i := 0; i < b.N; i++ {
		writeInt64ByteWriter(dw, 1, false)
	}
}

func BenchmarkWriteInt64(b *testing.B) {
	dw := dummyWriter{}
	for i := 0; i < b.N; i++ {
		WriteInt64(dw, 1, false)
	}
}

func BenchmarkWriteInt(b *testing.B) {
	dw := dummyWriter{}

	opts := []struct {
		Descending, Comparable bool
		name                   string
	}{
		{
			name: "AscNonComp",
		},
		{
			name:       "DescNonComp",
			Descending: true,
		},
		{
			name:       "AscComp",
			Comparable: true,
		},
		{
			name:       "DescComp",
			Comparable: true, Descending: true,
		},
	}
	vals := []struct {
		name string
		val  int
	}{
		{"0", 0},
		{"1", 1},
		{"-1", -1},
		{"MaxInt64", math.MaxInt64},
		{"MinInt64", math.MinInt64},
		{"MaxInt32", math.MaxInt32},
		{"MinInt32", math.MinInt32},
	}
	for _, v := range vals {
		b.Run(v.name, func(b *testing.B) {
			for _, opt := range opts {
				b.Run(opt.name, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						WriteInt(dw, 2143376758786529841, opt.Descending, opt.Comparable)
					}
				})
			}
		})
	}
}
