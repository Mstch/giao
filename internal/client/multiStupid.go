package client

import (
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/selector"
	"github.com/Mstch/giao/internal/session"
	"net"
)

type MultiConnClient struct {
	handlers map[int]*giao.Handler
	selector giao.Selector
}

func NewMultiConnStupidClient() giao.MultiConnClient {
	return &MultiConnClient{
		handlers: make(map[int]*giao.Handler, 8),
		selector: &selector.RoundRobinSelector{},
	}
}

func (m *MultiConnClient) Connect(network, address string) (giao.MultiConnClient, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	connSession := session.CreateSession(conn, nil)
	m.selector.AddSession(connSession)
	return m, nil
}

func (m *MultiConnClient) Serve() chan error {
	errChan := make(chan error)
	for _, connSession := range m.selector.GetAllSession() {
		go func(connSession giao.Session) {
			connSession.(*session.Session).Serve(m.handlers)
		}(connSession)
	}
	return errChan
}

func (m *MultiConnClient) RegWithId(id int, handler *giao.Handler) giao.MultiConnClient {
	m.handlers[id] = handler
	return m
}

func (m *MultiConnClient) Go(id int, req giao.Msg) error {
	return m.selector.Select().Write(id, req)
}

func (m *MultiConnClient) Broadcast(id int, req giao.Msg) chan error {
	errChan := make(chan error)
	for _, connSession := range m.selector.GetAllSession() {
		go func(connSession giao.Session) {
			errChan <- connSession.(*session.Session).Write(id, req)
		}(connSession)
	}
	return errChan
}

func (m *MultiConnClient) SetSelector(selector giao.Selector) {
	m.selector = selector
}

func (m *MultiConnClient) Shutdown() chan error {
	errChan := make(chan error)
	for _, connSession := range m.selector.GetAllSession() {
		go func(connSession giao.Session) {
			connSession.(*session.Session).Serve(m.handlers)
		}(connSession)
	}
	return errChan
}
