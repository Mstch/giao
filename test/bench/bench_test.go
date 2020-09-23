package bench

import (
	"context"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/client"
	test "github.com/Mstch/giao/test/msg"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"net/rpc"
	"sync"
	"testing"
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

func BenchmarkStp1C(b *testing.B) {
	b.SetBytes(343 * 2)
	w := sync.WaitGroup{}
	chandler := &giao.Handler{
		H: func(req proto.Message, respWriter giao.ProtoWriter) {
			w.Done()
		},
		ReqPool: echoPool,
	}
	w.Add(b.N)
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
	b.ResetTimer()
	go func() {
		for j := 0; j < b.N; j++ {
			err := c.Go(EchoRpc, EchoMsg[j%12])
			if err != nil {
				panic(err)
			}
		}
	}()
	w.Wait()
}
func BenchmarkStp16C(b *testing.B) {
	b.SetBytes(343 * 2)
	w := sync.WaitGroup{}
	chandler := &giao.Handler{
		H: func(req proto.Message, respWriter giao.ProtoWriter) {
			w.Done()
		},
		ReqPool: echoPool,
	}
	w.Add(b.N)
	for i := 0; i < 16; i++ {
		c, err := client.NewStupidClient().RegWithId(EchoRpc, chandler).Connect("tcp", "localhost:8888")
		if err != nil {
			panic(err)
		}
		go func(c giao.Client) {
			err := c.Serve()
			if err != nil {
				panic(err)
			}
		}(c)
		go func(index int) {
			for j := 0; j < b.N/16; j++ {
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
	b.ResetTimer()
	go func(index int) {
		for j := 0; j < b.N%16; j++ {
			msg := &test.Echo{}
			msg.Content = EchoMsg[j%12].Content
			msg.Index = int32(index)
			err := c.Go(EchoRpc, msg)
			if err != nil {
				panic(err)
			}
		}
	}(16)
	w.Wait()
}
func BenchmarkSyncStd1C(b *testing.B) {
	b.SetBytes(343 * 2)
	c, err := rpc.Dial("tcp", "localhost:8080")
	w := sync.WaitGroup{}
	w.Add(b.N)
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		go func(k int) {
			err := c.Call("StandardEcho.DoEcho", EchoMsg[k%12], &test.Echo{})
			if err != nil {
				panic(err)
			}
			w.Done()
		}(j)
	}
	w.Wait()
}
func BenchmarkSyncStd16C(b *testing.B) {
	b.SetBytes(343 * 2)
	w := sync.WaitGroup{}
	w.Add(b.N)
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
				err := c.Call("StandardEcho.DoEcho", msg, &test.Echo{})
				if err != nil {
					panic(err)
				}
				w.Done()
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
			err := c.Call("StandardEcho.DoEcho", msg, &test.Echo{})
			if err != nil {
				panic(err)
			}
			w.Done()
		}
	}(0)
	b.ResetTimer()
	w.Wait()
}

func BenchmarkGrpc1C(b *testing.B) {
	b.SetBytes(343 * 2)
	conn, err := grpc.Dial("localhost:8181", grpc.WithInsecure())
	c := test.NewEchoServiceClient(conn)
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		_, err := c.DoEcho(context.Background(), EchoMsg[j%12])
		if err != nil {
			panic(err)
		}
	}
	err = conn.Close()
	if err != nil {
		panic(err)
	}
}
func BenchmarkGrpc16C(b *testing.B) {
	w := sync.WaitGroup{}
	w.Add(b.N)
	b.SetBytes(343 * 2)
	for i := 0; i < 16; i++ {
		conn, err := grpc.Dial("localhost:8181", grpc.WithInsecure())
		c := test.NewEchoServiceClient(conn)
		if err != nil {
			panic(err)
		}
		go func(index int) {
			for j := 0; j < b.N/16; j++ {
				msg := &test.Echo{}
				msg.Content = EchoMsg[j%12].Content
				msg.Index = int32(index)
				_, err := c.DoEcho(context.Background(), msg)
				if err != nil {
					panic(err)
				}
				w.Done()
			}
		}(i)
	}
	conn, err := grpc.Dial("localhost:8181", grpc.WithInsecure())
	c := test.NewEchoServiceClient(conn)
	if err != nil {
		panic(err)
	}
	go func(index int) {
		for j := 0; j < b.N%16; j++ {
			msg := &test.Echo{}
			msg.Content = EchoMsg[j%12].Content
			msg.Index = int32(index)
			_, err := c.DoEcho(context.Background(), msg)
			if err != nil {
				panic(err)
			}
			w.Done()
		}
	}(0)
	w.Wait()
}

//func BenchmarkAStd1C(b *testing.B) {
//	done := make(chan *rpc.Call, 1024)
//	c, err := rpc.Dial("tcp", "localhost:8080")
//	if err != nil {
//		panic(err)
//	}
//	go func() {
//		for j := 0; j < b.N; j++ {
//			call := c.Go("StandardEcho.DoEcho", EchoMsg[j%12], &test.StandardEcho{}, done)
//			if call.Error != nil {
//				panic(call.Error)
//			}
//		}
//	}()
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		<-done
//	}
//}
//func BenchmarkAStd16C(b *testing.B) {
//	done := make([]chan *rpc.Call, 16)
//	for i := 0; i < 16; i++ {
//		done[i] = make(chan *rpc.Call, 1024)
//	}
//	for i := 0; i < 16; i++ {
//		c, err := rpc.Dial("tcp", "localhost:8080")
//		if err != nil {
//			panic(err)
//		}
//		go func(index int) {
//			for j := 0; j < b.N/16; j++ {
//				msg := &test.StandardEcho{}
//				msg.Content = EchoMsg[j%12].Content
//				msg.Index = int32(index)
//				c.Go("StandardEcho.DoEcho", msg, &test.StandardEcho{}, done[index])
//
//			}
//		}(i)
//	}
//	c, err := rpc.Dial("tcp", "localhost:8080")
//	if err != nil {
//		panic(err)
//	}
//	go func(index int) {
//		for j := 0; j < b.N%16; j++ {
//			msg := &test.StandardEcho{}
//			msg.Content = EchoMsg[j%12].Content
//			msg.Index = int32(index)
//			c.Go("StandardEcho.DoEcho", msg, &test.StandardEcho{}, done[index])
//
//		}
//	}(0)
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		var call *rpc.Call
//		select {
//		case call = <-done[0]:
//		case call = <-done[1]:
//		case call = <-done[2]:
//		case call = <-done[3]:
//		case call = <-done[4]:
//		case call = <-done[5]:
//		case call = <-done[6]:
//		case call = <-done[7]:
//		case call = <-done[8]:
//		case call = <-done[9]:
//		case call = <-done[0]:
//		case call = <-done[11]:
//		case call = <-done[12]:
//		case call = <-done[13]:
//		case call = <-done[14]:
//		case call = <-done[15]:
//		}
//		if call.Error != nil {
//			panic(call.Error)
//		}
//	}
//}
