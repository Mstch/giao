package session

import (
	"context"
	"github.com/Mstch/giao/lib/bytes"
	"github.com/Mstch/giao/lib/local/buf"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/Mstch/giao"
)

var lastSessionId = uint64(0)

type Session struct {
	net.Conn
	Id              uint64
	WriteBuffer     *buf.BytesBuffer
	Ctx             context.Context
	writeBufferSize int
	ReadDoneCh      chan struct{}
	errCh           chan error
}

func (s *Session) GetId() uint64 {
	return s.Id
}

func NewSession(conn net.Conn, ctx context.Context) *Session {
	s := &Session{
		Id:          atomic.AddUint64(&lastSessionId, 1),
		Conn:        conn,
		WriteBuffer: buf.NewBytesBuf(4*1024*1024, 1*time.Millisecond, conn, ctx),
		ReadDoneCh:  make(chan struct{}, 1),
		errCh:       make(chan error, 1),
		Ctx:         ctx,
	}
	return s
}

func (s *Session) WithErrCh(ech chan error) {
	s.errCh = ech
}

func (s *Session) Shutdown() error {
	<-s.ReadDoneCh
	s.WriteBuffer.Shutdown()
	return s.Conn.Close()
}
func (s *Session) Flush() error {
	return s.WriteBuffer.ForceFlush()
}
func (s *Session) Serve(handlers map[int]*giao.Handler) {
	defer func() {
		s.ReadDoneCh <- struct{}{}
	}()
	var readBuff = bytes.NewBuffer(make([]byte, 0, 4*1024*1024))
	go func() {
		err := s.WriteBuffer.StartFlush()
		if err != nil {
			s.errCh <- err
		}
	}()
	go func() {
		for {
			err := s.readAndHandle(readBuff, handlers)
			if err != nil && err != io.EOF {
				select {
				case <-s.Ctx.Done():
					return
				default:
				}
				s.errCh <- err
				return
			}
		}
	}()
	<-s.Ctx.Done()
}

func (s *Session) Error() chan error {
	return s.errCh
}
