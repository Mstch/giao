package bench

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/client"
	"github.com/Mstch/giao/internal/server"
	test "github.com/Mstch/giao/test/msg"
)

const EchoRpc = 0

var EchoMsg = make([]*test.Echo, 12)

func init() {
	for i := range EchoMsg {
		EchoMsg[i] = &test.Echo{}
		ContentLen := 16
		for j := 0; j < i; j++ {
			ContentLen <<= 1
		}
		EchoMsg[i].Content = make([]byte, ContentLen)
		for j := range EchoMsg[i].Content {
			EchoMsg[i].Content[j] = 'f'
		}
	}
}

var spool = &giao.NoCachePool{New: func() interface{} {
	return &test.Echo{}
}}
var cpool = &sync.Pool{New: func() interface{} {
	return &test.Echo{}
}}

func BenchmarkStp1C(b *testing.B) {
	b.SetBytes(5470 * 2)
	scheck := make([]int32, b.N)
	ccheck := make([]int32, b.N)
	w := sync.WaitGroup{}
	sw := sync.WaitGroup{}
	w.Add(b.N)
	sw.Add(b.N)
	b.ReportAllocs()
	shandler := &giao.Handler{
		Handle: func(req giao.Msg, session giao.Session) {
			echo := req.(*test.Echo)
			if ok := atomic.CompareAndSwapInt32(&scheck[echo.Index], 0, 1); !ok {
				// panic(echo.Index)
			}
			respMsg := spool.Get().(*test.Echo)
			respMsg.Index = echo.Index
			respMsg.Content = echo.Content
			err := session.Write(EchoRpc, respMsg)
			if err != nil {
				panic(err)
			}
			spool.Put(respMsg)
			sw.Done()
		},
		InputPool: spool,
	}
	chandler := &giao.Handler{
		Handle: func(req giao.Msg, session giao.Session) {
			echo := req.(*test.Echo)
			if ok := atomic.CompareAndSwapInt32(&ccheck[echo.Index], 0, 1); !ok {
				// panic(echo.Index)
			}
			w.Done()
		},
		InputPool: cpool,
	}
	s, err := server.NewStupidServer().RegWithId(EchoRpc, shandler).Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	go func() {
		err := s.Serve()
		if err != nil {
			panic(err)
		}
	}()
	c, err := client.NewStupidClient().RegWithId(EchoRpc, chandler).Connect("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}
	go func() {
		err := c.Serve()
		if err != nil {
			panic(err)
		}
	}()
	b.StartTimer()
	for j := 0; j < b.N; j++ {
		echo := EchoMsg[j%12]
		echo.Index = int32(j)
		err := c.Go(EchoRpc, echo)
		if err != nil {
			panic(err)
		}
	}
	sw.Wait()
	w.Wait()
	b.StopTimer()
	err = c.Shutdown()
	if err != nil {
		panic(err)
	}
	err = s.Shutdown()
	if err != nil {
		panic(err)
	}
}
func BenchmarkStp16C(b *testing.B) {
	b.SetBytes(5460 * 2)
	done := int32(0)
	w := sync.WaitGroup{}
	shandler := &giao.Handler{
		Handle: func(req giao.Msg, session giao.Session) {
			respMsg := spool.Get().(*test.Echo)
			respMsg.Content = req.(*test.Echo).Content
			err := session.Write(EchoRpc, respMsg)
			spool.Put(respMsg)
			if err != nil {
				if !strings.HasSuffix(err.Error(), "use of closed network connection") {
					panic(err)
				}
			}
		},
		InputPool: spool,
	}
	chandler := &giao.Handler{
		Handle: func(req giao.Msg, session giao.Session) {
			w.Done()
			atomic.AddInt32(&done, 1)
		},
		InputPool: spool,
	}
	s, err := server.NewStupidServer().RegWithId(EchoRpc, shandler).Listen("tcp", ":8888")
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
				// msg := &test.Echo{}
				// msg.Content = EchoMsg[j%12].Content
				// msg.Index = int32(index)
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
			// msg := &test.Echo{}
			// msg.Content = EchoMsg[j%12].Content
			// msg.Index = int32(index)
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
