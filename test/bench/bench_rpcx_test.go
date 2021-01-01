package bench

import (
	"context"
	test "github.com/Mstch/giao/test/msg"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"testing"
)

type Fuck int

func (t *Fuck) Echo(ctx context.Context, args *test.Echo, reply *test.Echo) error {
	reply.Content = args.Content
	reply.Index = args.Index
	return nil
}
func BenchmarkRpcx1C(b *testing.B) {
	b.SetBytes(5462 * 2)
	s := server.NewServer()
	//s.RegisterName("Arith", new(example.Arith), "")
	s.Register(new(Fuck), "")
	go func() {
		err := s.Serve("tcp", ":8080")
		if err != nil && err != server.ErrServerClosed {
			panic(err)
		}
	}()
	option := client.DefaultOption
	option.SerializeType = protocol.ProtoBuffer
	d, err := client.NewPeer2PeerDiscovery("tcp@localhost:8080", "")
	if err != nil {
		panic(err)
	}
	xclient := client.NewXClient("Fuck", client.Failtry, client.RandomSelect, d, option)
	defer xclient.Close()
	done := make(chan *client.Call, b.N)
	go func() {
		for i := 0; i < b.N; i++ {
			_, err = xclient.Go(context.Background(), "Echo", EchoMsg[i%12], echoPool.Get().(*test.Echo), done)
			if err != nil {
				panic(err)
			}
		}
	}()
	for i := 0; i < b.N; i++ {
		<-done
	}
	err = xclient.Close()
	if err != nil {
		panic(err)
	}
	err = s.Close()
	if err != nil {
		panic(err)
	}
}
func BenchmarkRpcx16C(b *testing.B) {
	b.N = 16 * (b.N/16 + 1)
	b.SetBytes(5462 * 2)
	s := server.NewServer()
	//s.RegisterName("Arith", new(example.Arith), "")
	s.Register(new(Fuck), "")
	go func() {
		err := s.Serve("tcp", ":8080")
		if err != nil && err != server.ErrServerClosed {
			panic(err)
		}
	}()
	option := client.DefaultOption
	option.SerializeType = protocol.ProtoBuffer
	d, err := client.NewPeer2PeerDiscovery("tcp@localhost:8080", "")
	if err != nil {
		panic(err)
	}
	done := make(chan *client.Call, b.N)
	for i := 0; i < 16; i++ {
		xclient := client.NewXClient("Fuck", client.Failtry, client.RandomSelect, d, option)
		defer xclient.Close()
		go func() {
			for i := 0; i < b.N/16; i++ {
				_, err = xclient.Go(context.Background(), "Echo", EchoMsg[i%12], echoPool.Get().(*test.Echo), done)
				if err != nil {
					panic(err)
				}
			}
		}()
	}
	for i := 0; i < b.N; i++ {
		<-done
	}
	err = s.Close()
	if err != nil {
		panic(err)
	}
}
