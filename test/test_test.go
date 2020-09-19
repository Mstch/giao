package test

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/Mstch/giao/common"
	"github.com/Mstch/giao/internal/buffer"
	"io"
	"io/ioutil"
	"net"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
	_ "unsafe"
)

//go:linkname memstat runtime.mstats
var mstat runtime.MemStats

func calDelay(f func()) int {
	start := time.Now().Nanosecond()
	f()
	return time.Now().Nanosecond() - start
}

func BenchmarkWaitGroup(b *testing.B) {
	wp := make([]*sync.WaitGroup, common.GoMaxProc)
	done := make(chan bool, common.GoMaxProc)
	for i := 0; i < common.GoMaxProc; i++ {
		wp[i] = &sync.WaitGroup{}
		wp[i].Add(100 * b.N)
		go func(w *sync.WaitGroup) {
			w.Wait()
			done <- true
		}(wp[i])
	}
	b.ResetTimer()
	for i := 0; i < 100; i++ {
		go func() {
			for i := 0; i < b.N; i++ {
				pid := runtime_procPin()
				runtime_procUnpin()
				wp[pid].Done()
			}
		}()
	}
	for i := 0; i < common.GoMaxProc; i++ {
		println("done", b.N)
		<-done
	}
}

func BenchmarkIdMid(b *testing.B) {
	w := sync.WaitGroup{}
	w.Add(100)
	b.ResetTimer()
	for i := 0; i < 100; i++ {
		go func() {
			for i := 0; i < b.N; i++ {
				runtime.Getmid()
			}
			b.StopTimer()
			w.Done()
			b.StartTimer()
		}()
	}
	w.Wait()
}

func BenchmarkIdProcpin(b *testing.B) {
	w := sync.WaitGroup{}
	w.Add(100)
	b.ResetTimer()
	for i := 0; i < 100; i++ {
		go func() {
			for i := 0; i < b.N; i++ {
				runtime_procPin()
				b.StopTimer()
				time.Sleep(5000)
				b.StartTimer()
				runtime_procUnpin()
			}
			b.StopTimer()
			w.Done()
			b.StartTimer()
		}()
	}
	w.Wait()
}

type testBuffer struct {
	buf []byte
}

func TestTest(t *testing.T) {
	slice := make([]byte, 0, 10)
	println(len(slice))
	println(cap(slice))
	slice = slice[:4]
	println(len(slice))
	println(cap(slice))
	slice = slice[:0]
	println(len(slice))
	println(cap(slice))
	tb := &testBuffer{}
	println(len(tb.buf))
	println(cap(tb.buf))
}

func BenchmarkBufferPool(b *testing.B) {
}

func formatFileSize(fileSize int64) (size string) {
	if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}

//go:linkname runtime_procPin runtime.procPin
func runtime_procPin() int

//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin()

func TestPin(t *testing.T) {
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	c, err := l.Accept()
	if err != nil {
		panic(err)
	}
	runtime_procPin()
	rs, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		panic(err)
	}
	println(rs)
	runtime_procUnpin()
}

func TestTcpServerForBenchmark(t *testing.T) {
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	for {
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go io.Copy(ioutil.Discard, c)
	}
}

func BenchmarkMemPool4(b *testing.B) {
	b.N *= 1024
	c, err := net.Dial("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		buf := buffer.EightBytesPool.Get().([]byte)
		binary.BigEndian.PutUint32(buf, 10086)
		l, err := c.Write(buf)
		if err != nil {
			panic(err)
		}
		if l != 4 {
			panic("write len != 4")
		}
		buffer.EightBytesPool.Put(buf)
	}
}
func BenchmarkMemMake4(b *testing.B) {
	b.N *= 1024
	c, err := net.Dial("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, 10086)
		l, err := c.Write(buf)
		if err != nil {
			panic(err)
		}
		if l != 4 {
			panic("write len != 4")
		}
	}
}


/**
BenchmarkLockLock-16              250785              4870 ns/op
BenchmarkLockChan-16               40602             28844 ns/op
BenchmarkLockDefer-16             217021              5551 ns/op
 */
func BenchmarkLockLock(b *testing.B) {
	test := 0
	l := sync.Mutex{}
	w := sync.WaitGroup{}
	w.Add(64)
	for i := 0; i < 64; i++ {
		go func() {
			for i := 0; i < b.N; i++ {
				l.Lock()
				test++
				l.Unlock()
			}
			w.Done()
		}()
	}
	w.Wait()
	if test != 64*b.N {
		panic("test = " + strconv.Itoa(test))
	}
}
func BenchmarkLockChan(b *testing.B) {
	test := 0
	c := make(chan bool, 1024)
	for i := 0; i < 64; i++ {
		go func() {
			for i := 0; i < b.N; i++ {
				c <- true
			}
		}()
	}
	for i := 0; i < 64*b.N; i++ {
		<-c
		test++
	}
}
func BenchmarkLockDefer(b *testing.B) {
	test := &struct {
		v int
	}{}
	l := &sync.Mutex{}
	w := &sync.WaitGroup{}
	w.Add(64)
	for i := 0; i < 64; i++ {
		go func() {
			for i := 0; i < b.N; i++ {
				deferLock(l, test)
			}
			w.Done()
		}()
	}
	w.Wait()
	if test.v != 64*b.N {
		panic("test = " + strconv.Itoa(test.v))
	}
}
func deferLock(l *sync.Mutex, test *struct{ v int }) {
	l.Lock()
	defer l.Unlock()
	test.v++
}
