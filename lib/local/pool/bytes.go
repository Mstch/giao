package pool

import (
	"github.com/Mstch/giao/lib/local"
	"math/bits"
	"runtime"
	_ "unsafe"
)

var (
	bytesPool []localPool
)

type (
	freeQueue []byte
	sizedPool []freeQueue
	localPool []sizedPool
)

func init() {
	bytesPool = make([]localPool, runtime.GOMAXPROCS(0))
	for i := range bytesPool {
		bytesPool[i] = make([]sizedPool, 26)
		for pi := range bytesPool[i] {
			bytesPool[i][pi] = make([]freeQueue, 0, 1024)
		}
	}
}
func GetBytes(need int) []byte {
	var localI int
	if need <= 64 {
		localI = 0
	} else {
		localI = bits.Len64(uint64(need)) - 6
		if need&(need-1) == 0 {
			localI--
		}
	}
	blp := bytesPool[local.Pin()]
	if len((blp)[localI]) == 0 {
		(blp)[localI] = append((blp)[localI], make([]byte, 64<<localI))
	}
	buf := (blp)[localI][len((blp)[localI])-1]
	(blp)[localI] = (blp)[localI][:len((blp)[localI])-1]
	local.Unpin()
	return buf[:need]
}

func PutBytes(buf []byte) {
	if cap(buf) < 64 {
		return
	}

	localI := bits.Len64(uint64(cap(buf))) - 7
	blp := bytesPool[local.Pin()]
	(blp)[localI] = append((blp)[localI], buf[:cap(buf)])
	local.Unpin()
}
