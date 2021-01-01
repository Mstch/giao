package bench

import (
	test "github.com/Mstch/giao/test/msg"
	"github.com/hprose/hprose-golang/rpc"
	"testing"
	"time"
)

func Say(in []byte) ([]byte, error) {
	args := echoPool.Get().(*test.Echo)
	err := args.Unmarshal(in)
	if err != nil {
		panic(err)
	}
	return args.Marshal()
}

type HproseEcho struct {
	Say func(args []byte) (reply []byte, err error) `simple:"true"`
}

func BenchmarkHprose1C(b *testing.B) {
	b.SetBytes(5462 * 2)
	server := rpc.NewTCPServer("tcp://localhost:8080")
	server.AddFunction("say", Say, rpc.Options{Simple: true})
	go func() {
		err := server.Start()
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(10*time.Millisecond)
	b.ResetTimer()
	c := rpc.NewTCPClient("tcp://localhost:8080")
	hEcho := &HproseEcho{}
	c.UseService(hEcho)
	for i := 0; i < b.N; i++ {
		buf, err := EchoMsg[i%12].Marshal()
		if err != nil {
			panic(err)
		}
		respBuf, err := hEcho.Say(buf)
		if err != nil {
			panic(err)
		}
		err = (&test.Echo{}).Unmarshal(respBuf)
		if err != nil {
			panic(err)
		}
	}
	server.Close()
}

func BenchmarkHprose16C(b *testing.B) {
	b.N = 16 * (b.N/16 + 1)
	b.SetBytes(5462 * 2)
	server := rpc.NewTCPServer("tcp://localhost:8080")
	server.AddFunction("say", Say, rpc.Options{Simple: true})
	go func() {
		err := server.Start()
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(10*time.Millisecond)
	b.ResetTimer()
	done := make(chan bool, 16)
	for i := 0; i < 16; i++ {
		go func() {
			c := rpc.NewTCPClient("tcp://localhost:8080")
			hEcho := &HproseEcho{}
			c.UseService(hEcho)
			for i := 0; i < b.N/16; i++ {
				buf, err := EchoMsg[i%12].Marshal()
				if err != nil {
					panic(err)
				}
				respBuf, err := hEcho.Say(buf)
				if err != nil {
					panic(err)
				}
				err = (&test.Echo{}).Unmarshal(respBuf)
				if err != nil {
					panic(err)
				}
			}
			done <- false
		}()
	}
	for i := 0; i < 16; i++ {
		<-done
	}
	server.Close()
}
