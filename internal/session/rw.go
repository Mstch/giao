package session

import (
	"context"
	"encoding/binary"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/errors"
	"io"
)

func (s *Session) Read() (int, []byte, error) {
	//读handlerId & 读length
	_, err := io.ReadFull(s.Conn, s.ReadHeaderBuf[:])
	if err != nil {
		return 0, nil, err
	}
	length := int(binary.BigEndian.Uint32(s.ReadHeaderBuf[4:8]))
	handlerId := int(binary.BigEndian.Uint32(s.ReadHeaderBuf[:4]))
	//申请proto的buf
	protoBytes := s.ReadBuf.Take(length)
	_, err = io.ReadFull(s.Conn, protoBytes)
	if err != nil {
		return 0, nil, err
	}
	return handlerId, protoBytes, nil
}

func (s *Session) Write(handlerId int, msg giao.Msg) error {
	if msg == nil {
		return nil
	}
	if s.Ctx.Err() == context.Canceled {
		return errors.ErrWriteToClosedConn
	}
	return s.WriteBuffer.Write(msg, handlerId)
}
