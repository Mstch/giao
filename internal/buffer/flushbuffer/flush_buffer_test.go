package flushbuffer

import (
	"encoding/binary"
	test "github.com/Mstch/giao/test/msg"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestFlushBuffer(t *testing.T) {
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}

	go func() {
		c, err := net.Dial("tcp", "localhost:8888")
		if err != nil {
			panic(err)
		}
		fb := NewFBuffers(1, c)
		go fb.StartFlushTimer()
		for i := 0; i < 1024; i++ {
			go writer(fb, 1024)
		}
	}()

	sc, err := l.Accept()
	if err != nil {
		panic(err)
	}
	idbuf := make([]byte, 4)
	sizebuf := make([]byte, 4)
	for i := 0; i < 1024*1024; i++ {
		_, err = io.ReadFull(sc, idbuf)
		if err != nil {
			panic(err)
		}
		_, err = io.ReadFull(sc, sizebuf)
		if err != nil {
			panic(err)
		}
		size := binary.BigEndian.Uint32(sizebuf)
		buf := make([]byte, size)
		_, err = io.ReadFull(sc, buf)
		if err != nil {
			panic(err)
		}
		echo := &test.Echo{}
		err = echo.Unmarshal(buf)
		if err != nil {
			panic(err)
		}
		if len(echo.Content) != int(echo.Index) {
			panic(echo.Index)
		}
	}

}
func BenchmarkFBuffers(b *testing.B) {
	b.N = 16 * (b.N/16 + 1)
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}

	go func() {
		c, err := net.Dial("tcp", "localhost:8888")
		if err != nil {
			panic(err)
		}
		fb := NewFBuffers(1*time.Microsecond, c)
		go fb.StartFlushTimer()
		for i := 0; i < 16; i++ {
			go writer(fb, b.N/16)
		}
	}()

	sc, err := l.Accept()
	if err != nil {
		panic(err)
	}
	idbuf := make([]byte, 4)
	sizebuf := make([]byte, 4)
	echo := &test.Echo{}
	for i := 0; i < b.N; i++ {
		_, err = io.ReadFull(sc, idbuf)
		if err != nil {
			panic(err)
		}
		_, err = io.ReadFull(sc, sizebuf)
		if err != nil {
			panic(err)
		}
		size := binary.BigEndian.Uint32(sizebuf)
		buf := make([]byte, size)
		_, err = io.ReadFull(sc, buf)
		if err != nil {
			panic(err)
		}
		err = echo.Unmarshal(buf)
		if err != nil {
			panic(err)
		}
		if len(echo.Content) != int(echo.Index) {
			panic(echo.Index)
		}
	}
}
func writer(fb *FBuffers, n int) {
	for i := 0; i < n; i++ {
		msg := &test.Echo{
			Index:   int32(i + 1),
			Content: make([]byte, i+1),
		}
		err := fb.Write(msg, 0)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkPool(b *testing.B) {
	b.N = 1024 * (b.N/1024 + 1)
	pool := &sync.Pool{New: func() interface{} {
		return 1
	}}
	wg := &sync.WaitGroup{}
	wg.Add(1024)
	for i := 0; i < 1024; i++ {
		go func() {
			for i := 0; i < b.N/1024; i++ {
				pool.Get()
				pool.Put(i)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
