package flushbuffer

import (
	"fmt"
	"testing"
)

func TestAccessor(t *testing.T) {
	a := &accessor{}
	i := a.access()
	fmt.Println(a)
	a.release(i)
	fmt.Println(a)
	b := a.tryAccess(1)
	fmt.Println(a, b)
	a.release(1)
	fmt.Println(a)
}
