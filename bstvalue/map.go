package bstvalue

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/devmodules/bst/bsterr"
	"github.com/devmodules/bst/bstio"
	"github.com/devmodules/bst/bstskip"
	"github.com/devmodules/bst/bsttype"
	"github.com/google/btree"
)

// Compile time check to ensure that MapValue implements the Value interface.
var _ Value = (*MapValue)(nil)

type (
	// MapValue is the value for a map type.
	MapValue struct {
		MapType *bsttype.Map
		btree   *btree.BTree
	}

	// MapValueKV is the key value pair for a map value.
	MapValueKV struct {
		Key   Value
		Value Value
	}

	// mapValueKV is the implementation of the btree.Item.
	mapValueKV struct {
		kb                 []byte
		keyDesc, valueDesc bool
		Key, Value         Value
	}
)

// EmptyMapValue returns an empty map value.
func EmptyMapValue(mt *bsttype.Map) *MapValue {
	return &MapValue{
		MapType: mt,
		btree:   btree.New(2),
	}
}

// MustNewMapValue returns a new map value or panics.
func MustNewMapValue(mt *bsttype.Map, kvs ...MapValueKV) *MapValue {
	mv, err := NewMapValue(mt, kvs...)
	if err != nil {
		panic(err)
	}
	return mv
}

// NewMapValue returns a new map value.
func NewMapValue(mt *bsttype.Map, kvs ...MapValueKV) (*MapValue, error) {
	mv := &MapValue{MapType: mt, btree: btree.New(2)}
	for _, kv := range kvs {
		if err := mv.Put(kv.Key, kv.Value); err != nil {
			return nil, err
		}
	}
	return mv, nil
}

func emptyMapValue(mt bsttype.Type) Value {
	return &MapValue{
		MapType: mt.(*bsttype.Map),
		btree:   btree.New(2),
	}
}

// String returns a string representation of the value.
func (x *MapValue) String() string {
	var b strings.Builder
	b.WriteString(x.MapType.String())
	b.WriteString("{")
	first := true
	x.btree.Ascend(func(i btree.Item) bool {
		if !first {
			b.WriteString(", ")
		}
		first = false
		kv := i.(*mapValueKV)
		b.WriteString(kv.Key.String())
		b.WriteString(": ")
		b.WriteString(kv.Value.String())
		return true
	})
	b.WriteString("}")
	return b.String()
}

// Put adds a key value pair to the map value.
func (x *MapValue) Put(key, value Value) error {
	if !bsttype.TypesEqual(x.MapType.Key.Type, key.Type()) {
		return bsterr.Err(bsterr.CodeMismatchingValueType, "map key type mismatch").
			WithDetails(
				bsterr.D("expected", x.MapType.Key),
				bsterr.D("actual", key.Type()),
			)
	}
	if !bsttype.TypesEqual(x.MapType.Value.Type, value.Type()) {
		return bsterr.Err(bsterr.CodeMismatchingValueType, "map value type mismatch").
			WithDetails(
				bsterr.D("expected", x.MapType.Value),
				bsterr.D("actual", value.Type()),
			)
	}

	data, err := key.MarshalValue(bstio.ValueOptions{})
	if err != nil {
		return err
	}

	i := &mapValueKV{
		kb:      data,
		keyDesc: x.MapType.Key.Descending,
		Key:     key,
		Value:   value,
	}
	x.btree.ReplaceOrInsert(i)

	return nil
}

// Get returns the value for the given key.
func (x *MapValue) Get(key Value) (Value, bool, error) {
	data, err := key.MarshalValue(bstio.ValueOptions{})
	if err != nil {
		return nil, false, err
	}

	k := &mapValueKV{kb: data}
	v := x.btree.Get(k)
	if v == nil {
		return nil, false, nil
	}
	return v.(*mapValueKV).Value, true, nil
}

// Has returns true if the map value has the given key.
func (x *MapValue) Has(key Value) (bool, error) {
	data, err := key.MarshalValue(bstio.ValueOptions{})
	if err != nil {
		return false, err
	}

	k := &mapValueKV{kb: data}
	return x.btree.Has(k), nil
}

// Delete removes the key value pair from the map value.
// Returns true if the key was found and removed.
func (x *MapValue) Delete(key Value) (bool, error) {
	data, err := key.MarshalValue(bstio.ValueOptions{})
	if err != nil {
		return false, err
	}

	k := &mapValueKV{kb: data}
	v := x.btree.Delete(k)
	return v != nil, nil
}

// KeyValues returns the key value pairs for the map value.
func (x *MapValue) KeyValues() []MapValueKV {
	kvs := make([]MapValueKV, 0, x.btree.Len())
	x.btree.Ascend(func(i btree.Item) bool {
		kvs = append(kvs, MapValueKV{
			Key:   i.(*mapValueKV).Key,
			Value: i.(*mapValueKV).Value,
		})
		return true
	})
	return kvs
}

// IterCtx returns an iterator for the map value.
func (x *MapValue) IterCtx(ctx context.Context) *MapValueIterator {
	iter := &MapValueIterator{
		ctx:  ctx,
		next: make(chan struct{}, 1),
		rc:   make(chan struct{}, 1),
		m:    x,
	}
	go iter.run()

	return iter
}

// Iter creates a new iterator for the map value.
func (x *MapValue) Iter() *MapValueIterator {
	return x.IterCtx(context.Background())
}

// MapValueIterator is an iterator for map values.
type MapValueIterator struct {
	m        *MapValue
	cur      MapValueKV
	ctx      context.Context
	next     chan struct{}
	rc       chan struct{}
	finished bool
}

func (x *MapValueIterator) run() {
	x.m.btree.Ascend(func(i btree.Item) bool {
		defer func() { x.rc <- struct{}{} }()
		select {
		case <-x.ctx.Done():
			return false
		case <-x.next:
		}
		x.cur = MapValueKV{
			Key:   i.(*mapValueKV).Key,
			Value: i.(*mapValueKV).Value,
		}
		return true
	})
	x.finished = true
}

// Next checks if there is
func (x *MapValueIterator) Next() bool {
	if x.finished {
		return false
	}
	x.next <- struct{}{}
	<-x.rc
	return true
}

// Key returns the key for the current kv pair.
func (x *MapValueIterator) Key() Value {
	return x.cur.Key
}

// Value returns the value for the current KV pair.
func (x *MapValueIterator) Value() Value {
	return x.cur.Value
}

// Type returns the type of the value.
// Implements the Value interface.
func (x *MapValue) Type() bsttype.Type {
	return x.MapType
}

// Kind returns the basic kind of the value.
// Implements the Value interface.
func (x *MapValue) Kind() bsttype.Kind {
	return bsttype.KindMap
}

// Skip skips the value in the map entries.
// Implements the Value interface.
func (x *MapValue) Skip(rs io.ReadSeeker, options bstio.ValueOptions) (int64, error) {
	return bstskip.SkipMap(rs, x.MapType, options)
}

// MarshalValue marshals the value to the writer.
// Implements the Value interface.
func (x *MapValue) MarshalValue(options bstio.ValueOptions) ([]byte, error) {
	buf := &bytes.Buffer{}
	_, err := x.WriteValue(buf, options)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalValue unmarshals the value from the reader.
// Implements the Value interface.
func (x *MapValue) UnmarshalValue(in []byte, options bstio.ValueOptions) error {
	br := bytes.NewReader(in)
	_, err := x.ReadValue(br, options)
	if err != nil {
		return err
	}
	return nil
}

// ReadValue reads the value from the reader.
// Implements the Value interface.
func (x *MapValue) ReadValue(r io.Reader, options bstio.ValueOptions) (int, error) {
	// 1. Read the number of entries.
	length, lt, err := bstio.ReadUint(r, options.Descending)
	if err != nil {
		return lt, err
	}

	bytesRead := lt

	x.btree = btree.New(2)

	// 2. Prepare the key reader and key, value options.
	ko := bstio.ValueOptions{Descending: x.MapType.Key.Descending}
	if options.Descending {
		ko.Descending = !ko.Descending
	}
	kr := newKeyReader(r, x.MapType.Key.Descending)

	vo := bstio.ValueOptions{Descending: x.MapType.Value.Descending}
	if options.Descending {
		vo.Descending = !vo.Descending
	}

	// 2. Read the entries.
	var n int
	for i := uint(0); i < length; i++ {
		// 2.1. Read the key.
		key := EmptyValueOf(x.MapType.Key.Type)

		kr.start()

		n, err = key.ReadValue(kr, ko)
		if err != nil {
			return n, err
		}
		bytesRead += n
		kr.stop()

		// 2.2. Read the value.
		value := EmptyValueOf(x.MapType.Value.Type)
		n, err = value.ReadValue(kr, vo)
		if err != nil {
			return n, err
		}
		bytesRead += n

		// 2.3. Add the key value pair to the map.
		item := &mapValueKV{
			kb:        kr.key(),
			keyDesc:   x.MapType.Key.Descending,
			valueDesc: x.MapType.Value.Descending,
			Key:       key,
			Value:     value,
		}

		x.btree.ReplaceOrInsert(item)
	}
	return bytesRead, nil
}

// WriteValue writes the value to the writer.
// Implements the Value interface.
func (x *MapValue) WriteValue(w io.Writer, options bstio.ValueOptions) (int, error) {
	// 1. Write the number of entries.
	total, err := bstio.WriteUint(w, uint(x.btree.Len()), options.Descending)
	if err != nil {
		return total, bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write map length")
	}

	// 2. Iterate over the map entries and write each entry.
	x.btree.Ascend(func(i btree.Item) bool {
		// 2.1. Write the key.
		var n int
		kv := i.(*mapValueKV)
		n, err = kv.Key.WriteValue(w, options)
		if err != nil {
			err = bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write map key")
			return false
		}
		total += n

		// 2.2. Write the value.
		n, err = kv.Value.WriteValue(w, options)
		if err != nil {
			err = bsterr.ErrWrap(err, bsterr.CodeEncodingBinaryValue, "failed to write map value")
			return false
		}
		total += n
		return true
	})
	if err != nil {
		return total, err
	}
	return total, nil
}

// Len returns the number of entries in the map.
func (x *MapValue) Len() int {
	return x.btree.Len()
}

// Less implements btree.Item.
func (m *mapValueKV) Less(than btree.Item) bool {
	cmp := bytes.Compare(m.kb, than.(*mapValueKV).kb)
	if m.keyDesc {
		return cmp > 0
	}
	return cmp < 0
}

var _ io.Reader = (*keyReader)(nil)

// keyReader is the reader specialized for partial reads.
// It is used for the map key reads.
type keyReader struct {
	r   io.Reader
	buf *bytes.Buffer

	wk   bool
	desc bool
}

func newKeyReader(r io.Reader, desc bool) *keyReader {
	return &keyReader{
		r:    r,
		buf:  &bytes.Buffer{},
		desc: desc,
	}
}

// Read implements io.Reader.
func (k *keyReader) Read(p []byte) (n int, err error) {
	n, err = k.r.Read(p)
	if err != nil {
		return n, err
	}

	if k.wk {
		cp := p
		if k.desc {
			cp = make([]byte, n)
			copy(cp, p)
			bstio.ReverseBytes(cp)
		}
		k.buf.Write(cp)
	}
	return n, nil
}

// ReadByte implements io.ByteReader.
func (k *keyReader) ReadByte() (byte, error) {
	b, err := bstio.ReadByte(k.r)
	if err != nil {
		return b, err
	}

	if k.wk {
		bt := b
		if k.desc {
			bt = ^b
		}
		k.buf.WriteByte(bt)
	}
	return b, nil
}

func (k *keyReader) start() {
	k.wk = true
}

func (k *keyReader) stop() {
	k.wk = false
}

func (k *keyReader) key() []byte {
	key := make([]byte, k.buf.Len())
	_, _ = k.buf.Read(key)
	k.buf.Reset()
	return key
}
