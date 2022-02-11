package mock

import (
	"io"
	"net"
)

func NewTcpConn() (net.Conn, net.Conn, error) {
	var l net.Listener
	var r net.Conn
	var w net.Conn
	var err error
	l, err = net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return nil, nil, err
	}
	connected := make(chan error)
	addr := l.Addr().String()
	go func() {
		var werr error
		w, werr = net.Dial("tcp", addr)
		connected <- werr
	}()
	r, err = l.Accept()
	if err != nil {
		return nil, nil, err
	}
	err = l.Close()
	if err != nil {
		return nil, nil, err
	}
	err = <-connected

	return r, w, err
}

func NewUdpConn() (r io.Reader, w io.Writer, err error) {
	var conn *net.UDPConn
	conn, err = net.ListenUDP("udp", nil)
	if err != nil {
		return nil, nil, err
	}
	r = conn
	w, err = net.DialUDP("udp", nil, conn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		return nil, nil, err
	}
	return
}
