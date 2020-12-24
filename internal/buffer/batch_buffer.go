// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buffer

// Simple byte buffer for marshaling data. //
import (
	"encoding/binary"
	"errors"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/common"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// smallBufferSize is an initial allocation minimal capacity.
const smallBufferSize = 64
const FlushSize = 512 * 1024
const ForceFlushInterval = 1 * time.Millisecond

// A MsgBuffer is a variable-sized buffer of bytes with Read and Write methods.
// The zero value for MsgBuffer is an empty buffer ready to use.
type MsgBuffer struct {
	lock      *sync.Mutex
	lastFlush int64
	buf       []byte // contents are the bytes buf[off : len(buf)]
	off       int    // read at &buf[off], write at &buf[len(buf)]
	lastRead  readOp // last read operation, so that Unread* can work correctly.
}

// The readOp constants describe the last action performed on
// the buffer, so that UnreadRune and UnreadByte can check for
// invalid usage. opReadRuneX constants are chosen such that
// converted to int they correspond to the rune size that was read.
type readOp int8

// Don't use iota for these, as the values need to correspond with the
// names and comments, which is easier to see when being explicit.
const (
	opRead      readOp = -1 // Any other read operation.
	opInvalid   readOp = 0  // Non-read operation.
	opReadRune1 readOp = 1  // Read rune of size 1.
	opReadRune2 readOp = 2  // Read rune of size 2.
	opReadRune3 readOp = 3  // Read rune of size 3.
	opReadRune4 readOp = 4  // Read rune of size 4.
)

// ErrTooLarge is passed to panic if memory cannot be allocated to store data in a buffer.
var ErrTooLarge = errors.New("bytes.MsgBuffer: too large")
var errNegativeRead = errors.New("bytes.MsgBuffer: reader returned negative count from Read")

const maxInt = int(^uint(0) >> 1)

// empty reports whether the unread portion of the buffer is empty.
func (b *MsgBuffer) empty() bool { return len(b.buf) <= b.off }

// Len returns the number of bytes of the unread portion of the buffer;
// b.Len() == len(b.Bytes()).
func (b *MsgBuffer) Len() int { return len(b.buf) - b.off }

// Cap returns the capacity of the buffer's underlying byte slice, that is, the
// total space allocated for the buffer's data.
func (b *MsgBuffer) Cap() int { return cap(b.buf) }

// Truncate discards all but the first n unread bytes from the buffer
// but continues to use the same allocated storage.
// It panics if n is negative or greater than the length of the buffer.
func (b *MsgBuffer) Truncate(n int) {
	if n == 0 {
		b.Reset()
		return
	}
	b.lastRead = opInvalid
	if n < 0 || n > b.Len() {
		panic("bytes.MsgBuffer: truncation out of range")
	}
	b.buf = b.buf[:b.off+n]
}

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
// Reset is the same as Truncate(0).
func (b *MsgBuffer) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
	b.lastRead = opInvalid
}

// tryGrowByReslice is a inlineable version of grow for the fast-case where the
// internal buffer only needs to be resliced.
// It returns the index where bytes should be written and whether it succeeded.
func (b *MsgBuffer) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

// grow grows the buffer to guarantee space for n more bytes.
// It returns the index where bytes should be written.
// If the buffer can't grow it will panic with ErrTooLarge.
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

// Grow grows the buffer's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to the
// buffer without another allocation.
// If n is negative, Grow will panic.
// If the buffer can't grow it will panic with ErrTooLarge.
func (b *MsgBuffer) Grow(n int) {
	if n < 0 {
		panic("bytes.MsgBuffer.Grow: negative count")
	}
	m := b.grow(n)
	b.buf = b.buf[:m]
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (b *MsgBuffer) Write(p []byte) (n int, err error) {
	b.lastRead = opInvalid
	m, ok := b.tryGrowByReslice(len(p))
	if !ok {
		m = b.grow(len(p))
	}
	return copy(b.buf[m:], p), nil
}

func (b *MsgBuffer) WriteMsg(handlerId int, msg giao.Msg) (n int, err error) {
	b.lastRead = opInvalid
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

// WriteString appends the contents of s to the buffer, growing the buffer as
// needed. The return value n is the length of s; err is always nil. If the
// buffer becomes too large, WriteString will panic with ErrTooLarge.
func (b *MsgBuffer) WriteString(s string) (n int, err error) {
	b.lastRead = opInvalid
	m, ok := b.tryGrowByReslice(len(s))
	if !ok {
		m = b.grow(len(s))
	}
	return copy(b.buf[m:], s), nil
}

// MinRead is the minimum slice size passed to a Read call by
// MsgBuffer.ReadFrom. As long as the MsgBuffer has at least MinRead bytes beyond
// what is required to hold the contents of r, ReadFrom will not grow the
// underlying buffer.
const MinRead = 512

// ReadFrom reads data from r until EOF and appends it to the buffer, growing
// the buffer as needed. The return value n is the number of bytes read. Any
// error except io.EOF encountered during the read is also returned. If the
// buffer becomes too large, ReadFrom will panic with ErrTooLarge.
func (b *MsgBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	b.lastRead = opInvalid
	for {
		i := b.grow(MinRead)
		b.buf = b.buf[:i]
		m, e := r.Read(b.buf[i:cap(b.buf)])
		if m < 0 {
			panic(errNegativeRead)
		}

		b.buf = b.buf[:i+m]
		n += int64(m)
		if e == io.EOF {
			return n, nil // e is EOF, so return nil explicitly
		}
		if e != nil {
			return n, e
		}
	}
}

// makeSlice allocates a slice of size n. If the allocation fails, it panics
// with ErrTooLarge.
func makeSlice(n int) []byte {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	return make([]byte, n)
}

// WriteByte appends the byte c to the buffer, growing the buffer as needed.
// The returned error is always nil, but is included to match bufio.Writer's
// WriteByte. If the buffer becomes too large, WriteByte will panic with
// ErrTooLarge.
func (b *MsgBuffer) WriteByte(c byte) error {
	b.lastRead = opInvalid
	m, ok := b.tryGrowByReslice(1)
	if !ok {
		m = b.grow(1)
	}
	b.buf[m] = c
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to the
// buffer, returning its length and an error, which is always nil but is
// included to match bufio.Writer's WriteRune. The buffer is grown as needed;
// if it becomes too large, WriteRune will panic with ErrTooLarge.
func (b *MsgBuffer) WriteRune(r rune) (n int, err error) {
	if r < utf8.RuneSelf {
		b.WriteByte(byte(r))
		return 1, nil
	}
	b.lastRead = opInvalid
	m, ok := b.tryGrowByReslice(utf8.UTFMax)
	if !ok {
		m = b.grow(utf8.UTFMax)
	}
	n = utf8.EncodeRune(b.buf[m:m+utf8.UTFMax], r)
	b.buf = b.buf[:m+n]
	return n, nil
}

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained. The return value n is the number of bytes read. If the
// buffer has no data to return, err is io.EOF (unless len(p) is zero);
// otherwise it is nil.
func (b *MsgBuffer) Read(p []byte) (n int, err error) {
	b.lastRead = opInvalid
	if b.empty() {
		// MsgBuffer is empty, reset to recover space.
		b.Reset()
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.off:])
	b.off += n
	if n > 0 {
		b.lastRead = opRead
	}
	return n, nil
}

// Next returns a slice containing the next n bytes from the buffer,
// advancing the buffer as if the bytes had been returned by Read.
// If there are fewer than n bytes in the buffer, Next returns the entire buffer.
// The slice is only valid until the next call to a read or write method.
func (b *MsgBuffer) Next(n int) []byte {
	b.lastRead = opInvalid
	m := b.Len()
	if n > m {
		n = m
	}
	data := b.buf[b.off : b.off+n]
	b.off += n
	if n > 0 {
		b.lastRead = opRead
	}
	return data
}

// ReadByte reads and returns the next byte from the buffer.
// If no byte is available, it returns error io.EOF.
func (b *MsgBuffer) ReadByte() (byte, error) {
	if b.empty() {
		// MsgBuffer is empty, reset to recover space.
		b.Reset()
		return 0, io.EOF
	}
	c := b.buf[b.off]
	b.off++
	b.lastRead = opRead
	return c, nil
}

// ReadRune reads and returns the next UTF-8-encoded
// Unicode code point from the buffer.
// If no bytes are available, the error returned is io.EOF.
// If the bytes are an erroneous UTF-8 encoding, it
// consumes one byte and returns U+FFFD, 1.
func (b *MsgBuffer) ReadRune() (r rune, size int, err error) {
	if b.empty() {
		// MsgBuffer is empty, reset to recover space.
		b.Reset()
		return 0, 0, io.EOF
	}
	c := b.buf[b.off]
	if c < utf8.RuneSelf {
		b.off++
		b.lastRead = opReadRune1
		return rune(c), 1, nil
	}
	r, n := utf8.DecodeRune(b.buf[b.off:])
	b.off += n
	b.lastRead = readOp(n)
	return r, n, nil
}

// UnreadRune unreads the last rune returned by ReadRune.
// If the most recent read or write operation on the buffer was
// not a successful ReadRune, UnreadRune returns an error.  (In this regard
// it is stricter than UnreadByte, which will unread the last byte
// from any read operation.)
func (b *MsgBuffer) UnreadRune() error {
	if b.lastRead <= opInvalid {
		return errors.New("bytes.MsgBuffer: UnreadRune: previous operation was not a successful ReadRune")
	}
	if b.off >= int(b.lastRead) {
		b.off -= int(b.lastRead)
	}
	b.lastRead = opInvalid
	return nil
}

var errUnreadByte = errors.New("bytes.MsgBuffer: UnreadByte: previous operation was not a successful read")

// UnreadByte unreads the last byte returned by the most recent successful
// read operation that read at least one byte. If a write has happened since
// the last read, if the last read returned an error, or if the read read zero
// bytes, UnreadByte returns an error.
func (b *MsgBuffer) UnreadByte() error {
	if b.lastRead == opInvalid {
		return errUnreadByte
	}
	b.lastRead = opInvalid
	if b.off > 0 {
		b.off--
	}
	return nil
}

func (b *MsgBuffer) Flush(writer io.Writer) error {
	b.lock.Lock()
	if l := b.Len(); l > 0 {
		atomic.StoreInt64(&b.lastFlush, time.Now().UnixNano()/1e6)
		buf := make([]byte, l)
		copy(buf, b.Next(l))
		b.lock.Unlock()
		_, err := writer.Write(buf)
		if err != nil {
			return err
		}
	} else {
		b.lock.Unlock()
	}
	return nil
}

// NewBuffer creates and initializes a new MsgBuffer using buf as its
// initial contents. The new MsgBuffer takes ownership of buf, and the
// caller should not use buf after this call. NewBuffer is intended to
// prepare a MsgBuffer to read existing data. It can also be used to set
// the initial size of the internal buffer for writing. To do that,
// buf should have the desired capacity but a length of zero.
//
// In most cases, new(MsgBuffer) (or just declaring a MsgBuffer variable) is
// sufficient to initialize a MsgBuffer.
func NewBuffer(buf []byte) *MsgBuffer { return &MsgBuffer{buf: buf} }

//getp->lock->write into MsgBuffer-> ( if MsgBuffer.Len >= FlushSize ->copy BytesToFlush and write to some Writer)->unlock
type BatchBuffer struct {
	mbufs  []*MsgBuffer
	Writer io.Writer
	stop   bool
}

func NewBatchBuffer(writer io.Writer) *BatchBuffer {
	bb := &BatchBuffer{
		Writer: writer,
	}
	mbufs := make([]*MsgBuffer, common.GoMaxProc)
	for i := range mbufs {
		mbufs[i] = &MsgBuffer{lastFlush: time.Now().UnixNano() / 1e6, lock: &sync.Mutex{}}
	}
	bb.mbufs = mbufs
	return bb
}
func (b *BatchBuffer) WriteMsg(handlerId int, msg giao.Msg) (n int, err error) {
	mbuf := b.mbufs[runtime.GetPid()]
	mbuf.lock.Lock()
	n, err = mbuf.WriteMsg(handlerId, msg)
	if err == nil {
		if l := mbuf.Len(); l >= FlushSize {
			atomic.StoreInt64(&mbuf.lastFlush, time.Now().UnixNano()/1e6)
			buf := make([]byte, l)
			copy(buf, mbuf.Next(l))
			mbuf.lock.Unlock()
			return b.Writer.Write(buf)
		}
	}
	mbuf.lock.Unlock()
	return
}

func (b *BatchBuffer) Stop() error {
	b.stop = true
	for i := 0; i < common.GoMaxProc; i++ {
		mbuf := b.mbufs[i]
		err := mbuf.Flush(b.Writer)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *BatchBuffer) StartFlushLooper() error {
	for !b.stop {
		<-time.After(ForceFlushInterval)
		for i := 0; i < common.GoMaxProc; i++ {
			now := time.Now().UnixNano() / 1e6
			mbuf := b.mbufs[i]
			if now >= atomic.LoadInt64(&mbuf.lastFlush)+ForceFlushInterval.Milliseconds() {
				err := mbuf.Flush(b.Writer)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
