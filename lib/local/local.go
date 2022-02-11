package local

import _ "unsafe"

//go:linkname Pin runtime.procPin
func Pin() int

//go:linkname Unpin runtime.procUnpin
func Unpin()
