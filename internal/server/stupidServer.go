package server

import (
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/session"
	"net"
)

type StupidServer struct {
	handlers map[int]*giao.Handler
	l        net.Listener
	sessions []*session.Session
}

func NewStupidServer() giao.Server {
	return &StupidServer{
		handlers: make(map[int]*giao.Handler, 8),
		sessions: make([]*session.Session, 0),
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
	errChan := make(chan error)
	go func() {
		for {
			conn, err := s.l.Accept()
			if err != nil {
				errChan <- err
				break
				//TODO log
			}
			go s.serve(conn, errChan)
		}
	}()
	select {
	case err := <-errChan:
		return err
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

func (s *StupidServer) serve(conn net.Conn, errChan chan error) {
	connSession := session.CreateSession(conn)
	s.sessions = append(s.sessions, connSession)
	err := connSession.Serve(s.handlers)
	if err != nil {
		<-errChan
	}
}

func (s *StupidServer) Shutdown() error {
	for _, ss := range s.sessions {
		_ = ss.Close()
	}
	return s.l.Close()
}
