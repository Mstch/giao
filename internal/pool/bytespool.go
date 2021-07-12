package pool

import (
	"math/bits"
	"runtime"
	_ "unsafe"
)

var bytesPool [][][][]byte

func init() {
	bytesPool = make([][][][]byte, runtime.GOMAXPROCS(0))
	for i := range bytesPool {
		bytesPool[i] = make([][][]byte, 26)
		for pi := range bytesPool[i] {
			bytesPool[i][pi] = make([][]byte, 0, 1024)
		}
	}
}
func GetBytes(need int) []byte {
	var localI int
	if need <= 64 {
		localI = 0
	} else {
		localI = bits.Len64(uint64(need)) - 6
	}
	blp := bytesPool[pin()]
	if len((blp)[localI]) == 0 {
		(blp)[localI] = append((blp)[localI], make([]byte, 64<<localI))
	}

	buf := (blp)[localI][len((blp)[localI])-1]
	(blp)[localI] = (blp)[localI][:len((blp)[localI])-1]
	unpin()
	return buf[:need]
}

func PutBytes(buf []byte) {
	if cap(buf) < 64 {
		return
	}
	localI := bits.Len64(uint64(cap(buf))) - 7
	blp := bytesPool[pin()]
	(blp)[localI] = append((blp)[localI], buf[:cap(buf)])
	unpin()
}

//go:linkname pin runtime.procPin
func pin() int

//go:linkname unpin runtime.procUnpin
func unpin() int
