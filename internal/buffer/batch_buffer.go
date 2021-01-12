package buffer

import (
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/common"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const smallBufferSize = 64
const FlushSize = 512 * 1024
const ForceFlushInterval = 1 * time.Millisecond

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
		mbufs[i] = &MsgBuffer{lastWrite: time.Now().UnixNano() / 1e6, lock: &sync.Mutex{}}
	}
	bb.mbufs = mbufs
	return bb
}
func (b *BatchBuffer) WriteMsg(handlerId int, msg giao.Msg) (n int, err error) {
	mbuf := b.mbufs[runtime.GetPid()]
	atomic.StoreInt64(&mbuf.lastWrite, time.Now().UnixNano()/1e6)
	mbuf.lock.LockWithCompetitor("write")
	n, err = mbuf.WriteMsg(handlerId, msg)
	if err == nil {
		if l := mbuf.Len(); l >= FlushSize {
			n, err = mbuf.WriteTo(b.Writer)
			mbuf.lock.Unlock()
			return
		}
	}
	mbuf.lock.Unlock()
	return
}

func (b *BatchBuffer) Stop() error {
	b.stop = true
	for i := 0; i < common.GoMaxProc; i++ {
		mbuf := b.mbufs[i]
		_, err := mbuf.Flush(b.Writer)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *BatchBuffer) StartFlushLooper() error {
	ticker := time.NewTicker(ForceFlushInterval)
	for !b.stop {
		<-ticker.C
		for i := 0; i < common.GoMaxProc; i++ {
			now := time.Now().UnixNano() / 1e6
			mbuf := b.mbufs[i]
			if now >= atomic.LoadInt64(&mbuf.lastWrite)+ForceFlushInterval.Milliseconds() {
				_, err := mbuf.Flush(b.Writer)
				if err != nil {
					return err
				}
			}
		}
	}
	ticker.Stop()
	return nil
}
