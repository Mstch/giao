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

func NewStupidServer(network, address string) (giao.Server, error) {
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return &StupidServer{
		handlers: make(map[int]*giao.Handler, 8),
		l:        l,
		sessions: make([]*session.Session, 0),
	}, nil
}

func (s *StupidServer) RegStruct(handlerStruct interface{}) {
	panic("implement me")
}

func (s *StupidServer) RegStructWithId(id int, handlerStruct interface{}) {
	panic("implement me")
}

func (s *StupidServer) RegFuncWithId(id int, handler *giao.Handler) {
	s.handlers[id] = handler
}

func (s *StupidServer) StartServe() error {
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
