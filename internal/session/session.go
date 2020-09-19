package session

import (
	"encoding/binary"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/buffer"
	"github.com/Mstch/giao/internal/errors"
	"github.com/gogo/protobuf/proto"
	"io"
	"net"
	"sync"
)

type Session struct {
	net.Conn
	ReadBuf      *buffer.Buffer
	WriteBufPool *sync.Pool
	Meta         sync.Map
	WriteLock    sync.Mutex
	Writer       giao.ProtoWriter
	closed       bool
}

func CreateSession(conn net.Conn) *Session {
	s := &Session{
		Conn:         conn,
		ReadBuf:      buffer.GetBuffer(),
		WriteBufPool: buffer.CommonBufferPool,
		WriteLock:    sync.Mutex{},
		Meta:         sync.Map{},
		closed:       false,
	}
	s.Writer = func(handlerId int, msg proto.Message) error {
		if msg == nil {
			return nil
		}
		if s.closed {
			return errors.ErrWriteToClosedConn
		}
		msgPb := msg.(giao.PB)
		size := msgPb.Size()
		writerBuf := s.WriteBufPool.Get().(*buffer.Buffer)
		totalBytes := writerBuf.Take(8 + size)
		headerBytes := totalBytes[:8]
		binary.BigEndian.PutUint32(headerBytes, uint32(handlerId))
		binary.BigEndian.PutUint32(headerBytes[4:8], uint32(size))
		protoBytes := totalBytes[8 : 8+size]
		marshalLen, err := msg.(giao.PB).MarshalTo(protoBytes)
		if err != nil {
			return err
		}
		if marshalLen != size {
			panic("len not match")
		}
		_, err = s.Write(totalBytes)
		s.WriteBufPool.Put(writerBuf)
		if err != nil {
			return err
		}
		return nil
	}
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

func (s *Session) doRead() (int, []byte, error) {
	//在fixedSizeBuffPool申请handlerId和length的buf
	headerBytes := buffer.EightBytesPool.Get().([]byte)
	handlerIdBytes := headerBytes[:4]
	lenBytes := headerBytes[4:8]
	//读handlerId
	_, err := io.ReadFull(s, handlerIdBytes)
	if err != nil {
		return 0, nil, err
	}
	handlerId := int(binary.BigEndian.Uint32(handlerIdBytes))
	//读length
	_, err = io.ReadFull(s, lenBytes)
	if err != nil {
		return 0, nil, err
	}
	protoLen := int(binary.BigEndian.Uint32(lenBytes))
	//归还buf
	buffer.EightBytesPool.Put(headerBytes)
	//申请proto的buf
	protoBytes := s.ReadBuf.Take(protoLen)
	_, err = io.ReadFull(s, protoBytes)
	if err != nil {
		return 0, nil, err
	}
	return handlerId, protoBytes, nil
}
