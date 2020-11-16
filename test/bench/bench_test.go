package bench

import (
	"context"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/client"
	"github.com/Mstch/giao/internal/server"
	test "github.com/Mstch/giao/test/msg"
	"google.golang.org/grpc"
	"net/rpc"
	"strings"
	"sync"
	"testing"
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
	b.SetBytes(5462 * 2)
	w := sync.WaitGroup{}
	w.Add(b.N)
	shandler := &giao.Handler{
		H: func(req giao.Msg, session giao.Session) {
			respMsg := &test.Echo{}
			respMsg.Content = req.(*test.Echo).Content
			err := session.Write(EchoRpc, respMsg)
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
	go func() {
		for j := 0; j < b.N; j++ {
			err := c.Go(EchoRpc, EchoMsg[j%12])
			if err != nil {
				if !strings.HasSuffix(err.Error(), "use of closed network connection") {
					panic(err)
				}
			}
		}
	}()
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
	w := sync.WaitGroup{}
	shandler := &giao.Handler{
		H: func(req giao.Msg, session giao.Session) {
			respMsg := &test.Echo{}
			respMsg.Content = req.(*test.Echo).Content
			err := session.Write(EchoRpc, respMsg)
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
	w.Add(b.N)
	for i := 0; i < 16; i++ {
		c, err := client.NewStupidClient().RegWithId(EchoRpc, chandler).Connect("tcp", "localhost:8888")
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
				msg := &test.Echo{}
				msg.Content = EchoMsg[j%12].Content
				msg.Index = int32(index)
				err := c.Go(EchoRpc, msg)
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
			msg := &test.Echo{}
			msg.Content = EchoMsg[j%12].Content
			msg.Index = int32(index)
			err := c.Go(EchoRpc, msg)
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
	err = benchmarkStupidEchoServer.Shutdown()
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
}
func BenchmarkSyncStd1C(b *testing.B) {
	b.SetBytes(5462 * 2)
	c, err := rpc.Dial("tcp", "localhost:8080")
	w := sync.WaitGroup{}
	w.Add(b.N)
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		go func(k int) {
			err := c.Call("StandardEcho.DoEcho", EchoMsg[k%12], &test.Echo{})
			if err != nil {
				if !strings.HasSuffix(err.Error(), "use of closed network connection") {
					panic(err)
				}
			}
			w.Done()
		}(j)
	}
	w.Wait()
}
func BenchmarkSyncStd16C(b *testing.B) {
	b.SetBytes(5462 * 2)
	w := sync.WaitGroup{}
	w.Add(b.N)
	for i := 0; i < 16; i++ {
		c, err := rpc.Dial("tcp", "localhost:8080")
		if err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				panic(err)
			}
		}
		go func(index int) {
			for j := 0; j < b.N/16; j++ {
				msg := &test.Echo{}
				msg.Content = EchoMsg[j%12].Content
				msg.Index = int32(index)
				err := c.Call("StandardEcho.DoEcho", msg, &test.Echo{})
				if err != nil {
					if !strings.HasSuffix(err.Error(), "use of closed network connection") {
						panic(err)
					}
				}
				w.Done()
			}
		}(i)
	}
	c, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
	go func(index int) {
		for j := 0; j < b.N%16; j++ {
			msg := &test.Echo{}
			msg.Content = EchoMsg[j%12].Content
			msg.Index = int32(index)
			err := c.Call("StandardEcho.DoEcho", msg, &test.Echo{})
			if err != nil {
				if !strings.HasSuffix(err.Error(), "use of closed network connection") {
					panic(err)
				}
			}
			w.Done()
		}
	}(0)
	b.ResetTimer()
	w.Wait()
}

func BenchmarkGrpc1C(b *testing.B) {
	b.SetBytes(5462 * 2)
	conn, err := grpc.Dial("localhost:8181", grpc.WithInsecure())
	c := test.NewEchoServiceClient(conn)
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		_, err := c.DoEcho(context.Background(), EchoMsg[j%12])
		if err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				panic(err)
			}
		}
	}
	err = conn.Close()
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
}
func BenchmarkGrpc16C(b *testing.B) {
	w := sync.WaitGroup{}
	w.Add(b.N)
	b.SetBytes(5462 * 2)
	for i := 0; i < 16; i++ {
		conn, err := grpc.Dial("localhost:8181", grpc.WithInsecure())
		c := test.NewEchoServiceClient(conn)
		if err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				panic(err)
			}
		}
		go func(index int) {
			for j := 0; j < b.N/16; j++ {
				msg := &test.Echo{}
				msg.Content = EchoMsg[j%12].Content
				msg.Index = int32(index)
				_, err := c.DoEcho(context.Background(), msg)
				if err != nil {
					if !strings.HasSuffix(err.Error(), "use of closed network connection") {
						panic(err)
					}
				}
				w.Done()
			}
		}(i)
	}
	conn, err := grpc.Dial("localhost:8181", grpc.WithInsecure())
	c := test.NewEchoServiceClient(conn)
	if err != nil {
		if !strings.HasSuffix(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}
	go func(index int) {
		for j := 0; j < b.N%16; j++ {
			msg := &test.Echo{}
			msg.Content = EchoMsg[j%12].Content
			msg.Index = int32(index)
			_, err := c.DoEcho(context.Background(), msg)
			if err != nil {
				if !strings.HasSuffix(err.Error(), "use of closed network connection") {
					panic(err)
				}
			}
			w.Done()
		}
	}(0)
	w.Wait()
}
