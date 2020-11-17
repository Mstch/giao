package session

import (
	"encoding/binary"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/buffer"
	"github.com/Mstch/giao/internal/errors"
	"io"
	"time"
)

func (s *Session) Read() (int, []byte, error) {
	//读handlerId
	_, err := io.ReadFull(s.Conn, s.ReadHeaderBuf[:4])
	if err != nil {
		return 0, nil, err
	}
	//读length
	_, err = io.ReadFull(s.Conn, s.ReadHeaderBuf[4:8])
	if err != nil {
		return 0, nil, err
	}
	length := int(binary.BigEndian.Uint32(s.ReadHeaderBuf[4:8]))
	//申请proto的buf
	if len(s.ReadBuf) < length {
		s.ReadBuf = make([]byte, 64*(length/64+1))
	}
	_, err = io.ReadFull(s.Conn, s.ReadBuf[:length])
	if err != nil {
		return 0, nil, err
	}
	return int(binary.BigEndian.Uint32(s.ReadHeaderBuf[:4])), s.ReadBuf[:length], nil
}

func (s *Session) Write(handlerId int, msg giao.Msg) error {
	if msg == nil {
		return nil
	}
	if s.closed {
		return errors.ErrWriteToClosedConn
	}
	s.writeMsgChan <- msg
	writerBuf := s.WriteBufPool.Get().(*buffer.Buffer)
	defer s.WriteBufPool.Put(writerBuf)
	writerBuf.Reset()
	needWrite := 0
	retries := 0
	for {
		select {
		case msgInChan := <-s.writeMsgChan:
			{
				size := msgInChan.Size()
				writerBuf.WriteUint32(uint32(handlerId))
				writerBuf.WriteUint32(uint32(size))
				_, err := writerBuf.WriteMsg(msgInChan)
				if err != nil {
					return err
				}
				needWrite++
			}
		default:
			if 0 < needWrite && needWrite < 5 && retries < 5 {
				retries++
				time.Sleep(200)
			} else {
				_, err := writerBuf.WriteTo(s.Conn)
				if err != nil {
					return err
				}
				goto end
			}
		}
	}
end:
	return nil
}
