package bench

import (
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/server"
	test "github.com/Mstch/giao/test/msg"
	"github.com/gogo/protobuf/proto"
	"net"
	"net/rpc"
	"testing"
)

var benchmarkStupidEchoServer giao.Server
var benchmarkStandardEchoServer *rpc.Server

func TestInitServer(t *testing.T) {
	var err error
	shandler := &giao.Handler{
		H: func(req proto.Message, respWriter giao.ProtoWriter) {
			respMsg := &test.Echo{}
			respMsg.Content = req.(*test.Echo).Content
			err := respWriter(EchoRpc, respMsg)
			if err != nil {
				panic(err)
			}
		},
		ReqPool: echoPool,
	}
	benchmarkStupidEchoServer = server.NewStupidServer().RegWithId(EchoRpc, shandler)
	go func() {
		err := benchmarkStupidEchoServer.Listen("tcp", ":8888")
		if err != nil {
			panic(err)
		}
	}()
	benchmarkStandardEchoServer = rpc.NewServer()
	err = benchmarkStandardEchoServer.Register(&Echo{})
	if err != nil {
		panic(err)
	}
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	benchmarkStandardEchoServer.Accept(l)
}
