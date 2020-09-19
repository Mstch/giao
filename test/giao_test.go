package test

import (
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/client"
	"github.com/Mstch/giao/internal/server"
	test "github.com/Mstch/giao/test/msg"
	"net"
	"net/rpc"
	"sync"
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
		EchoMsg[i].Content = str
	}
}

var echoPool = &sync.Pool{New: func() interface{} {
	return &test.Echo{}
}}

func TestEcho(t *testing.T) {
	w := sync.WaitGroup{}
	shandler := &giao.Handler{
		H: func(reqReader giao.ProtoReader, respWriter giao.ProtoWriter) {
			reqMsg := reqReader()
			println(reqMsg.(giao.PB).Size())
			err := respWriter(reqMsg)
			if err != nil {
				panic(err)
			}
			w.Done()
		},
		ReqPool: echoPool,
	}
	chandler := &giao.Handler{
		H: func(reqReader giao.ProtoReader, respWriter giao.ProtoWriter) {
			reqReader()
			w.Done()
		},
		ReqPool: echoPool,
	}
	w.Add(24)
	echoServer, err := server.NewStupidServer("tcp", ":8880")
	if err != nil {
		panic(err)
	}
	echoServer.RegFuncWithId(EchoRpc, shandler)
	go func() {
		err := echoServer.StartServe()
		if err != nil {
			panic(err)
		}
	}()
	c, err := client.NewStupidClient("tcp", "localhost:8880")
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
	for i := 0; i < 12; i++ {
		go func(i int) {
			err := c.ACall(EchoRpc, EchoMsg[i])
			if err != nil {
				panic(err)
			}
		}(i)
	}
	w.Wait()
}

/**
goos: darwin
goarch: amd64
pkg: github.com/Mstch/giao/test
BenchmarkRpcStupid1C-16           113473              9360 ns/op            1152 B/op         10 allocs/op
BenchmarkRpcStupid1C-16           112281             10188 ns/op             979 B/op          9 allocs/op
BenchmarkRpcStupid1C-16           110148             10647 ns/op            1041 B/op          9 allocs/op
BenchmarkRpcStupid16C-16           10044            160715 ns/op           30046 B/op        169 allocs/op
BenchmarkRpcStupid16C-16            8600            189177 ns/op           22323 B/op        162 allocs/op
BenchmarkRpcStupid16C-16            8826            156581 ns/op           26255 B/op        168 allocs/op
BenchmarkRpcStandard1C-16         107690             13218 ns/op             603 B/op          8 allocs/op
BenchmarkRpcStandard1C-16         111004             14724 ns/op             577 B/op          8 allocs/op
BenchmarkRpcStandard1C-16          89536             12569 ns/op             584 B/op          8 allocs/op
BenchmarkRpcStandard16C-16          6992            202737 ns/op            9830 B/op        128 allocs/op
BenchmarkRpcStandard16C-16          6751            211441 ns/op            9848 B/op        129 allocs/op
BenchmarkRpcStandard16C-16          5496            184521 ns/op            9876 B/op        128 allocs/op
PASS
ok      github.com/Mstch/giao/test      21.266s
*/
var benchmarkStupidEchoServer giao.Server
var benchmarkStandardEchoServer *rpc.Server

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
func BenchmarkRpcStupid1C(b *testing.B) {
	w := sync.WaitGroup{}
	shandler := &giao.Handler{
		H: func(reqReader giao.ProtoReader, respWriter giao.ProtoWriter) {
			reqMsg := reqReader()
			respMsg := &test.Echo{}
			respMsg.Content = reqMsg.(*test.Echo).Content
			err := respWriter(respMsg)
			if err != nil {
				panic(err)
			}
			w.Done()
		},
		ReqPool: echoPool,
	}
	chandler := &giao.Handler{
		H: func(reqReader giao.ProtoReader, respWriter giao.ProtoWriter) {
			reqReader()
			w.Done()
		},
		ReqPool: echoPool,
	}
	w.Add(b.N * 2)
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
	go func(n int) {
		for j := 0; j < n; j++ {
			c.ACall(EchoRpc, EchoMsg[j%12])
		}
	}(b.N)
	w.Wait()
}
func BenchmarkRpcStupid16C(b *testing.B) {
	w := sync.WaitGroup{}
	shandler := &giao.Handler{
		H: func(reqReader giao.ProtoReader, respWriter giao.ProtoWriter) {
			reqMsg := reqReader()
			respMsg := &test.Echo{}
			respMsg.Content = reqMsg.(*test.Echo).Content
			err := respWriter(respMsg)
			if err != nil {
				panic(err)
			}
			w.Done()
		},
		ReqPool: echoPool,
	}
	chandler := &giao.Handler{
		H: func(reqReader giao.ProtoReader, respWriter giao.ProtoWriter) {
			reqReader()
			w.Done()
		},
		ReqPool: echoPool,
	}
	w.Add(b.N * 32)
	for i := 0; i < 16; i++ {
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
		go func(n int) {
			for j := 0; j < n; j++ {
				c.ACall(EchoRpc, EchoMsg[j%12])
			}
		}(b.N)
	}
	w.Wait()
}

type Echo struct {
}

func (e *Echo) GoEcho(req *test.Echo, resp *test.Echo) error {
	resp = &test.Echo{}
	resp.Content = req.Content
	return nil
}
func BenchmarkRpcStandard1C(b *testing.B) {
	c, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	doneCh := make(chan *rpc.Call, 1024)
	go func(n int) {
		for j := 0; j < n; j++ {
			c.Go("Echo.GoEcho", EchoMsg[j%12], &Echo{}, doneCh)
		}
	}(b.N)
	for i := 0; i < b.N; i++ {
		<-doneCh
	}
}
func BenchmarkRpcStandard16C(b *testing.B) {
	doneCh := make(chan *rpc.Call, b.N*16)
	for i := 0; i < 16; i++ {
		c, err := rpc.Dial("tcp", "localhost:8080")
		if err != nil {
			panic(err)
		}
		go func(n int) {
			for j := 0; j < n; j++ {
				c.Go("Echo.GoEcho", EchoMsg[j%12], &Echo{}, doneCh)
			}
		}(b.N)
	}
	for i := 0; i < b.N*16; i++ {
		<-doneCh
	}
}

func TestQPSStp1C(t *testing.T) {
	<-time.NewTimer(5 * time.Second).C
	count := 0
	shandler := &giao.Handler{
		H: func(reqReader giao.ProtoReader, respWriter giao.ProtoWriter) {
			reqMsg := reqReader()
			respMsg := &test.Echo{}
			respMsg.Content = reqMsg.(*test.Echo).Content
			err := respWriter(respMsg)
			if err != nil {
				panic(err)
			}
		},
		ReqPool: echoPool,
	}
	chandler := &giao.Handler{
		H: func(reqReader giao.ProtoReader, respWriter giao.ProtoWriter) {
			reqReader()
			count++
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
	go func() {
		for j := 0; ; j++ {
			err := c.ACall(EchoRpc, EchoMsg[j%12])
			if err != nil {
				panic(err)
			}
		}
	}()
	<-time.NewTimer(10 * time.Second).C
	println("[STUPID] 1C QPS:", count/10)
}
func TestQPSStp16C(t *testing.T) {
	<-time.NewTimer(5 * time.Second).C
	count := make([]int, 16)
	shandler := &giao.Handler{
		H: func(reqReader giao.ProtoReader, respWriter giao.ProtoWriter) {
			reqMsg := reqReader()
			respMsg := &test.Echo{}
			respMsg.Content = reqMsg.(*test.Echo).Content
			err := respWriter(respMsg)
			if err != nil {
				panic(err)
			}
		},
		ReqPool: echoPool,
	}
	chandler := &giao.Handler{
		H: func(reqReader giao.ProtoReader, respWriter giao.ProtoWriter) {
			count[reqReader().(*test.Echo).Index] ++
		},
		ReqPool: echoPool,
	}
	benchmarkStupidEchoServer.RegFuncWithId(EchoRpc, shandler)
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
			for j := 0; ; j++ {
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
	<-time.NewTimer(10 * time.Second).C
	countCount := 0
	for i := 0; i < 16; i++ {
		countCount += count[i]
	}
	println("[STUPID] 16C QPS:", countCount/10)
}
func TestQPSStd1C(t *testing.T) {
	<-time.NewTimer(10 * time.Second).C
	c, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	count := 0
	go func() {
		for j := 0; ; j++ {
			err := c.Call("Echo.GoEcho", EchoMsg[j%12], &test.Echo{})
			if err != nil {
				panic(err)
			}
			count++
		}
	}()
	<-time.NewTimer(10 * time.Second).C
	println("[STANDARD] 1C QPS:", count/10)
}
func TestQPSStd16C(t *testing.T) {
	<-time.NewTimer(10 * time.Second).C
	done := make([]chan *rpc.Call, 16)
	for i := 0; i < 16; i++ {
		done[i] = make(chan *rpc.Call, 1<<15)
	}
	for i := 0; i < 16; i++ {
		c, err := rpc.Dial("tcp", "localhost:8080")
		if err != nil {
			panic(err)
		}
		go func(index int) {
			for j := 0; ; j++ {
				msg := &test.Echo{}
				msg.Content = EchoMsg[j%12].Content
				msg.Index = int32(index)
				c.Go("Echo.GoEcho", msg, &test.Echo{}, done[index])
			}
		}(i)
	}
	count := 0
	go func() {
		for {
			select {
			case <-done[0]:
			case <-done[1]:
			case <-done[2]:
			case <-done[3]:
			case <-done[4]:
			case <-done[5]:
			case <-done[6]:
			case <-done[7]:
			case <-done[8]:
			case <-done[9]:
			case <-done[0]:
			case <-done[11]:
			case <-done[12]:
			case <-done[13]:
			case <-done[14]:
			case <-done[15]:
			}
			count++
		}
	}()
	<-time.NewTimer(10 * time.Second).C
	println("[STANDARD] 16C QPS:", count/10)
}
