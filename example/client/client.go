package main

import (
	"github.com/Mstch/giao"
	test "github.com/Mstch/giao/example/msg"
	"github.com/Mstch/giao/internal/client"
	"sync"
)

const TestRpc = 0

var done = make(chan bool)

func TestRespHandler(in giao.Msg, s giao.Session) {
	println(string(in.(*test.Echo).Content))
	done <- false
}

func main() {
	testMsgPool := &sync.Pool{New: func() interface{} {
		return &test.Echo{}
	}}
	chandler := &giao.Handler{
		H:         TestRespHandler,
		InputPool: testMsgPool,
	}
	c, err := client.NewStupidClient().RegWithId(TestRpc, chandler).Connect("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	msg := testMsgPool.Get().(*test.Echo)
	msg.Content = []byte("fuck")
	go func() {
		err = c.Serve()
		if err != nil {
			panic(err)
		}
	}()
	err = c.Go(TestRpc, msg)
	if err != nil {
		panic(err)
	}
	<-done
}
