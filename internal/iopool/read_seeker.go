package iopool

import (
	"errors"
	"io"
)

var (
	_ io.ReadSeeker = (*SharedReadSeeker)(nil)
	_ io.ByteReader = (*SharedReadSeeker)(nil)
)

// SharedReadSeeker is an implementation of io.ReadSeeker and io.ByteReader.
// It could be either used to wrap io.Reader or to be set on top of byte slice.
// It
type SharedReadSeeker struct {
	root                 io.Reader
	buffer               []byte
	streamPos, bufferTop int64
	eof                  bool
}

// Root returns the root reader.
func (w *SharedReadSeeker) Root() io.Reader {
	return w.root
}

// ResetWithRoot sets SharedReadSeeker to an initial position and set the root reader.
func (w *SharedReadSeeker) ResetWithRoot(r io.Reader) {
	w.root = r
	w.streamPos = 0
	w.bufferTop = 0
	w.eof = false
}

// ResetWithBytes sets SharedReadSeeker to an initial position with the buffer set to given byte slice.
func (w *SharedReadSeeker) ResetWithBytes(in []byte) {
	w.streamPos = 0
	w.bufferTop = 0
	w.eof = false
	if len(in) > len(w.buffer) {
		w.buffer = in
	} else {
		copy(w.buffer, in)
	}
	w.bufferTop = int64(len(in))
}

// Seek implements the io.Seeker interface.
func (w *SharedReadSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		w.streamPos = offset
	case io.SeekCurrent:
		w.streamPos += offset
	case io.SeekEnd:
		w.streamPos = w.bufferTop + offset
	default:
		return 0, errors.New("reader.wrappedReadSeeker.Seek invalid whence")
	}
	if w.streamPos > w.bufferTop {
		if w.eof || w.root == nil {
			return w.streamPos, io.EOF
		}
		br, err := w.fillBuffer(int(w.streamPos - w.bufferTop))
		if err != nil {
			return w.streamPos, err
		}
		if w.eof && br == 0 {
			return w.streamPos, io.EOF
		}
	}
	return w.streamPos, nil
}

// Read implements the io.Reader interface.
func (w *SharedReadSeeker) Read(p []byte) (int, error) {
	if w.streamPos >= w.bufferTop {
		if w.eof || w.root == nil {
			return 0, io.EOF
		}
		br, err := w.fillBuffer(len(p))
		if err != nil {
			return 0, err
		}
		if w.eof && br == 0 {
			return 0, io.EOF
		}
	}

	toRead := minInt64(int64(len(p)), w.bufferTop-w.streamPos)
	copy(p, w.buffer[w.streamPos:w.streamPos+toRead])
	w.streamPos += toRead
	return int(toRead), nil
}

// ReadByte implements the io.ByteReader interface.
func (w *SharedReadSeeker) ReadByte() (byte, error) {
	if w.streamPos >= w.bufferTop {
		if w.eof || w.root == nil {
			return 0, io.EOF
		}
		br, err := w.fillBuffer(1)
		if err != nil {
			return 0, err
		}
		if w.eof && br == 0 {
			return 0, io.EOF
		}
	}

	b := w.buffer[w.streamPos]
	w.streamPos++
	return b, nil
}

func (w *SharedReadSeeker) fillBuffer(minToRead int) (int, error) {
	// 1. Check if we need to extend the buffer.
	if w.bufferTop+int64(minToRead) > int64(len(w.buffer)) {
		// 2. Extend the buffer - at least twice.
		size := int64(len(w.buffer)) * 2
		for size < w.bufferTop+int64(minToRead) {
			size *= 2
		}
		newBuffer := make([]byte, size)
		copy(newBuffer, w.buffer)
		w.buffer = newBuffer
	}

	// 3. Estimate the number of bytes to read.
	//    We want to read at least minToRead bytes, but we don't want to read more than the buffer size.
	toRead := maxInt64(int64(minToRead), int64(len(w.buffer))-w.bufferTop)

	// 4. Read the bytes.
	bytesRead, err := w.root.Read(w.buffer[w.streamPos : w.streamPos+toRead])
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return bytesRead, err
		}
		w.eof = true
	}

	w.bufferTop = +int64(bytesRead)
	return bytesRead, nil
}

func (w *SharedReadSeeker) reset() {
	w.streamPos = 0
	w.bufferTop = 0
	w.eof = false
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
