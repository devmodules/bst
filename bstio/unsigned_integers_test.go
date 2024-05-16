package bstio

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestWriteUint(t *testing.T) {
	type args struct {
		uv   uint
		desc bool
	}
	tests := []struct {
		name    string
		args    args
		wantW   []byte
		want    int
		wantErr bool
	}{
		{
			name: "0",
			args: args{
				uv:   0,
				desc: false,
			},
			wantW:   []byte{0},
			want:    1,
			wantErr: false,
		},
		{
			name: "1",
			args: args{
				uv:   1,
				desc: false,
			},
			wantW:   []byte{0x01, 0x01},
			want:    2,
			wantErr: false,
		},
		{
			name: "1/desc",
			args: args{
				uv:   1,
				desc: true,
			},
			wantW:   []byte{^byte(0x01), ^byte(0x01)},
			want:    2,
			wantErr: false,
		},
		{
			name: "256",
			args: args{
				uv:   256,
				desc: false,
			},
			wantW:   []byte{0x02, 0x01, 0x00},
			want:    3,
			wantErr: false,
		},
		{
			name: "256/desc",
			args: args{
				uv:   256,
				desc: true,
			},
			wantW:   []byte{^byte(0x02), ^byte(0x01), ^byte(0x00)},
			want:    3,
			wantErr: false,
		},
		{
			name: "65536",
			args: args{
				uv:   65536,
				desc: false,
			},
			wantW:   []byte{0x03, 0x01, 0x00, 0x00},
			want:    4,
			wantErr: false,
		},
		{
			name: "65536/desc",
			args: args{
				uv:   65536,
				desc: true,
			},
			wantW: []byte{^byte(0x03), ^byte(0x01), ^byte(0x00), ^byte(0x00)},
			want:  4,
		},
		{
			name: "16777216",
			args: args{
				uv:   16777216,
				desc: false,
			},
			wantW: []byte{0x04, 0x01, 0x00, 0x00, 0x00},
			want:  5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			got, err := WriteUint(w, tt.args.uv, tt.args.desc)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteUint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.Bytes(); !bytes.Equal(gotW, tt.wantW) {
				t.Errorf("WriteUint() gotW = %v, want %v", gotW, tt.wantW)
			}

			if got != tt.want {
				t.Errorf("WriteUint() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadUint(t *testing.T) {
	type args struct {
		b    []byte
		desc bool
	}
	tests := []struct {
		name    string
		args    args
		want    uint
		wantErr bool
	}{
		{
			name: "0",
			args: args{
				b:    []byte{0},
				desc: false,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "1",
			args: args{
				b:    []byte{0x01, 0x01},
				desc: false,
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "1/desc",
			args: args{
				b:    []byte{^byte(0x01), ^byte(0x01)},
				desc: true,
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "256",
			args: args{
				b:    []byte{0x02, 0x01, 0x00},
				desc: false,
			},
			want:    256,
			wantErr: false,
		},
		{
			name: "256/desc",
			args: args{
				b:    []byte{^byte(0x02), ^byte(0x01), ^byte(0x00)},
				desc: true,
			},
			want:    256,
			wantErr: false,
		},
		{
			name: "65536",
			args: args{
				b:    []byte{0x03, 0x01, 0x00, 0x00},
				desc: false,
			},
			want:    65536,
			wantErr: false,
		},
		{
			name: "65536/desc",
			args: args{
				b:    []byte{^byte(0x03), ^byte(0x01), ^byte(0x00), ^byte(0x00)},
				desc: true,
			},
			want:    65536,
			wantErr: false,
		},
		{
			name: "16777216",
			args: args{
				b:    []byte{0x04, 0x01, 0x00, 0x00, 0x00},
				desc: false,
			},
			want:    16777216,
			wantErr: false,
		},
		{
			name: "16777216/desc",
			args: args{
				b:    []byte{^byte(0x04), ^byte(0x01), ^byte(0x00), ^byte(0x00), ^byte(0x00)},
				desc: true,
			},
			want:    16777216,
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader(tc.args.b)
			got, bytesRead, err := ReadUint(r, tc.args.desc)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ReadUint() error = %v, wantErr %v", err, tc.wantErr)
			}

			if got != tc.want {
				t.Errorf("ReadUint() got = %v, want %v", got, tc.want)
			}

			if bytesRead != len(tc.args.b) {
				t.Errorf("ReadUint() bytesRead = %v, want %v", bytesRead, len(tc.args.b))
			}
		})
	}
}

type discard struct{}

func (d discard) Write(p []byte) (int, error) {
	return len(p), nil
}

func (d discard) WriteByte(b byte) error {
	return nil
}

func BenchmarkUint(b *testing.B) {
	bw := io.Writer(discard{})
	testCases := []struct {
		name string
		v    uint
		bin  []byte
	}{
		{
			name: "1-byte",
			v:    1,
			bin:  []byte{0x01, 0x01},
		},
		{
			name: "2-byte",
			v:    256,
			bin:  []byte{0x02, 0x01, 0x00},
		},
		{
			name: "3-byte",
			v:    65536,
			bin:  []byte{0x03, 0x01, 0x00, 0x00},
		},
		{
			name: "4-byte",
			v:    16777216,
			bin:  []byte{0x04, 0x01, 0x00, 0x00, 0x00},
		},
		{
			name: "5-byte",
			v:    4294967296,
			bin:  []byte{0x05, 0x01, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "6-byte",
			v:    1099511627776,
			bin:  []byte{0x06, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "7-byte",
			v:    281474976710656,
			bin:  []byte{0x07, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "8-byte",
			v:    72057594037927936,
			bin:  []byte{0x08, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}

	for _, tc := range testCases {
		b.Run("Write", func(b *testing.B) {
			b.Run(tc.name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					WriteUint(bw, tc.v, false)
				}
			})
		})
		b.Run("Read", func(b *testing.B) {
			b.Run(tc.name, func(b *testing.B) {
				r := bytes.NewReader(tc.bin)
				for i := 0; i < b.N; i++ {
					_, _, _ = ReadUint(r, false)
					r.Seek(0, io.SeekStart)
				}
			})
		})
	}
}

func BenchmarkUint64(b *testing.B) {
	b.Run("Write", func(b *testing.B) {
		w := io.Writer(discard{})
		for i := 0; i < b.N; i++ {
			WriteUint64(w, 72057594037927936, false)
		}
	})

	b.Run("Read", func(b *testing.B) {
		r := bytes.NewReader([]byte{0x08, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		for i := 0; i < b.N; i++ {
			_, _, _ = ReadUint64(r, false)
			r.Seek(0, io.SeekStart)
		}
	})
}

func TestCompareUint(t *testing.T) {
	var tests = []struct {
		v1, v2 uint
		res    int
	}{
		{v1: 0, v2: 0, res: 0},
		{v1: 0, v2: 1, res: -1},
		{v1: 1, v2: 0, res: 1},
		{v1: 256, v2: 1, res: 1},
		{v1: 65535, v2: 256, res: 1},
		{v1: 65536, v2: 65536, res: 0},
		{v1: 16777216, v2: 65536, res: 1},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d-%d", tc.v1, tc.v2), func(t *testing.T) {
			buf := &bytes.Buffer{}

			_, err := WriteUint(buf, tc.v1, false)
			if err != nil {
				t.Errorf("WriteUint() error = %v", err)
			}

			v1 := make([]byte, buf.Len())
			copy(v1, buf.Bytes())
			buf.Reset()
			t.Logf("v1: %0x", v1)

			_, err = WriteUint(buf, tc.v2, false)
			if err != nil {
				t.Errorf("WriteUint() error = %v", err)
			}

			v2 := buf.Bytes()
			buf.Reset()
			t.Logf("v2: %0x", v2)

			res := bytes.Compare(v1, v2)
			if res != tc.res {
				t.Errorf("CompareUint() res = %v, want %v", res, tc.res)
			}
		})
	}
}
