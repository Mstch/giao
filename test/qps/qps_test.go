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
func TestQPSStp1C(t *testing.T) {
	count := 0
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
	memStat := &runtime.MemStats{}
	runtime.ReadMemStats(memStat)
	println("[STUPID] 1C QPS:", count/10, "now memory :", memStat.TotalAlloc/1E6, "mb")
}
func TestQPSStp16C(t *testing.T) {
	count := make([]int, 16)
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
			count[req.(*test.Echo).Index] ++
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
	memStat := &runtime.MemStats{}
	runtime.ReadMemStats(memStat)
	println("[STUPID] 16C QPS:", countCount/10, "now memory :", memStat.TotalAlloc/1E6, "mb")
}
func TestQPSStd1C(t *testing.T) {
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
	memStat := &runtime.MemStats{}
	runtime.ReadMemStats(memStat)
	println("[STANDARD] 1C QPS:", count/10, memStat.TotalAlloc/1E6, "mb")
}
func TestQPSStd16C(t *testing.T) {
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
	memStat := &runtime.MemStats{}
	runtime.ReadMemStats(memStat)
	println("[STANDARD] 16C QPS:", count/10, memStat.TotalAlloc/1E6, "mb")
}
