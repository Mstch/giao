package buf

import (
	"sync/atomic"
	"unsafe"
)

const ImpossibleStat = int64(1 + 1<<32)

type accessor [2]int32

func (a *accessor) access() int {
	for {
		stat := atomic.LoadInt64((*int64)(unsafe.Pointer(a)))
		if ImpossibleStat == stat {
			panic("impossible accessors stat")
		}
		if int32(stat) == 0 {
			if atomic.CompareAndSwapInt32(&(a[0]), 0, 1) {
				return 0
			}
		} else if stat>>32 == 0 {
			if atomic.CompareAndSwapInt32(&(a[1]), 0, 1) {
				return 1
			}
		} else {
			panic("impossible accessors stat!")
		}
	}
}

func (a *accessor) tryAccess(i int) bool {
	stat := atomic.LoadInt64((*int64)(unsafe.Pointer(a)))
	if ImpossibleStat == stat {
		panic("impossible accessors stat")
	}
	return atomic.CompareAndSwapInt32(&a[i], 0, 1)
}

func (a *accessor) release(i int) {
	if atomic.CompareAndSwapInt32(&a[i], 1, 0) {
		return
	} else {
		panic("cas accessStat failed when release")
	}
}
