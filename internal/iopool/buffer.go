package iopool

import (
	"io"
	"sort"
	"sync"
	"sync/atomic"
)

const (
	minBitSize = 6 // 2**6=64 is a CPU cache line size
	steps      = 20

	minStepSize = 1 << minBitSize
	maxStepSize = 1 << (minBitSize + steps - 1)

	calibrateCallsThreshold = 42000
	maxPercentile           = 0.95
)

// bufferPool represents byte SharedBuffer pool.
//
// Distinct pools may be used for distinct types of byte buffers.
// Properly determined byte SharedBuffer types with their own pools may help reducing memory waste.
type bufferPool struct {
	calls       [steps]uint64
	calibrating uint64

	defaultSize uint64
	maxSize     uint64

	pool sync.Pool
}

var defaultBufferPool = &bufferPool{}

// GetBuffer returns an empty byte SharedBuffer from the pool.
//
// Got byte SharedBuffer may be returned to the pool via Put call.
// This reduces the number of memory allocations required for byte SharedBuffer
// management.
func GetBuffer(root io.Writer) *SharedBuffer {
	return defaultBufferPool.getBuffer(root)
}

// GetBuffer returns new byte SharedBuffer with zero length.
//
// The byte SharedBuffer may be returned to the pool via Put after the use
// in order to minimize GC overhead.
func (p *bufferPool) getBuffer(root io.Writer) *SharedBuffer {
	v := p.pool.Get()
	var buf *SharedBuffer
	if v != nil {
		buf = v.(*SharedBuffer)
	} else {
		buf = &SharedBuffer{Bytes: make([]byte, 0, atomic.LoadUint64(&p.defaultSize))}
	}
	buf.Root = root
	return buf
}

// ReleaseBuffer returns byte SharedBuffer to the pool.
//
// SharedBuffer Bytes mustn't be touched after returning it to the pool.
// Otherwise, data races will occur.
func ReleaseBuffer(b *SharedBuffer) { defaultBufferPool.release(b) }

// release SharedBuffer obtained via GetBuffer to the pool.
//
// The SharedBuffer mustn't be accessed after returning to the pool.
func (p *bufferPool) release(b *SharedBuffer) {
	idx := buffIndex(len(b.Bytes))

	if atomic.AddUint64(&p.calls[idx], 1) > calibrateCallsThreshold {
		p.calibrate()
	}

	maxSize := int(atomic.LoadUint64(&p.maxSize))
	if maxSize == 0 || cap(b.Bytes) <= maxSize {
		b.Reset()
		p.pool.Put(b)
	}
}

func (p *bufferPool) calibrate() {
	if !atomic.CompareAndSwapUint64(&p.calibrating, 0, 1) {
		return
	}

	a := make(callSizes, 0, steps)
	var callsSum uint64
	for i := uint64(0); i < steps; i++ {
		calls := atomic.SwapUint64(&p.calls[i], 0)
		callsSum += calls
		a = append(a, callSize{
			calls: calls,
			size:  minStepSize << i,
		})
	}
	sort.Sort(a)

	defaultSize := a[0].size
	maxSize := defaultSize

	maxSum := uint64(float64(callsSum) * maxPercentile)
	callsSum = 0
	for i := 0; i < steps; i++ {
		if callsSum > maxSum {
			break
		}
		callsSum += a[i].calls
		size := a[i].size
		if size > maxSize {
			maxSize = size
		}
	}

	atomic.StoreUint64(&p.defaultSize, defaultSize)
	atomic.StoreUint64(&p.maxSize, maxSize)

	atomic.StoreUint64(&p.calibrating, 0)
}

type callSize struct {
	calls uint64
	size  uint64
}

type callSizes []callSize

func (ci callSizes) Len() int {
	return len(ci)
}

func (ci callSizes) Less(i, j int) bool {
	return ci[i].calls > ci[j].calls
}

func (ci callSizes) Swap(i, j int) {
	ci[i], ci[j] = ci[j], ci[i]
}

func buffIndex(n int) int {
	n--
	n >>= minBitSize
	idx := 0
	for n > 0 {
		n >>= 1
		idx++
	}
	if idx >= steps {
		idx = steps - 1
	}
	return idx
}

//
// Define a field fieldBuffer.
//

// SharedBuffer provides byte SharedBuffer, which can be used for minimizing memory allocations.
type SharedBuffer struct {
	Bytes []byte
	Root  io.Writer
}

// Len returns the size of the byte fieldBuffer.
func (b *SharedBuffer) Len() int {
	return len(b.Bytes)
}

// ReadFrom implements io.ReaderFrom.
func (b *SharedBuffer) ReadFrom(r io.Reader) (int64, error) {
	p := b.Bytes
	nStart := int64(len(p))
	nMax := int64(cap(p))
	n := nStart
	if nMax == 0 {
		nMax = 64
		p = make([]byte, nMax)
	} else {
		p = p[:nMax]
	}
	for {
		if n == nMax {
			nMax *= 2
			bNew := make([]byte, nMax)
			copy(bNew, p)
			p = bNew
		}
		nn, err := r.Read(p[n:])
		n += int64(nn)
		if err != nil {
			b.Bytes = p[:n]
			n -= nStart
			if err == io.EOF {
				return n, nil
			}
			return n, err
		}
	}
}

// WriteTo implements io.WriterTo.
func (b *SharedBuffer) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(b.Bytes)
	return int64(n), err
}

// Write implements io.Writer - it appends p to fieldBuffer.Bytes
func (b *SharedBuffer) Write(p []byte) (int, error) {
	b.Bytes = append(b.Bytes, p...)
	return len(p), nil
}

// WriteByte appends the byte c to the fieldBuffer.
// Implements io.ByteWriter.
func (b *SharedBuffer) WriteByte(c byte) error {
	b.Bytes = append(b.Bytes, c)
	return nil
}

// Set sets bytes to the input byte slice.
func (b *SharedBuffer) Set(p []byte) {
	b.Bytes = append(b.Bytes[:0], p...)
}

// String returns string representation of fieldBuffer.Bytes.
func (b *SharedBuffer) String() string {
	return string(b.Bytes)
}

// BytesCopy gets a copy of the underlying byte slice.
func (b *SharedBuffer) BytesCopy() []byte {
	cp := make([]byte, len(b.Bytes))
	copy(cp, b.Bytes)
	return cp
}

// Reset makes fieldBuffer.Bytes empty.
func (b *SharedBuffer) Reset() {
	b.Bytes = b.Bytes[:0]
	b.Root = nil
}
