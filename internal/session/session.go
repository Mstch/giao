package session

import (
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/buffer"
	"github.com/gogo/protobuf/proto"
	"io"
	"net"
	"sync"
)

type Session struct {
	net.Conn
	ReadBuf       *buffer.Buffer
	ReadHeaderBuf []byte
	WriteBufPool  *sync.Pool
	Meta          sync.Map
	WriteLock     sync.Mutex
	Writer        giao.ProtoWriter
	closed        bool
}

func CreateSession(conn net.Conn) *Session {
	s := &Session{
		Conn:          conn,
		ReadBuf:       buffer.GetBuffer(),
		ReadHeaderBuf: make([]byte, 8),
		WriteBufPool:  buffer.CommonBufferPool,
		WriteLock:     sync.Mutex{},
		Meta:          sync.Map{},
		closed:        false,
	}
	s.Writer = s.doWrite
	return s
}

func (s *Session) Close() error {
	s.closed = true
	s.Meta.Range(func(key interface{}, value interface{}) bool {
		s.Meta.Delete(key)
		return true
	})
	return s.Conn.Close()
}

func (s *Session) Serve(handlers map[int]*giao.Handler) error {
	for {
		handlerId, protoBytes, err := s.doRead()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		if handler, ok := handlers[handlerId]; ok {
			reqPool := handler.ReqPool
			pb := reqPool.Get().(giao.PB)
			err := pb.Unmarshal(protoBytes)
			if err != nil {
				//todo
				continue
			}
			reqPool.Put(pb)
			go handler.H(pb.(proto.Message), s.Writer)
		}
	}
	return s.Close()
}
