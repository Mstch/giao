package main

import (
	"fmt"
	"github.com/Mstch/giao"
	"github.com/Mstch/giao/internal/client"
	"github.com/Mstch/giao/internal/server"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var testMsgs = 10000

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}
func RandStringRunes() string {
	b := make([]rune, 64+rand.Intn(4*1024))
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
func main() {
	startServer()
	startClient()
}

func startClient() {
	checkArr := make([]int32, testMsgs)
	wg := &sync.WaitGroup{}
	wg.Add(testMsgs)
	c, err := client.NewStupidClient().RegWithId(0, &giao.Handler{
		Handle: func(in giao.Msg, session giao.Session) {
			echo, ok := in.(*Echo)
			if !ok {
				panic(in)
			}
			if ok := atomic.CompareAndSwapInt32(&checkArr[echo.Index], 0, 1); !ok {
				panic(echo.Index)
			}
			wg.Done()
		},
		InputPool: &giao.NoCachePool{New: func() interface{} {
			return &Echo{}
		}}}).Connect("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	go func() {
		err := c.Serve()
		if err != nil {
			panic(err)
		}
	}()
	for i := 0; i < testMsgs; i++ {
		var j = i
		err = c.Go(0, &Echo{
			Index:   int32(j),
			Content: []byte(RandStringRunes()),
		})
	}
	if err != nil {
		panic(err)
	}
	wg.Wait()
	for _, i := range checkArr {
		if i != 1 {
			panic(i)
		}
	}
	fmt.Println("test pass")
}

func startServer() {
	checkArr := make([]int32, testMsgs)
	msgPool := &giao.NoCachePool{New: func() interface{} {
		return &Echo{}
	}}
	s, err := server.NewStupidServer().RegWithId(0, &giao.Handler{
		Handle: func(in giao.Msg, session giao.Session) {
			echo, ok := in.(*Echo)
			if !ok {
				panic(in)
			}

			if ok := atomic.CompareAndSwapInt32(&checkArr[echo.Index], 0, 1); !ok {
				panic(echo.Index)
			}
			resp := &Echo{}
			resp.Index = echo.Index
			resp.Content = echo.Content
			err := session.Write(0, resp)
			if err != nil {
				panic(err)
			}
		},
		InputPool: msgPool}).Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	go func() {
		err := s.Serve()
		if err != nil {
			panic(err)
		}
	}()
}
