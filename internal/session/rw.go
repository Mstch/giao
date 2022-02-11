package session

import (
	"encoding/binary"
	"errors"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/lib/bytes"
	"github.com/Mstch/giao/lib/local/pool"
)

var (
	ErrHandlerNotFound = errors.New("handler not found")
)

func (s *Session) handle(handler *giao.Handler, msg giao.Msg) {
	handler.Handle(msg, s)
	handler.InputPool.Put(msg)
}

func (s *Session) readAndHandle(buffer *bytes.Buffer, handlers map[int]*giao.Handler) (err error) {
	var hid, length int
	if buffer.Len() >= 8 {
		hid, _ = buffer.Read32Int()
		length, _ = buffer.Read32Int()
	} else {
		_, err = buffer.ReadFromAtLeast(s.Conn, 8)
		if err != nil {
			return
		}
		hid, _ = buffer.Read32Int()
		length, _ = buffer.Read32Int()
	}
	// read body
	if handler, ok := handlers[hid]; ok {
		msg := handler.InputPool.Get().(giao.Msg)
		if buffer.Len() >= length {
			err = buffer.ReadAMsg(length, msg)
			if err != nil {
				return
			}
		} else {
			_, err = buffer.ReadFromAtLeast(s.Conn, length)
			if err != nil {
				return
			}
			err = buffer.ReadAMsg(length, msg)
			if err != nil {
				return
			}
		}
		go s.handle(handler, msg)
	} else {
		err = ErrHandlerNotFound
	}
	return
}

func (s *Session) Write(handlerId int, msg giao.Msg) error {
	if msg == nil {
		return nil
	}
	headerBuf := pool.GetBytes(8)
	binary.BigEndian.PutUint32(headerBuf, uint32(handlerId))
	binary.BigEndian.PutUint32(headerBuf[4:], uint32(msg.Size()))
	err := s.WriteBuffer.Write(headerBuf, msg)
	pool.PutBytes(headerBuf)
	return err
}
