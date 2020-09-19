package buffer

import (
	"sync"
)

var EightBytesPool = &sync.Pool{New: func() interface{} {
	return make([]byte, 8)
}}

var CommonBufferPool = &sync.Pool{New: func() interface{} {
	return GetBuffer()
}}

//预热
func init() {

}
