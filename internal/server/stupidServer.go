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
	errChan := make(chan error)
	go func() {
		for {
			conn, err := s.l.Accept()
			if err != nil {
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
		err := ss.Close()
		if err != nil {
			return err
		}
	}
	return s.l.Close()
}
