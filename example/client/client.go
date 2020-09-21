package main

import (
	"github.com/Mstch/giao"
	test "github.com/Mstch/giao/example/msg"
	"github.com/Mstch/giao/internal/client"
	"github.com/gogo/protobuf/proto"
	"sync"
)

const TestRpc = 0

func TestRespHandler(in proto.Message, out giao.ProtoWriter) {
	println(string(in.(*test.Echo).Content))
}

func main() {
	testMsgPool := &sync.Pool{New: func() interface{} {
		return &test.Echo{}
	}}
	chandler := &giao.Handler{
		H:       TestRespHandler,
		ReqPool: testMsgPool,
	}
	c, err := client.NewStupidClient().RegWithId(TestRpc, chandler).Connect("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	msg := testMsgPool.Get().(*test.Echo)
	msg.Content = []byte("fuck")
	err = c.Go(TestRpc, msg)
	if err != nil {
		panic(err)
	}
	err = c.Serve()
	if err != nil {
		panic(err)
	}
}
