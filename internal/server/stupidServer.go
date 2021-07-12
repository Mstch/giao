package server

import (
	"context"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/session"
	"net"
)

type StupidServer struct {
	handlers map[int]*giao.Handler
	Ctx      context.Context
	cancel   context.CancelFunc
	l        net.Listener
	sessions []*session.Session
}

func NewStupidServer() giao.Server {
	ctx := context.WithValue(context.Background(), "name", "server")
	ctx, cancel := context.WithCancel(ctx)
	return &StupidServer{
		handlers: make(map[int]*giao.Handler, 8),
		sessions: make([]*session.Session, 0),
		Ctx:      ctx,
		cancel:   cancel,
	}
}

func (s *StupidServer) Listen(network, address string) error {
	l, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	s.l = l
	return nil

}

func (s *StupidServer) Serve() error {
	for {
		conn, err := s.l.Accept()
		if err != nil {
			s.cancel()
			return err
		}
		go s.serve(conn)
	}
}

func (s *StupidServer) ListenAndServe(network, address string) error {
	err := s.Listen(network, address)
	if err != nil {
		return err
	}
	return s.Serve()
}

func (s *StupidServer) RegWithId(id int, handler *giao.Handler) giao.Server {
	s.handlers[id] = handler
	return s
}

func (s *StupidServer) serve(conn net.Conn) {
	connSession := session.CreateSession(conn, s.Ctx)
	s.sessions = append(s.sessions, connSession)
	connSession.Serve(s.handlers)
	s.Ctx.Err()
}

func (s *StupidServer) Shutdown() error {
	s.cancel()
	for _, ss := range s.sessions {
		_ = ss.Close()
	}
	return s.l.Close()
}


func (s *StupidServer) Flush() {
	for _, ss := range s.sessions {
		_ = ss.Flush()
	}
}
