package pool

import (
	"runtime"
	"sync"
	"testing"
)

/**
To make sure the bufs alloc on heap that cannot automatic recycling
@See https://zh.wikipedia.org/wiki/%E9%80%83%E9%80%B8%E5%88%86%E6%9E%90
*/
func Escape(_ []byte) {

}

//concurrency : To make sure goroutines can be schedule by all p;
var concurrency = runtime.GOMAXPROCS(0) * 512

func BenchmarkPLBytesPool(b *testing.B) {
	b.ReportAllocs()
	b.N = concurrency * (b.N/concurrency + 1)
	wg := &sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			for i := 0; i < b.N/concurrency; i++ {
				b := GetBytes(4096)
				Escape(b)
				PutBytes(b)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkStdBytesPool(b *testing.B) {
	b.ReportAllocs()
	b.N = concurrency * (b.N/concurrency + 1)
	wg := &sync.WaitGroup{}
	wg.Add(concurrency)
	bufpool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, 4096)
		},
	}
	for i := 0; i < concurrency; i++ {
		go func() {
			for i := 0; i < b.N/concurrency; i++ {
				b := bufpool.Get().([]byte)
				Escape(b)
				bufpool.Put(b)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}


