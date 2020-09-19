package client

import (
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/session"
	"github.com/gogo/protobuf/proto"
	"net"
)

type StupidClient struct {
	connSession *session.Session
	handlers    map[int]*giao.Handler
}

func NewStupidClient(network, address string) (*StupidClient, error) {
	c, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return &StupidClient{
		handlers:    make(map[int]*giao.Handler, 8),
		connSession: session.CreateSession(c),
	}, nil
}

func (s *StupidClient) ACall(id int, req proto.Message) error {
	return s.connSession.Writer(id, req)
}

func (s *StupidClient) Call(id int, req proto.Message) proto.Message {
	panic("implement me")
}

func (s *StupidClient) RegFuncWithId(id int, handler *giao.Handler) {
	s.handlers[id] = handler
}

func (s *StupidClient) StartServe() error {
	return s.connSession.Serve(s.handlers)
}

func (s *StupidClient) Shutdown() error {
	return s.connSession.Close()
}
