package giao



type NoCachePool struct {
	New func() interface{}
}

func (p *NoCachePool) Get() interface{} {
	return p.New()
}

func (p *NoCachePool) Put(_ interface{}) {
}
