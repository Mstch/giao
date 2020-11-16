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
	ReadBuf       []byte
	ReadHeaderBuf []byte
	WriteBufPool  *sync.Pool
	Meta          sync.Map
	WriteLock     sync.Mutex
	writeMsgChan  chan giao.Msg
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
		ReadBuf:       make([]byte, 64),
		ReadHeaderBuf: make([]byte, 8),
		WriteBufPool:  buffer.CommonBufferPool,
		WriteLock:     sync.Mutex{},
		Meta:          sync.Map{},
		writeMsgChan:  make(chan giao.Msg, 65535),
	}
	return s
}

func (s *Session) Close() error {
	s.closed = true
	s.Meta.Range(func(key interface{}, value interface{}) bool {
		s.Meta.Delete(key)
		return true
	})
	err := s.Conn.Close()
	return err
}

func (s *Session) Serve(handlers map[int]*giao.Handler) error {
	for !s.closed {
		handlerId, protoBytes, err := s.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if handler, ok := handlers[handlerId]; ok {
			reqPool := handler.InputPool
			pb := reqPool.Get().(giao.Msg)
			err := pb.Unmarshal(protoBytes)
			if err != nil {
				return err
			}
			reqPool.Put(pb)
			go handler.H(pb, s)
		}
	}
	return nil
}
