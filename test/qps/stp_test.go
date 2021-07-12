package qps

import (
	"fmt"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/client"
	"github.com/Mstch/giao/internal/server"
	test "github.com/Mstch/giao/test/msg"
	"strings"
	"sync"
	"testing"
	"time"
)

const EchoRpc = 0

var EchoMsg = make([]*test.Echo, 12)

func init() {
	for i := range EchoMsg {
		EchoMsg[i] = &test.Echo{}
		ContentLen := 512
		str := ""
		for i := 0; i < ContentLen; i++ {
			str += "f"
		}
		EchoMsg[i].Content = []byte(str)
	}
}

var echoPool = &sync.Pool{New: func() interface{} {
	return &test.Echo{}
}}

func TestQPS(t *testing.T) {
	start := time.Now()
	defer func() {
		fmt.Println(" 1000_0000 cost ", (time.Now().Sub(start)).Seconds())
	}()
	n := 1000_0000
	sw := &sync.WaitGroup{}
	w := sync.WaitGroup{}
	w.Add(n)
	sw.Add(n)
	shandler := &giao.Handler{
		H: func(req giao.Msg, session giao.Session) {
			respMsg := echoPool.Get().(*test.Echo)
			respMsg.Content = req.(*test.Echo).Content
			err := session.Write(EchoRpc, respMsg)
			sw.Done()
			echoPool.Put(respMsg)
			if err != nil {
				if !strings.HasSuffix(err.Error(), "use of closed network connection") {
					panic(err)
				}
			}
		},
		InputPool: echoPool,
	}
	chandler := &giao.Handler{
		H: func(req giao.Msg, session giao.Session) {
			w.Done()
		},
		InputPool: echoPool,
	}
	benchmarkStupidEchoServer := server.NewStupidServer().RegWithId(EchoRpc, shandler)
	err := benchmarkStupidEchoServer.Listen("tcp", ":8888")
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
	go func() {
		err = benchmarkStupidEchoServer.Serve()
		if err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				panic(err)
			}
		}
	}()
	c, err := client.NewStupidClient().RegWithId(EchoRpc, chandler).Connect("tcp", "localhost:8888")
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
	go func() {
		err := c.Serve()
		if err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				panic(err)
			}
		}
	}()
	for j := 0; j < n; j++ {
		err := c.Go(EchoRpc, EchoMsg[j%12])
		if err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				panic(err)
			}
		}
	}
	err = c.Flush()
	if err != nil {
		panic(err)
	}
	sw.Wait()
	benchmarkStupidEchoServer.Flush()
	w.Wait()
	err = c.Shutdown()
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
	err = benchmarkStupidEchoServer.Shutdown()
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
}
func TestQPS2(t *testing.T) {
	start := time.Now()
	defer func() {
		fmt.Println(" 2000_0000 cost ", (time.Now().Sub(start)).Seconds())
	}()
	n := 2000_0000
	sw := &sync.WaitGroup{}
	w := sync.WaitGroup{}
	w.Add(n)
	sw.Add(n)
	shandler := &giao.Handler{
		H: func(req giao.Msg, session giao.Session) {
			respMsg := echoPool.Get().(*test.Echo)
			respMsg.Content = req.(*test.Echo).Content
			err := session.Write(EchoRpc, respMsg)
			sw.Done()
			echoPool.Put(respMsg)
			if err != nil {
				if !strings.HasSuffix(err.Error(), "use of closed network connection") {
					panic(err)
				}
			}
		},
		InputPool: echoPool,
	}
	chandler := &giao.Handler{
		H: func(req giao.Msg, session giao.Session) {
			w.Done()
		},
		InputPool: echoPool,
	}
	benchmarkStupidEchoServer := server.NewStupidServer().RegWithId(EchoRpc, shandler)
	err := benchmarkStupidEchoServer.Listen("tcp", ":8888")
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
	go func() {
		err = benchmarkStupidEchoServer.Serve()
		if err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				panic(err)
			}
		}
	}()
	c, err := client.NewStupidClient().RegWithId(EchoRpc, chandler).Connect("tcp", "localhost:8888")
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
	go func() {
		err := c.Serve()
		if err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				panic(err)
			}
		}
	}()
	for j := 0; j < n; j++ {
		err := c.Go(EchoRpc, EchoMsg[j%12])
		if err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				panic(err)
			}
		}
	}
	err = c.Flush()
	if err != nil {
		panic(err)
	}
	sw.Wait()
	benchmarkStupidEchoServer.Flush()
	w.Wait()
	err = c.Shutdown()
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
	err = benchmarkStupidEchoServer.Shutdown()
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
}
