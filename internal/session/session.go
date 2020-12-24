package session

import (
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/buffer"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

var lastSessionId = uint64(0)

type Session struct {
	net.Conn
	Id            uint64
	ReadBuf       *buffer.Buffer
	ReadHeaderBuf []byte
	WriteBufPool  *sync.Pool
	WriteBatchBuf *buffer.BatchBuffer
	Meta          sync.Map
	WriteLock     sync.Mutex
	closed        bool
}

func (s *Session) Get(key interface{}) (interface{}, bool) {
	return s.Meta.Load(key)
}

func (s *Session) Set(key, value interface{}) {
	s.Meta.Store(key, value)
}

func (s *Session) GetId() uint64 {
	return s.Id
}

func CreateSession(conn net.Conn) *Session {
	s := &Session{
		Id:            atomic.AddUint64(&lastSessionId, 1),
		Conn:          conn,
		ReadBuf:       buffer.GetBuffer(),
		ReadHeaderBuf: make([]byte, 8),
		WriteBufPool:  buffer.CommonBufferPool,
		WriteBatchBuf: buffer.NewBatchBuffer(conn),
		WriteLock:     sync.Mutex{},
		Meta:          sync.Map{},
	}

	return s
}

func (s *Session) Close() error {
	s.closed = true
	err := s.WriteBatchBuf.Stop()
	if err != nil {
		return err
	}
	s.Meta.Range(func(key interface{}, value interface{}) bool {
		s.Meta.Delete(key)
		return true
	})
	err = s.Conn.Close()
	return err
}

func (s *Session) Serve(handlers map[int]*giao.Handler) error {
	eChan := make(chan error, 2)
	go func() {
		eChan <- s.WriteBatchBuf.StartFlushLooper()
	}()
	go func() {
		for !s.closed {
			handlerId, protoBytes, err := s.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				eChan <- err
				return
			}
			if handler, ok := handlers[handlerId]; ok {
				reqPool := handler.InputPool
				pb := reqPool.Get().(giao.Msg)
				err := pb.Unmarshal(protoBytes)
				if err != nil {
					eChan <- err
				}
				reqPool.Put(pb)
				go handler.H(pb, s)
			}
		}
		eChan <- nil
	}()
	if err := <-eChan; err == nil {
		return <-eChan
	} else {
		return err
	}
}
