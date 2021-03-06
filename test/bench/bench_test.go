package bench

import (
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/client"
	"github.com/Mstch/giao/internal/server"
	test "github.com/Mstch/giao/test/msg"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

const EchoRpc = 0

var EchoMsg = make([]*test.Echo, 12)

func init() {
	for i := range EchoMsg {
		EchoMsg[i] = &test.Echo{}
		ContentLen := 512
		//for j := 0; j < i; j++ {
		//	ContentLen <<= 1
		//}
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

func BenchmarkStp1C(b *testing.B)  {
	b.SetBytes(512)
	sw := &sync.WaitGroup{}
	w := sync.WaitGroup{}
	w.Add(b.N)
	sw.Add(b.N)
	b.ReportAllocs()
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
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
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
	b.StopTimer()
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
func BenchmarkStp16C(b *testing.B) {
	b.SetBytes(5462 * 2)
	done := int32(0)
	w := sync.WaitGroup{}
	shandler := &giao.Handler{
		H: func(req giao.Msg, session giao.Session) {
			respMsg := echoPool.Get().(*test.Echo)
			respMsg.Content = req.(*test.Echo).Content
			err := session.Write(EchoRpc, respMsg)
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
			atomic.AddInt32(&done, 1)
		},
		InputPool: echoPool,
	}
	s := server.NewStupidServer().RegWithId(EchoRpc, shandler)
	err := s.Listen("tcp", ":8888")
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
	go func() {
		err = s.Serve()
		if err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				panic(err)
			}
		}
	}()

	w.Add(b.N)
	cs := make([]giao.Client, 16)
	for i := 0; i < 16; i++ {
		c, err := client.NewStupidClient().RegWithId(EchoRpc, chandler).Connect("tcp", "localhost:8888")
		cs[i] = c
		if err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				panic(err)
			}
		}
		go func(c giao.Client) {
			err := c.Serve()
			if err != nil {
				if !strings.HasSuffix(err.Error(), "use of closed network connection") {
					panic(err)
				}
			}
		}(c)
		go func(index int) {
			for j := 0; j < b.N/16; j++ {
				//msg := &test.Echo{}
				//msg.Content = EchoMsg[j%12].Content
				//msg.Index = int32(index)
				err := c.Go(EchoRpc, EchoMsg[j%12])
				if err != nil {
					if !strings.HasSuffix(err.Error(), "use of closed network connection") {
						panic(err)
					}
				}
			}
		}(i)
	}
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
	b.ResetTimer()
	go func(index int) {
		for j := 0; j < b.N%16; j++ {
			//msg := &test.Echo{}
			//msg.Content = EchoMsg[j%12].Content
			//msg.Index = int32(index)
			err := c.Go(EchoRpc, EchoMsg[j%12])
			if err != nil {
				if !strings.HasSuffix(err.Error(), "use of closed network connection") {
					panic(err)
				}
			}
		}
	}(16)
	w.Wait()
	b.StopTimer()
	err = c.Shutdown()
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
	for _, c1 := range cs {
		err = c1.Shutdown()
		if err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				panic(err)
			}
		}
	}
	err = s.Shutdown()
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
}
