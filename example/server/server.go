package main

import (
	"github.com/Mstch/giao"
	test "github.com/Mstch/giao/example/msg"
	"github.com/Mstch/giao/internal/server"
	"sync"
)

const TestRpc = 0

func TestRespHandler(in giao.Msg, s giao.Session) {
	respMsg := &test.Echo{}
	respMsg.Content = in.(*test.Echo).Content
	err := s.Write(TestRpc, respMsg)
	if err != nil {
		panic(err)
	}
}

func main() {
	testMsgPool := &sync.Pool{New: func() interface{} {
		return &test.Echo{}
	}}
	shandler := &giao.Handler{
		H:         TestRespHandler,
		InputPool: testMsgPool,
	}
	s := server.NewStupidServer().RegWithId(TestRpc, shandler)
	err := s.ListenAndServe("tcp", ":8888")
	if err != nil {
		panic(err)
	}
}
