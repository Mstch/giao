package main

import (
	"github.com/Mstch/giao"
	test "github.com/Mstch/giao/example/msg"
	"github.com/Mstch/giao/internal/server"
	"github.com/gogo/protobuf/proto"
	"sync"
)

const TestRpc = 0

func Test(req proto.Message, writer giao.ProtoWriter) {
	respMsg := &test.Echo{}
	respMsg.Content = req.(*test.Echo).Content
	err := writer(TestRpc, respMsg)
	if err != nil {
		panic(err)
	}
}

func main() {
	testMsgPool := &sync.Pool{New: func() interface{} {
		return &test.Echo{}
	}}
	shandler := &giao.Handler{
		H:       Test,
		ReqPool: testMsgPool,
	}
	s := server.NewStupidServer().RegWithId(TestRpc, shandler)
	err := s.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
}
