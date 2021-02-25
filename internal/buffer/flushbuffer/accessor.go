package flushbuffer

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type accessor struct {
	stat0 int32
	stat1 int32
}

func (a *accessor) access() int {
	for {
		oldStat := atomic.LoadInt64((*int64)(unsafe.Pointer(a)))
		stat0 := int32(oldStat)
		stat1 := int32(oldStat >> 32)
		if stat0 == 0 {
			if atomic.CompareAndSwapInt32(&(a.stat0), 0, 1) {
				return 0
			}
		} else if stat1 == 0 {
			if atomic.CompareAndSwapInt32(&(a.stat1), 0, 1) {
				return 1
			}
		} else {
			panic(fmt.Sprint("invalid fb.accessStat when access", oldStat))
		}
	}
}

func (a *accessor) tryAccess(i int) bool {
	oldStat := atomic.LoadInt64((*int64)(unsafe.Pointer(a)))
	stat0 := int32(oldStat)
	stat1 := int32(oldStat >> 32)
	if i == 0 {
		if stat0 == 0 {
			if atomic.CompareAndSwapInt32(&(a.stat0), 0, 1) {
				return true
			}
		}
	} else if i == 1 {
		if stat1 == 0 {
			if atomic.CompareAndSwapInt32(&(a.stat1), 0, 1) {
				return true
			}
		}
	} else {
		panic(fmt.Sprint("invalid fb.accessStat ", i, " when tryAccess ", oldStat))
	}
	return false
}

func (a *accessor) release(i int) {
	if i == 0 {
		if atomic.CompareAndSwapInt32(&(a.stat0), 1, 0) {
			return
		}
	}
	if i == 1 {
		if atomic.CompareAndSwapInt32(&(a.stat1), 1, 0) {
			return
		}
	}
	panic("cas fb.accessStat failed when release")
}
