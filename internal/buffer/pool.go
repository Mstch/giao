package buffer

import (
	"sync"
)

var CommonBufferPool = &sync.Pool{New: func() interface{} {
	return &Buffer{}
}}

//预热
func init() {

}
