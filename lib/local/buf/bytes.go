package buf

import (
	"context"
	"errors"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/lib/local"
	"github.com/Mstch/giao/lib/local/pool"
	"github.com/Mstch/giao/lib/lock"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var ErrTooLarge = errors.New("too large")

type (
	Statics struct {
		needFlush  uint64
		flushFlush uint64
		writeFlush uint64
	}
	LocalBytesBuf struct {
		// originally it is a buf like full[0] but take it out alone now
		buf []byte
		// to hold full buf when other routine flushing
		full      [][]byte
		flushLock *lock.Lock
		id        int
	}
	BytesBuffer struct {
		accessors     []*accessor
		lbbs          [][2]*LocalBytesBuf
		flushLock     *lock.Lock
		writer        io.Writer
		flushDuration time.Duration
		timer         *time.Timer
		ctx           context.Context
		cancel        context.CancelFunc
		flushDone     chan struct{}
		cacheFullSize int
		fullPool      *sync.Pool
		statics       Statics
	}
)

func NewBytesBuf(size int, flushDuration time.Duration, writer io.Writer, ctx context.Context) *BytesBuffer {
	maxp := runtime.GOMAXPROCS(0)
	lbbs := make([][2]*LocalBytesBuf, maxp)
	accessors := make([]*accessor, maxp)
	cacheFullSize := 16
	cctx, cancel := context.WithCancel(ctx)
	l := &lock.Lock{}
	for i := 0; i < maxp; i++ {
		lbbs[i] = [2]*LocalBytesBuf{{
			buf:       make([]byte, 0, size),
			full:      make([][]byte, 0, cacheFullSize),
			flushLock: l,
			id:        i,
		}, {
			buf:       make([]byte, 0, size),
			full:      make([][]byte, 0, cacheFullSize),
			flushLock: l,
			id:        i,
		}}
		accessors[i] = &accessor{}

	}
	return &BytesBuffer{
		accessors:     accessors,
		lbbs:          lbbs,
		flushLock:     l,
		writer:        writer,
		flushDuration: flushDuration,
		timer:         time.NewTimer(time.Hour),
		cacheFullSize: cacheFullSize,
		flushDone:     make(chan struct{}, 1),
		fullPool: &sync.Pool{New: func() interface{} {
			return make([][]byte, cacheFullSize)
		}},
		ctx:    cctx,
		cancel: cancel,
	}
}

func (bb *BytesBuffer) StartFlush() error {
	for {
		select {
		case <-bb.timer.C:
			var lbb *LocalBytesBuf
			allEmpty := true
			for i, a := range bb.accessors {
				lbb = bb.lbbs[i][0]
				if lbb.flushLock.TryLock() {
					if a.tryAccess(0) {
						wl, err := bb.flush(bb.lbbs[i][0].buf, bb.lbbs[i][0].full)
						atomic.AddUint64(&bb.statics.flushFlush, uint64(wl))
						lbb.buf = lbb.buf[:0]
						for _, fullBuff := range lbb.full {
							pool.PutBytes(fullBuff)
						}
						lbb.full = lbb.full[:0]
						a.release(0)
						if err != nil {
							return err
						}
						allEmpty = allEmpty && (wl == 0)
					}
					lbb.flushLock.Unlock()
				}
				lbb = bb.lbbs[i][1]
				if lbb.flushLock.TryLock() {
					if a.tryAccess(1) {
						wl, err := bb.flush(bb.lbbs[i][1].buf, bb.lbbs[i][1].full)
						atomic.AddUint64(&bb.statics.flushFlush, uint64(wl))
						lbb.buf = lbb.buf[:0]
						for _, fullBuff := range lbb.full {
							pool.PutBytes(fullBuff)
						}
						lbb.full = lbb.full[:0]
						a.release(1)
						if err != nil {
							return err
						}
						allEmpty = allEmpty && (wl == 0)
					}
					lbb.flushLock.Unlock()
				}
			}
			bb.timer.Reset(bb.flushDuration)
			// todo use allEmpty to slow down bb.timer
		case <-bb.ctx.Done():
			bb.flushDone <- struct{}{}
			return nil
		}
	}
}

func (bb *BytesBuffer) Shutdown() {
	bb.cancel()
	<-bb.flushDone
}

func (bb *BytesBuffer) ForceFlush() error {
	bb.timer.Reset(bb.flushDuration)
	var lbb *LocalBytesBuf
	for i, a := range bb.accessors {
		lbb = bb.lbbs[i][0]
		if lbb.flushLock.TryLock() && a.tryAccess(0) {
			wl, err := bb.flush(bb.lbbs[i][0].buf, bb.lbbs[i][0].full)
			atomic.AddUint64(&bb.statics.flushFlush, uint64(wl))
			lbb.flushLock.Unlock()
			lbb.buf = lbb.buf[:0]
			for _, fullBuff := range lbb.full {
				pool.PutBytes(fullBuff)
			}
			lbb.full = lbb.full[:0]
			a.release(0)
			if err != nil {
				return err
			}
		}
		lbb = bb.lbbs[i][1]
		if lbb.flushLock.TryLock() && a.tryAccess(1) {
			wl, err := bb.flush(bb.lbbs[i][1].buf, bb.lbbs[i][1].full)
			atomic.AddUint64(&bb.statics.flushFlush, uint64(wl))
			lbb.flushLock.Unlock()
			lbb.buf = lbb.buf[:0]
			for _, fullBuff := range lbb.full {
				pool.PutBytes(fullBuff)
			}
			lbb.full = lbb.full[:0]
			a.release(1)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (bb *BytesBuffer) flush(flushBuff []byte, flushBuffs [][]byte) (int, error) {

	if len(flushBuff) > 0 {
		if len(flushBuffs) == 0 {
			return bb.writer.Write(flushBuff)
		}
		flushBuffs = append(flushBuffs, flushBuff)
	}
	if len(flushBuffs) > 0 {
		tmpBuffs := make([][]byte, len(flushBuffs))
		for i, buff := range flushBuffs {
			tmpBuffs[i] = buff
		}
		wl, err := (*net.Buffers)(&tmpBuffs).WriteTo(bb.writer)
		if err != nil {
			return 0, err
		}
		return int(wl), err
	}

	return 0, nil
}

func (bb *BytesBuffer) Write(prefix []byte, obj giao.Msg) error {
	atomic.AddUint64(&bb.statics.needFlush, uint64(len(prefix)+obj.Size()))
	bb.timer.Reset(bb.flushDuration)
	pid := local.Pin()
	aid := bb.accessors[pid].access()
	lbb := bb.lbbs[pid][aid]
	flushBuff, err := lbb.Write(prefix, obj)
	flush := false
	forceFlush := false
	var flushBuffs [][]byte
	if len(flushBuff) > 0 || len(lbb.full) > 0 {
		if bb.flushLock.TryLock() {
			flush = true
			if len(lbb.full) > 0 {
				flushBuffs = lbb.full
				lbb.full = make([][]byte, 0, bb.cacheFullSize)
			}
		} else if len(flushBuff) > 0 {
			if len(lbb.full) >= bb.cacheFullSize-1 {
				flushBuffs = lbb.full
				forceFlush = true
				lbb.full = make([][]byte, 0, bb.cacheFullSize)
			} else {
				lbb.full = append(lbb.full, flushBuff)
			}
		}
	}

	bb.accessors[pid].release(aid)
	local.Unpin()
	if err != nil {
		return err
	}
	var wl int
	if flush || forceFlush {
		wl, err = bb.flush(flushBuff, flushBuffs)
		atomic.AddUint64(&bb.statics.writeFlush, uint64(wl))
		if !forceFlush {
			bb.flushLock.Unlock()
		}
		if len(flushBuff) > 0 && len(flushBuffs) == 0 {
			pool.PutBytes(flushBuff)
		}
		for _, buff := range flushBuffs {
			if len(buff) > 0 {
				pool.PutBytes(buff)
			}
		}
	}

	return err
}

func (lbb *LocalBytesBuf) Write(header []byte, obj giao.Msg) (flushBuf []byte, err error) {
	bl := len(lbb.buf)
	bc := cap(lbb.buf)
	hl := len(header)
	need := hl + obj.Size()
	if need > bc {
		return nil, ErrTooLarge
	}
	var appendBuf []byte
	if bl+need < bc {
		lbb.buf = lbb.buf[:bl+need]
		appendBuf = lbb.buf[bl : bl+need]
	} else if bl+need == bc {
		appendBuf = lbb.buf[bl:bc]
		flushBuf = lbb.buf[:bc]
		lbb.buf = pool.GetBytes(bc)[:0]
	} else {
		flushBuf = lbb.buf
		lbb.buf = pool.GetBytes(bc)[:need]
		appendBuf = lbb.buf
	}
	copy(appendBuf, header)
	_, err = obj.MarshalTo(appendBuf[hl:])
	return
}
