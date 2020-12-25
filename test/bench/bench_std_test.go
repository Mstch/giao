package bench

import (
	test "github.com/Mstch/giao/test/msg"
	"net"
	"net/rpc"
	"strings"
	"testing"
)

type StandardEcho struct {
}

func (e *StandardEcho) DoEcho(req *test.Echo, resp *test.Echo) error {
	resp = &test.Echo{}
	resp.Content = req.Content
	return nil
}

func BenchmarkStd1C(b *testing.B) {
	b.SetBytes(5462 * 2)
	var err error
	benchmarkStandardEchoServer := rpc.NewServer()
	err = benchmarkStandardEchoServer.Register(&StandardEcho{})
	if err != nil {
		panic(err)
	}
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	go benchmarkStandardEchoServer.Accept(l)
	c, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	callCh := make(chan *rpc.Call, b.N/2+10)
	go func() {
		for j := 0; j < b.N; j++ {
			_ = c.Go("StandardEcho.DoEcho", EchoMsg[j%12], &test.Echo{}, callCh)
		}
	}()

	for j := 0; j < b.N; j++ {
		call := <-callCh
		if call.Error != nil {
			panic(call.Error)
		}
	}

	err = l.Close()
	if err != nil {
		panic(err)
	}
	err = c.Close()
	if err != nil {
		panic(err)
	}
}
func BenchmarkStd16C(b *testing.B) {
	b.SetBytes(5462 * 2)

	var err error
	benchmarkStandardEchoServer := rpc.NewServer()
	err = benchmarkStandardEchoServer.Register(&StandardEcho{})
	if err != nil {
		panic(err)
	}
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	go benchmarkStandardEchoServer.Accept(l)
	callCh := make(chan *rpc.Call, b.N)
	for i := 0; i < 16; i++ {
		c, err := rpc.Dial("tcp", "localhost:8080")
		if err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				panic(err)
			}
		}
		go func(index int) {
			for j := 0; j < b.N/16; j++ {
				msg := &test.Echo{}
				msg.Content = EchoMsg[j%12].Content
				msg.Index = int32(index)
				_ = c.Go("StandardEcho.DoEcho", EchoMsg[j%12], &test.Echo{}, callCh)
			}
		}(i)
	}
	c, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
	go func(index int) {
		for j := 0; j < b.N%16; j++ {
			msg := &test.Echo{}
			msg.Content = EchoMsg[j%12].Content
			msg.Index = int32(index)
			_ = c.Go("StandardEcho.DoEcho", EchoMsg[j%12], &test.Echo{}, callCh)
		}
	}(0)
	for j := 0; j < b.N; j++ {
		call := <-callCh
		if call.Error != nil {
			panic(call.Error)
		}
	}

	err = l.Close()
	if err != nil {
		panic(err)
	}
	err = c.Close()
	if err != nil {
		panic(err)
	}
}
