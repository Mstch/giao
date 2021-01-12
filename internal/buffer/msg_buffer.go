package buffer

import (
	"encoding/binary"
	"errors"
	"github.com/Mstch/giao"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type bufStat int32

const (
	idle bufStat = iota
	working
)

type MsgBuffer struct {
	stat      bufStat
	lock      *sync.Mutex
	lastWrite int64

	buf []byte // contents are the bytes buf[off : len(buf)]

	off int // read at &buf[off], write at &buf[len(buf)]
}

var ErrTooLarge = errors.New("bytes.MsgBuffer: too large")

const maxInt = int(^uint(0) >> 1)

func (b *MsgBuffer) empty() bool { return len(b.buf) <= b.off }
func (b *MsgBuffer) Len() int    { return len(b.buf) - b.off }
func (b *MsgBuffer) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
}
func (b *MsgBuffer) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}
func (b *MsgBuffer) grow(n int) int {
	m := b.Len()
	// If buffer is empty, reset to recover space.
	if m == 0 && b.off != 0 {
		b.Reset()
	}
	// Try to grow by means of a reslice.
	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}
	if b.buf == nil && n <= smallBufferSize {
		b.buf = make([]byte, n, smallBufferSize)
		return 0
	}
	c := cap(b.buf)
	if n <= c/2-m {
		// We can slide things down instead of allocating a new
		// slice. We only need m+n <= c to slide, but
		// we instead let capacity get twice as large so we
		// don't spend all our time copying.
		copy(b.buf, b.buf[b.off:])
	} else if c > maxInt-c-n {
		panic(ErrTooLarge)
	} else {
		// Not enough space anywhere, we need to allocate.
		buf := makeSlice(2*c + n)
		copy(buf, b.buf[b.off:])
		b.buf = buf
	}
	// Restore b.off and len(b.buf).
	b.off = 0
	b.buf = b.buf[:m+n]
	return m
}
func (b *MsgBuffer) Grow(n int) {
	if n < 0 {
		panic("bytes.MsgBuffer.Grow: negative count")
	}
	m := b.grow(n)
	b.buf = b.buf[:m]
}
func (b *MsgBuffer) WriteMsg(handlerId int, msg giao.Msg) (n int, err error) {

	m, ok := b.tryGrowByReslice(4)
	if !ok {
		m = b.grow(4)
	}
	binary.BigEndian.PutUint32(b.buf[m:], uint32(handlerId))
	m, ok = b.tryGrowByReslice(4)
	if !ok {
		m = b.grow(4)
	}
	binary.BigEndian.PutUint32(b.buf[m:], uint32(msg.Size()))
	size := msg.Size()
	m, ok = b.tryGrowByReslice(size)
	if !ok {
		m = b.grow(size)
	}
	return msg.MarshalTo(b.buf[m:])
}
func makeSlice(n int) []byte {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	return make([]byte, n)
}
func (b *MsgBuffer) WriteTo(w io.Writer) (n int, err error) {

	if nBytes := b.Len(); nBytes > 0 {
		m, e := w.Write(b.buf[b.off:])
		if m > nBytes {
			panic("bytes.Buffer.WriteTo: invalid Write count")
		}
		b.off += m
		n = m
		if e != nil {
			return n, e
		}
		// all bytes should have been written, by definition of
		// Write method in io.Writer
		if m != nBytes {
			return n, io.ErrShortWrite
		}
	}
	// Buffer is now empty; reset.
	b.Reset()
	return n, nil
}
func (b *MsgBuffer) Flush(writer io.Writer) (n int, err error) {
	b.lock.LockWithCompetitor("flush")
	if l := b.Len(); l > 0 {
		atomic.StoreInt64(&b.lastWrite, time.Now().UnixNano()/1e6)
		n, err = b.WriteTo(writer)
		b.lock.Unlock()
		if err != nil {
			return
		}
	} else {
		b.lock.Unlock()
	}
	return
}
func NewBuffer(buf []byte) *MsgBuffer { return &MsgBuffer{buf: buf} }
