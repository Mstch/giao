package giao

import (
	"sync"
)

type ChainMsgHandler func(in Msg, session Session) bool
type PipelineHandler struct {
	Handler
	hs []ChainMsgHandler
}

func NewPipelineHandler(reqPool *sync.Pool) *PipelineHandler {
	ph := &PipelineHandler{}
	ph.InputPool = reqPool
	ph.hs = make([]ChainMsgHandler, 0)
	return ph
}

func (p *PipelineHandler) Append(h ChainMsgHandler) *PipelineHandler {
	p.hs = append(p.hs, h)
	return p
}

func (p *PipelineHandler) Build() *Handler {
	p.H = func(req Msg, session Session) {
		for _, h := range p.hs {
			if !h(req, session) {
				break
			}
		}
	}
	return &p.Handler
}
