package flushbuffer

import (
	"context"
	"encoding/binary"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/pool"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	_ "unsafe"
)

const FlushSize = 1024 * 1024

//The buffer size is fixed that when it is full will flush all data and discard this buf
type (
	FBuffers struct {
		flushLock    *sync.Mutex
		buffers      []*FBuffer
		flushSize    int
		flushTimer   *time.Timer
		flushTimeout time.Duration
		writer       io.Writer
		DoneCh       chan struct{}
		Len          uint64
	}
	FBuffer struct {
		/*
			只有flusher正在flush buf 时，有goroutine试图来write msg 才会导致buf的竞争。
			倘若buf0正在被flush则把msg写入buf1
			倘若buf1正在被flush则把msg写入buf0
			flusher每次flush都会flush buf0和buf1
		*/
		buf     [2][]byte
		optTime int64
		accessor *accessor
	}
)

func NewFBuffers(flushTimeout time.Duration, writer io.Writer) *FBuffers {
	buffers := make([]*FBuffer, runtime.GOMAXPROCS(0))
	for i := range buffers {
		buffers[i] = NewFBuffer(FlushSize)
	}
	return &FBuffers{buffers: buffers, flushSize: FlushSize, flushTimeout: flushTimeout, writer: writer, flushLock: &sync.Mutex{}, DoneCh: make(chan struct{})}

}

func NewFBuffer(flushSize int) *FBuffer {
	buf := [2][]byte{}
	buf[0] = make([]byte, 0, flushSize)
	buf[1] = make([]byte, 0, flushSize)
	return &FBuffer{
		buf:      buf,
		accessor: &accessor{},
	}
}

func (fbs *FBuffers) Write(msg giao.Msg, id int) error {
	pid := pin()
	fb := fbs.buffers[pid] //找到线程私有的f-buffer
	if fbs.flushSize < msg.Size()+8 {
		panic("panic err too large msg")
	}
	atomic.StoreInt64(&fb.optTime, time.Now().UnixNano())
	size := msg.Size()
	bufI := fb.accessor.access()
	buf := fb.buf[bufI]
	l := len(buf)
	if cap(buf)-l >= size+8 {
		binary.BigEndian.PutUint32(buf[l:l+4], uint32(id))
		binary.BigEndian.PutUint32(buf[l+4:l+8], uint32(size))
		_, err := msg.MarshalTo(buf[l+8 : l+8+size])
		fb.buf[bufI] = fb.buf[bufI][:l+8+size]
		fb.accessor.release(bufI)
		unpin()
		return err
	}
	flushableBuf := buf[:l]
	//write msg into new buf
	fb.buf[bufI] = pool.GetBytes(FlushSize)
	binary.BigEndian.PutUint32(fb.buf[bufI][:4], uint32(id))
	binary.BigEndian.PutUint32(fb.buf[bufI][4:8], uint32(size))
	_, err := msg.MarshalTo(fb.buf[bufI][8 : 8+size])
	fb.buf[bufI] = fb.buf[bufI][:8+size]
	fb.accessor.release(bufI)
	unpin()
	if err != nil {
		return err
	}
	//write old buf into tcp conn
	_, err = fbs.writer.Write(flushableBuf)
	pool.PutBytes(flushableBuf)
	return err
}

func (fbs *FBuffers) StartFlusher(sessionCtx context.Context) error {
	fctx := context.WithValue(sessionCtx, "name", "flusher")
	fbs.flushTimer = time.NewTimer(fbs.flushTimeout)
	for {
		select {
		case t := <-fbs.flushTimer.C:
			if err := fbs.flush(t, false); err != nil {
				return err
			}
			fbs.flushTimer.Reset(fbs.flushTimeout)
		case <-fctx.Done():
			return nil
		}
	}
}
func (fbs *FBuffers) flush(t time.Time, force bool) error {
	fbs.flushLock.Lock()
	defer fbs.flushLock.Unlock()
	for _, buffer := range fbs.buffers {
		if force || t.UnixNano()-atomic.LoadInt64(&buffer.optTime) > int64(fbs.flushTimeout) {
			if ok := buffer.accessor.tryAccess(0); ok {
				buf := buffer.buf[0]
				if len(buf) > 0 {
					_, err := fbs.writer.Write(buf)
					buffer.buf[0] = buffer.buf[0][:0]
					if err != nil {
						return err
					}
				}
				buffer.accessor.release(0)
			} else if force { //when force flush,must promise other writer is done
				panic("force flush when other writer undone")
			}
			if ok := buffer.accessor.tryAccess(1); ok {
				buf := buffer.buf[1]
				if len(buf) > 0 {
					_, err := fbs.writer.Write(buf)
					buffer.buf[1] = buffer.buf[1][:0]
					if err != nil {
						return err
					}
				}
				buffer.accessor.release(1)
			} else if force { //when force flush,must promise other writer is done
				panic("force flush when other writer undone")
			}
		}
	}
	return nil
}

func (fbs *FBuffers) ForceFlush() error {
	return fbs.flush(time.Time{}, true)
}

func (fbs *FBuffers) StopFlushTimer() error {
	if !fbs.flushTimer.Stop() {
		t := <-fbs.flushTimer.C
		return fbs.flush(t, true)
	}
	return nil
}

//go:linkname pin runtime.procPin
func pin() int

//go:linkname unpin runtime.procUnpin
func unpin() int
