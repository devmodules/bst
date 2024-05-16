package bstio

import (
	"bytes"
	"io"
	"testing"
)

func BenchmarkString(b *testing.B) {
	const longString = "Why does the moon view? Combine okra, caviar and leek. decorate with melted green curry and serve roasted with asparagus. Enjoy!, The individual is like the cow. Cannabiss messis in moscua! Est fatalis danista, cesaris. How wet. You drink like a corsair. What’s the secret to shredded and tasty chicken lard? Always use delicious black cardamon."
	testCases := []struct {
		Name                   string
		Comparable, Descending bool
		Value                  string
	}{
		{
			Name:       "Short/Asc/Comparable",
			Comparable: true,
			Value:      "short string",
		},
		{
			Name:       "Short/Desc/Comparable",
			Descending: true,
			Comparable: true,
			Value:      "short string",
		},
		{
			Name:  "Short/Asc/Uncomparable",
			Value: "short string",
		},
		{
			Name:       "Short/Desc/Uncomparable",
			Descending: true,
			Value:      "short string",
		},
		{
			Name:       "Long/Asc/Comparable",
			Comparable: true,
			Value:      longString,
		},
		{
			Name:       "Long/Desc/Comparable",
			Descending: true,
			Comparable: true,
			Value:      longString,
		},
		{
			Name:  "Long/Asc/Uncomparable",
			Value: longString,
		},
		{
			Name:       "Long/Desc/Uncomparable",
			Descending: true,
			Value:      longString,
		},
	}

	dw := dummyWriter{}
	b.Run("Write", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(tc.Name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					WriteString(dw, tc.Value, tc.Descending, tc.Comparable)
				}
			})
		}
	})

	b.Run("Read", func(b *testing.B) {
		for _, tc := range testCases {
			buf := bytes.NewBuffer(nil)
			WriteString(buf, tc.Value, tc.Descending, tc.Comparable)
			r := bytes.NewReader(buf.Bytes())
			b.Run(tc.Name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					ReadString(r, tc.Descending, tc.Comparable)
					r.Seek(0, io.SeekStart)
				}
			})
		}
	})

	b.Run("Skip", func(b *testing.B) {
		for _, tc := range testCases {
			b.StopTimer()
			buf := bytes.NewBuffer(nil)
			WriteString(buf, tc.Value, tc.Descending, tc.Comparable)
			r := bytes.NewReader(buf.Bytes())
			b.StartTimer()
			b.Run(tc.Name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					SkipString(r, tc.Descending, tc.Comparable)
					r.Seek(0, io.SeekStart)
				}
			})
		}
	})
}
