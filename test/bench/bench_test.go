package qps

import (
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/client"
	"github.com/Mstch/giao/internal/server"
	test "github.com/Mstch/giao/test/msg"
	"github.com/gogo/protobuf/proto"
	"net"
	"net/rpc"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const EchoRpc = 0

var EchoMsg = make([]*test.Echo, 12)

func init() {
	for i := range EchoMsg {
		EchoMsg[i] = &test.Echo{}
		ContentLen := 1
		for j := 0; j < i; j++ {
			ContentLen <<= 1
		}
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

var benchmarkStupidEchoServer giao.Server
var benchmarkStandardEchoServer *rpc.Server

type Echo struct {
}

func (e *Echo) GoEcho(req *test.Echo, resp *test.Echo) error {
	resp = &test.Echo{}
	resp.Content = req.Content
	return nil
}

func init() {
	var err error
	benchmarkStupidEchoServer, err = server.NewStupidServer("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	go func() {
		err := benchmarkStupidEchoServer.StartServe()
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
	go benchmarkStandardEchoServer.Accept(l)
}
func BenchmarkStp1C(b *testing.B) {
	w := sync.WaitGroup{}
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
	chandler := &giao.Handler{
		H: func(reqReader proto.Message, respWriter giao.ProtoWriter) {
			w.Done()
		},
		ReqPool: echoPool,
	}
	benchmarkStupidEchoServer.RegFuncWithId(EchoRpc, shandler)
	c, err := client.NewStupidClient("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}
	c.RegFuncWithId(EchoRpc, chandler)
	go func() {
		err := c.StartServe()
		if err != nil {
			panic(err)
		}
	}()
	w.Add(b.N)
	go func() {
		for j := 0; j < b.N; j++ {
			err := c.ACall(EchoRpc, EchoMsg[j%12])
			if err != nil {
				panic(err)
			}
		}
	}()
	b.ResetTimer()
	w.Wait()
}
func BenchmarkStp16C(b *testing.B) {
	w := sync.WaitGroup{}
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
	chandler := &giao.Handler{
		H: func(req proto.Message, respWriter giao.ProtoWriter) {
			w.Done()
		},
		ReqPool: echoPool,
	}
	benchmarkStupidEchoServer.RegFuncWithId(EchoRpc, shandler)
	w.Add(b.N)
	for i := 0; i < 16; i++ {
		c, err := client.NewStupidClient("tcp", "localhost:8888")
		if err != nil {
			panic(err)
		}
		c.RegFuncWithId(EchoRpc, chandler)
		go func() {
			err := c.StartServe()
			if err != nil {
				panic(err)
			}
		}()
		go func(index int) {
			for j := 0; j < b.N/16; j++ {
				msg := &test.Echo{}
				msg.Content = EchoMsg[j%12].Content
				msg.Index = int32(index)
				err := c.ACall(EchoRpc, msg)
				if err != nil {
					panic(err)
				}
			}
		}(i)
	}
	c, err := client.NewStupidClient("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}
	c.RegFuncWithId(EchoRpc, chandler)
	go func() {
		err := c.StartServe()
		if err != nil {
			panic(err)
		}
	}()
	go func(index int) {
		for j := 0; j < b.N%16; j++ {
			msg := &test.Echo{}
			msg.Content = EchoMsg[j%12].Content
			msg.Index = int32(index)
			err := c.ACall(EchoRpc, msg)
			if err != nil {
				panic(err)
			}
		}
	}(16)
	b.ResetTimer()
	w.Wait()
}
func BenchmarkStdSync1C(b *testing.B) {
	c, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		err := c.Call("Echo.GoEcho", EchoMsg[j%12], &test.Echo{})
		if err != nil {
			panic(err)
		}
	}
}
func BenchmarkStdASync1C(b *testing.B) {
	done := make(chan *rpc.Call, b.N)
	count := int32(0)
	inflight := 0
	c, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	for j := 0; j < b.N; j++ {
		runtime.Gosched()
		call := c.Go("Echo.GoEcho", EchoMsg[j%12], &test.Echo{}, done)
		if call.Error != nil {
			panic(call.Error)
		}
		inflight++
	}
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			<-ticker.C
			println(b.N, atomic.LoadInt32(&count), inflight)
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		call := <-done
		atomic.AddInt32(&count, 1)
		if call.Error != nil {
			panic(call.Error)
		}
	}
	ticker.Stop()
}
func BenchmarkStdASync16C(b *testing.B) {
	done := make([]chan *rpc.Call, b.N)
	for i := 0; i < 16; i++ {
		done[i] = make(chan *rpc.Call, 10)
	}
	for i := 0; i < 16; i++ {
		c, err := rpc.Dial("tcp", "localhost:8080")
		if err != nil {
			panic(err)
		}
		go func(index int) {
			for j := 0; j < b.N/16; j++ {
				msg := &test.Echo{}
				msg.Content = EchoMsg[j%12].Content
				msg.Index = int32(index)
				c.Go("Echo.GoEcho", msg, &test.Echo{}, done[index])
				runtime.Gosched()
			}
		}(i)
	}
	c, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	go func(index int) {
		for j := 0; j < b.N%16; j++ {
			msg := &test.Echo{}
			msg.Content = EchoMsg[j%12].Content
			msg.Index = int32(index)
			c.Go("Echo.GoEcho", msg, &test.Echo{}, done[index])
		}
	}(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var call *rpc.Call
		select {
		case call = <-done[0]:
		case call = <-done[1]:
		case call = <-done[2]:
		case call = <-done[3]:
		case call = <-done[4]:
		case call = <-done[5]:
		case call = <-done[6]:
		case call = <-done[7]:
		case call = <-done[8]:
		case call = <-done[9]:
		case call = <-done[0]:
		case call = <-done[11]:
		case call = <-done[12]:
		case call = <-done[13]:
		case call = <-done[14]:
		case call = <-done[15]:
		}
		if call.Error != nil {
			panic(call.Error)
		}
	}
}
