package buffer

type Buffer struct {
	using int
	buf   []byte
}

func GetBuffer() *Buffer {
	return &Buffer{}
}
func (b *Buffer) Take(size int) []byte {
	if b.using == 1 {
		panic("concurrent use buffer")
	}
	b.using = 1
	l := len(b.buf)
	c := cap(b.buf)
	if l+size > c {
		b.grow(size)
		l = 0
	}
	takeBuf := b.buf[l : l+size]
	b.buf = b.buf[:l+size]
	b.using = 0
	return takeBuf
}

func (b *Buffer) grow(need int) {
	if len(b.buf) >= need {
		b.buf = b.buf[:0]
	} else {
		newb := make([]byte, ceiling64(need+2*cap(b.buf)))
		b.buf = newb
	}
}

func ceiling64(size int) int {
	return size + (size - size%64)
}
