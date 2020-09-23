package qps

import (
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/client"
	test "github.com/Mstch/giao/test/msg"
	"github.com/gogo/protobuf/proto"
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

type Echo struct {
}

func (e *Echo) GoEcho(req *test.Echo, resp *test.Echo) error {
	resp = &test.Echo{}
	resp.Content = req.Content
	return nil
}

func TestQPSStp1C(t *testing.T) {
	w := sync.WaitGroup{}
	w.Add(1e9)
	chandler := &giao.Handler{
		H: func(reqReader proto.Message, respWriter giao.ProtoWriter) {
			w.Done()
		},
		ReqPool: echoPool,
	}
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
	go func() {
		for j := 0; j < 1e9; j++ {
			err := c.Go(EchoRpc, EchoMsg[j%12])
			if err != nil {
				panic(err)
			}
		}
	}()
	start := time.Now().UnixNano()
	w.Wait()
	println("[STUPID] 1C QPS:", int(1e9/(float64(time.Now().UnixNano()-start)/1e9)))
}
func TestQPSStp16C(t *testing.T) {
	count := make([]int, 16)
	chandler := &giao.Handler{
		H: func(req proto.Message, respWriter giao.ProtoWriter) {
			count[req.(*test.Echo).Index] ++
		},
		ReqPool: echoPool,
	}
	for i := 0; i < 16; i++ {
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
		go func(index int) {
			for j := 0; ; j++ {
				msg := &test.Echo{}
				msg.Content = EchoMsg[j%12].Content
				msg.Index = int32(index)
				err := c.Go(EchoRpc, msg)
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
	println("[STUPID] 16C QPS:", countCount/10, "now memory :", memStat.HeapInuse/1E6, "mb")
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
			runtime.Gosched()
		}
	}()
	<-time.NewTimer(10 * time.Second).C
	memStat := &runtime.MemStats{}
	runtime.ReadMemStats(memStat)
	println("[STANDARD] 1C QPS:", count/10, memStat.HeapInuse/1E6, "mb")
}
func TestQPSStdAsync1C(t *testing.T) {
	done := make(chan *rpc.Call, 10)
	c, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	count := 0
	go func() {
		for j := 0; ; j++ {
			c.Go("Echo.GoEcho", EchoMsg[j%12], &test.Echo{}, done)
			runtime.Gosched()
		}
	}()
	go func() {
		for {
			call := <-done
			count++
			if call.Error != nil {
				panic(call.Error)
			}
		}
	}()
	<-time.NewTimer(10 * time.Second).C
	memStat := &runtime.MemStats{}
	runtime.ReadMemStats(memStat)
	println("[STANDARD] 1C QPS:", count/10, memStat.HeapInuse/1E6, "mb")
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
	println("[STANDARD] 16C QPS:", count/10, memStat.HeapInuse/1E6, "mb")
}
