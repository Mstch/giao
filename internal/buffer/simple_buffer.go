package buffer

type SimpleBuffer struct {
	buf []byte
}

func NewSimpleBuffer() *SimpleBuffer {
	return &SimpleBuffer{buf: make([]byte, 64)}
}

func (sb *SimpleBuffer) Take(length int) []byte {
	sb.grow(length)
	return sb.buf[:length]
}

func (sb *SimpleBuffer) grow(length int) {
	length = 64 * (length/64 + 1)
	l := len(sb.buf)
	if l < length {
		sb.buf = make([]byte, 2*l+length)
	}
}
