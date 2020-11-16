package buffer

import (
	"fmt"
	"io"
	"testing"
)

func TestBuffer_ReadFromLen(t *testing.T) {
	r, w := io.Pipe()
	go func() {
		buf := make([]byte, 4)
		buf[0] = 1
		buf[1] = 2
		buf[2] = 3
		buf[3] = 4
		for i := 0; i < 100; i++ {
			_, err := w.Write(buf)
			if err != nil {
				panic(err)
			}
		}
	}()

	buffer := &Buffer{}
	for i := 0; i < 50; i++ {
		buf, err := buffer.ReadFromLen(r, 8)
		if err != nil {
			panic(err)
		}
		for i := 0; i < 4; i++ {
			if buf[i] != byte(i+1) {
				panic(fmt.Sprint("i:", i, "buf i:", buf[i]))
			}
		}
		for i := 4; i < 8; i++ {
			if buf[i] != byte(i+1-4) {
				panic(fmt.Sprint("i:", i, "buf i:", buf[i]))
			}
		}

	}
}
