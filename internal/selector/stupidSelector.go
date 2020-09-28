package selector

import (
	"github.com/Mstch/giao"
	"sync/atomic"
	_ "unsafe"
)

type BaseSelector struct {
	sessions []giao.Session
}

func (bs *BaseSelector) AddSession(session giao.Session) {
	bs.sessions = append(bs.sessions, session)
}
func (bs *BaseSelector) GetAllSession() []giao.Session {
	return bs.sessions
}

type RoundRobinSelector struct {
	BaseSelector
	next uint32
}

func (rr *RoundRobinSelector) Select() giao.Session {
	return rr.sessions[int(atomic.AddUint32(&rr.next, 1))%len(rr.sessions)]
}

type RandomSelector struct {
	BaseSelector
}

func (rd *RandomSelector) Select() giao.Session {
	return rd.sessions[int(runtime_fastrand())%len(rd.sessions)]
}

//go:linkname runtime_fastrand runtime.fastrand
func runtime_fastrand() uint32
