package bench

import (
	"context"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/server"
	test "github.com/Mstch/giao/test/msg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/rpc"
	"testing"
)

type GrpcEcho struct {
}

func (g *GrpcEcho) DoEcho(ctx context.Context, echo *test.Echo) (*test.Echo, error) {
	out := &test.Echo{}
	out.Content = echo.Content
	out.Index = echo.Index
	return out, nil
}

type StandardEcho struct {
}

func (e *StandardEcho) DoEcho(req *test.Echo, resp *test.Echo) error {
	resp = &test.Echo{}
	resp.Content = req.Content
	return nil
}

func TestStpServer(t *testing.T) {
	shandler := &giao.Handler{
		H: func(req giao.Msg, session giao.Session) {
			respMsg := &test.Echo{}
			respMsg.Content = req.(*test.Echo).Content
			err := session.Write(EchoRpc, respMsg)
			if err != nil {
				panic(err)
			}
		},
		InputPool: echoPool,
	}
	benchmarkStupidEchoServer := server.NewStupidServer().RegWithId(EchoRpc, shandler)
	err := benchmarkStupidEchoServer.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	err = benchmarkStupidEchoServer.Shutdown()
	if err != nil {
		panic(err)
	}
}
func TestStdServer(t *testing.T) {
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
	benchmarkStandardEchoServer.Accept(l)
}
func TestGRpcServer(t *testing.T) {
	var err error
	lis, err := net.Listen("tcp", ":8181")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	test.RegisterEchoServiceServer(s, &GrpcEcho{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
