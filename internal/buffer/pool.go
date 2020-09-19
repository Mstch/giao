package buffer

import (
	"sync"
)



var CommonBufferPool = &sync.Pool{New: func() interface{} {
	return GetBuffer()
}}

//预热
func init() {

}
