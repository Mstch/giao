package session

import (
	"context"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/buffer"
	"github.com/Mstch/giao/internal/buffer/flushbuffer"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

var lastSessionId = uint64(0)

type Session struct {
	net.Conn
	Id            uint64
	ReadHeaderBuf []byte
	ReadBuf       *buffer.SimpleBuffer
	WriteBuffer   *flushbuffer.FBuffers
	Ctx           context.Context
	Cancel        context.CancelFunc
	ReadDoneCh    chan struct{}
}

func (s *Session) GetId() uint64 {
	return s.Id
}

func CreateSession(conn net.Conn, parentCtx context.Context) *Session {
	ctx := context.WithValue(parentCtx, "name", "session")
	ctx, cancel := context.WithCancel(ctx)
	s := &Session{
		Id:            atomic.AddUint64(&lastSessionId, 1),
		Conn:          conn,
		ReadHeaderBuf: make([]byte, 8),
		ReadBuf:       buffer.NewSimpleBuffer(),
		WriteBuffer:   flushbuffer.NewFBuffers(1*time.Millisecond, conn),
		ReadDoneCh:    make(chan struct{}),
		Cancel:        cancel,
		Ctx:           ctx,
	}
	return s
}

func (s *Session) Close() error {
	s.Cancel()
	err := s.WriteBuffer.StopFlushTimer()
	if err != nil {
		return err
	}
	return s.Conn.Close()
}
func (s *Session) Flush() error {
	return s.WriteBuffer.ForceFlush()
}
func (s *Session) Serve(handlers map[int]*giao.Handler) error {
	go func() {
		s.WriteBuffer.StartFlusher(s.Ctx)
	}()
	go func() {
		for {
			select {
			case <-s.Ctx.Done():
				return
			default:
				handlerId, protoBytes, err := s.Read()
				if err != nil {
					if err == io.EOF || strings.HasSuffix(err.Error(), "use of closed network connection") {
						return
					}
					panic(err) //todo
				}
				if handler, ok := handlers[handlerId]; ok {
					reqPool := handler.InputPool
					pb := reqPool.Get().(giao.Msg)
					err := pb.Unmarshal(protoBytes)
					if err != nil {
						panic(err) //todo
					}
					reqPool.Put(pb)
					go handler.H(pb, s)
				}
			}
		}
	}()
	return nil
}
