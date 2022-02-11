package lock

import (
	"sync/atomic"
)

type Lock struct {
	locked int32
}

func (l *Lock) TryLock() bool {
	return atomic.CompareAndSwapInt32(&l.locked, 0, 1)
}

func (l *Lock) Unlock() bool {
	return atomic.CompareAndSwapInt32(&l.locked, 1, 0)
}
