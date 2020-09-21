package session

import (
	"encoding/binary"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/buffer"
	"github.com/Mstch/giao/internal/errors"
	"github.com/gogo/protobuf/proto"
	"io"
)

func (s *Session) doRead() (int, []byte, error) {
	//读handlerId
	_, err := io.ReadFull(s, s.ReadHeaderBuf[:4])
	if err != nil {
		return 0, nil, err
	}
	//读length
	_, err = io.ReadFull(s, s.ReadHeaderBuf[4:8])
	if err != nil {
		return 0, nil, err
	}
	//申请proto的buf
	protoBytes := s.ReadBuf.Take(int(binary.BigEndian.Uint32(s.ReadHeaderBuf[4:8])))
	_, err = io.ReadFull(s, protoBytes)
	if err != nil {
		return 0, nil, err
	}
	return int(binary.BigEndian.Uint32(s.ReadHeaderBuf[:4])), protoBytes, nil
}

func (s *Session) doWrite(handlerId int, msg proto.Message) error {
	if msg == nil {
		return nil
	}
	if s.closed {
		return errors.ErrWriteToClosedConn
	}
	msgPb := msg.(giao.PB)
	size := msgPb.Size()
	writerBuf := s.WriteBufPool.Get().(*buffer.Buffer)
	defer s.WriteBufPool.Put(writerBuf)
	totalBytes := writerBuf.Take(8 + size)
	headerBytes := totalBytes[:8]
	binary.BigEndian.PutUint32(headerBytes, uint32(handlerId))
	binary.BigEndian.PutUint32(headerBytes[4:8], uint32(size))
	protoBytes := totalBytes[8 : 8+size]
	_, err := msg.(giao.PB).MarshalTo(protoBytes)
	if err != nil {
		return err
	}
	_, err = s.Write(totalBytes)
	return err
}
