package iopool

import (
	"io"
	"sort"
	"sync"
	"sync/atomic"
)

type readersPool struct {
	calls       [steps]uint64
	calibrating uint64
	defaultSize uint64
	maxSize     uint64

	pool sync.Pool
}

var defaultReadersPool = &readersPool{defaultSize: 2}

// WrapReader wraps input reader to be a SharedReadSeeker.
func WrapReader(root io.Reader) *SharedReadSeeker {
	return defaultReadersPool.wrap(root)
}

// GetReadSeeker returns a SharedReadSeeker with the given bytes.
func GetReadSeeker(in []byte) *SharedReadSeeker {
	return defaultReadersPool.get(in)
}

// ReleaseReadSeeker releases the SharedReadSeeker.
// The reader mustn't be used after calling ReleaseReadSeeker.
func ReleaseReadSeeker(r *SharedReadSeeker) {
	defaultReadersPool.release(r)
}

func (p *readersPool) wrap(root io.Reader) *SharedReadSeeker {
	v := p.pool.Get()
	var r *SharedReadSeeker
	if v == nil {
		r = &SharedReadSeeker{root: root, buffer: make([]byte, 0, atomic.LoadUint64(&p.defaultSize))}
	} else {
		r = v.(*SharedReadSeeker)
		r.ResetWithRoot(root)
	}
	return r
}

func (p *readersPool) get(in []byte) *SharedReadSeeker {
	v := p.pool.Get()
	var r *SharedReadSeeker
	if v != nil {
		r = v.(*SharedReadSeeker)
		r.ResetWithBytes(in)
	} else {
		size := atomic.LoadUint64(&p.defaultSize)
		for len(in) > int(size) {
			size *= 2
		}
		r = &SharedReadSeeker{root: nil, buffer: make([]byte, len(in), size), bufferTop: int64(len(in))}
		copy(r.buffer, in)
	}

	return r
}

func (p *readersPool) release(r *SharedReadSeeker) {
	idx := buffIndex(len(r.buffer))

	if atomic.AddUint64(&p.calls[idx], 1) > calibrateCallsThreshold {
		p.calibrate()
	}

	maxSize := int(atomic.LoadUint64(&p.maxSize))
	if maxSize == 0 || len(r.buffer) <= maxSize {
		r.reset()
		p.pool.Put(r)
	}
}

func (p *readersPool) calibrate() {
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
