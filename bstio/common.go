package bstio

import (
	"io"
)

// Constant binary values.
const (
	BinarySizeZero   uint8 = 0x00
	BinarySizeUint8  uint8 = 0x01
	BinarySizeUint16 uint8 = 0x02
	BinarySizeUint32 uint8 = 0x04
	BinarySizeUint64 uint8 = 0x08

	BinaryPositiveZero = 0x00 | PositiveBit8Mask

	NegativeBit8Mask = uint8(2<<6 - 1)
	PositiveBit8Mask = uint8(1 << 7)
)

// ValueOptions is a set of options for marshaling values.
type ValueOptions struct {
	// Descending lets know encoders whether the value should have
	// its binary encoded in a descending manner.
	// This might be used for the indexes with descending order, where the binaries
	// are compared directly.
	Descending bool
	// Comparable determines that the value binary maintains sortable order, while comparing
	// raw bytes. This is useful for indexes, where the binaries could be stored
	// and compared directly if its values are sortable.
	// This option makes the parsing and encoding slower.
	Comparable bool
	// CompatibilityMode determines that the value binary is compatible with the old
	// encoding format.
	CompatibilityMode bool
}

// ReadByte reads a single byte from the reader.
func ReadByte(r io.Reader) (byte, error) {
	if br, ok := r.(io.ByteReader); ok {
		return br.ReadByte()
	}
	b := make([]byte, 1)
	_, err := r.Read(b)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

// WriteByte writes a single byte in an efficient way.
func WriteByte(w io.Writer, b byte) error {
	if bw, ok := w.(io.ByteWriter); ok {
		return bw.WriteByte(b)
	}
	_, err := w.Write([]byte{b})
	return err
}
