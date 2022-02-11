package client

import (
	"context"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/session"
	"net"
)

type StupidClient struct {
	connSession *session.Session
	handlers    map[int]*giao.Handler
	Ctx         context.Context
	Cancel      context.CancelFunc
}

func (c *StupidClient) Flush() error {
	return c.connSession.Flush()
}

func NewStupidClient() *StupidClient {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "name", "client")
	return &StupidClient{
		handlers: make(map[int]*giao.Handler, 8),
		Ctx:      ctx,
		Cancel:   cancel,
	}
}

func (c *StupidClient) Connect(network, address string) (giao.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	c.connSession = session.NewSession(conn, c.Ctx)
	return c, nil
}

func (c *StupidClient) Serve() error {
	c.connSession.Serve(c.handlers)
	select {
	case err := <-c.connSession.Error():
		return err
	case <-c.Ctx.Done():
		return nil
	}

}

func (c *StupidClient) Go(id int, req giao.Msg) error {
	return c.connSession.Write(id, req)
}

func (c *StupidClient) RegWithId(id int, handler *giao.Handler) giao.Client {
	c.handlers[id] = handler
	return c
}

func (c *StupidClient) Shutdown() error {
	c.Cancel()
	return c.connSession.Shutdown()
}
