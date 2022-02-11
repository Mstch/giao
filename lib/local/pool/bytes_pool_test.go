package pool

import "testing"

func TestBytesPool(t *testing.T) {
	if buf := GetBytes(1023); len(buf) != 1023 || cap(buf) != 1024 {
		panic(len(buf))
	}
	if buf := GetBytes(1024); len(buf) != 1024 || cap(buf) != 1024 {
		panic(len(buf))
	}
	buf := GetBytes(1023)
	PutBytes(buf)
	buf1 := GetBytes(1024)
	if len(buf1) != 1024 || cap(buf1) != 1024 {
		panic(len(buf1))
	}
	if &buf1[0] != &buf[0] {
		panic(&buf1[0])
	}
}
