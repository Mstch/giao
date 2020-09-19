package common

import "runtime"

var GoMaxProc = 2 * runtime.NumCPU()

func init() {
	runtime.GOMAXPROCS(2 * runtime.NumCPU())
}
