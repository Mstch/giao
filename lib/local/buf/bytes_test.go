package buf_test

import (
	"context"
	"encoding/binary"
	"github.com/Mstch/giao/lib/local/buf"
	"github.com/Mstch/giao/mock"
	"io"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}
func RandStringRunes() string {
	b := make([]rune, 64+rand.Intn(1024))
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type TestMarshallAble []byte

func (t *TestMarshallAble) Unmarshal(data []byte) error {
	return nil
}

func (t *TestMarshallAble) MarshalTo(b []byte) (int, error) {
	copy(b, *t)
	return len(*t), nil
}
func (t *TestMarshallAble) Size() int {
	return len(*t)
}

func TestBytesBuf(t *testing.T) {
	tests := 10000
	check := make([]int32, tests)
	r, w, err := mock.NewTcpConn()
	if err != nil {
		panic(err)
	}
	bytesBuf := buf.NewBytesBuf(1128, 1*time.Second, w, context.Background())
	go func() {
		err := bytesBuf.StartFlush()
		if err != nil {
			panic(err)
		}
	}()
	for i := 0; i < tests; i++ {
		go func(i int) {
			m := TestMarshallAble(RandStringRunes())
			header := make([]byte, 8)
			binary.BigEndian.PutUint32(header, uint32(i))
			binary.BigEndian.PutUint32(header[4:], uint32(m.Size()))
			err := bytesBuf.Write(header, &m)
			if err != nil {
				panic(err)
			}
		}(i)
	}
	for i := 0; i < tests; i++ {
		header := make([]byte, 8)
		_, err := io.ReadFull(r, header)
		if err != nil {
			panic(err)
		}
		tag := binary.BigEndian.Uint32(header)
		if ok := atomic.CompareAndSwapInt32(&check[tag], 0, 1); !ok {
			panic(tag)
		}
		length := binary.BigEndian.Uint32(header[4:])
		b := make([]byte, length)
		_, err = io.ReadFull(r, b)
		if err != nil {
			panic(err)
		}
	}
	for i, v := range check {
		if v != 1 {
			panic(i)
		}
	}
	bytesBuf.Shutdown()
	err = w.Close()
	if err != nil {
		panic(err)
	}
	err = r.Close()
	if err != nil {
		panic(err)
	}
}

func BenchmarkBytesBuf(b *testing.B) {
	bufSize := 4 * 1024 * 1024
	gos := runtime.GOMAXPROCS(0) * 2048
	b.N = (b.N/gos + 1) * gos
	packet := 4096
	r, w, err := mock.NewTcpConn()
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(packet))
	ctx, cancel := context.WithCancel(context.Background())
	bytesBuf := buf.NewBytesBuf(bufSize, 1*time.Millisecond, w, ctx)
	writeDone := make(chan struct{})
	go func() {
		flushDone := make(chan struct{})
		tag := []byte{0, 1, 2, 3}
		buff := make([]byte, 4092)
		for i := range buff {
			buff[i] = byte(i + 4)
		}
		wwg := &sync.WaitGroup{}
		for i := 0; i < gos; i++ {
			wwg.Add(1)
			go func() {
				for i := 0; i < b.N/gos; i++ {
					m := TestMarshallAble(buff)
					err := bytesBuf.Write(tag, &m)
					if err != nil {
						panic(err)
					}
				}
				wwg.Done()
			}()
		}
		go func() {
			err := bytesBuf.StartFlush()
			if err != nil {
				panic(err)
			}
			flushDone <- struct{}{}
		}()
		wwg.Wait()
		<-flushDone
		writeDone <- struct{}{}
	}()
	buff := make([]byte, 4*1024*1024)
	readed := 0
	for readed < packet*b.N {
		rr, err := r.Read(buff)
		if err != nil {
			panic(err)
		}
		readed += rr
	}
	err = w.Close()
	if err != nil {
		panic(err)
	}
	err = r.Close()
	if err != nil {
		panic(err)
	}
	cancel()
	bytesBuf.Shutdown()
	<-writeDone
}

func BenchmarkStdWrite(b *testing.B) {
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	packet := 4 * 1024
	b.SetBytes(int64(packet))
	go func() {
		bytes := make([]byte, packet)
		c, err := net.Dial("tcp", "localhost:8888")
		if err != nil {
			panic(err)
		}
		for i := 0; i < b.N; i++ {
			_, err := c.Write(bytes)
			if err != nil {
				panic(err)
			}
		}
	}()
	c, err := l.Accept()
	if err != nil {
		panic(err)
	}
	bytes := make([]byte, packet)
	for totalLen := 0; totalLen < b.N*packet; {
		length, err := c.Read(bytes)
		if err != nil {
			panic(err)
		}
		totalLen += length
	}
}
