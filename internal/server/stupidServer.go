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
	errCh    chan error
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

func (s *StupidServer) Listen(network, address string) (giao.Server, error) {
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	s.l = l
	return s, nil

}

func (s *StupidServer) Serve() error {
	connCh := make(chan net.Conn, 64)
	go func() {
		for {
			conn, err := s.l.Accept()
			if err != nil {
				select {
				case <-s.Ctx.Done():
					return
				default:
				}
			}
			go s.serve(conn)
		}
	}()
	for {
		select {
		case <-s.Ctx.Done():
			return nil
		case conn := <-connCh:
			go s.serve(conn)
		case err := <-s.errCh:
			s.cancel()
			return err
		}
	}
}

func (s *StupidServer) ListenAndServe(network, address string) error {
	_, err := s.Listen(network, address)
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
	connSession := session.NewSession(conn, s.Ctx)
	connSession.WithErrCh(s.errCh)
	s.sessions = append(s.sessions, connSession)
	connSession.Serve(s.handlers)
}

func (s *StupidServer) Shutdown() error {
	s.cancel()
	for _, ss := range s.sessions {
		err := ss.Shutdown()
		if err != nil {
			return err
		}
	}
	return s.l.Close()
}

func (s *StupidServer) Flush() {
	for _, ss := range s.sessions {
		_ = ss.Flush()
	}
}
