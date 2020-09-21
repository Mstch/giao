package main

import (
	"github.com/Mstch/giao"
	test "github.com/Mstch/giao/example/msg"
	"github.com/Mstch/giao/internal/server"
	"github.com/gogo/protobuf/proto"
	"sync"
)

const TestRpc = 0

func TestRespHandler(in proto.Message, out giao.ProtoWriter) {
	respMsg := &test.Echo{}
	respMsg.Content = in.(*test.Echo).Content
	err := out(TestRpc, respMsg)
	if err != nil {
		panic(err)
	}
}

func main() {
	testMsgPool := &sync.Pool{New: func() interface{} {
		return &test.Echo{}
	}}
	shandler := &giao.Handler{
		H:       TestRespHandler,
		ReqPool: testMsgPool,
	}
	s := server.NewStupidServer().RegWithId(TestRpc, shandler)
	err := s.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
}
