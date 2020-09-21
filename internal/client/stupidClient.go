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

func NewStupidClient() *StupidClient {
	return &StupidClient{
		handlers: make(map[int]*giao.Handler, 8),
	}
}

func (c *StupidClient) Connect(network, address string) (giao.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	c.connSession = session.CreateSession(conn)
	return c, nil
}

func (c *StupidClient) Serve() error {
	return c.connSession.Serve(c.handlers)
}

func (c *StupidClient) Go(id int, req proto.Message) error {
	return c.connSession.Writer(id, req)
}

func (c *StupidClient) RegWithId(id int, handler *giao.Handler) giao.Client {
	c.handlers[id] = handler
	return c
}

func (c *StupidClient) Shutdown() error {
	return c.connSession.Close()
}
