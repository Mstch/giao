package giao

import (
	"github.com/gogo/protobuf/proto"
	"sync"
)

type Version int

const (
	StupidRpc Version = iota
)

type PipelineHandler struct {
	Handler
	hs []ProtoHandler
}

func NewPipelineHandler(reqPool *sync.Pool) *PipelineHandler {
	ph := &PipelineHandler{}
	ph.ReqPool = reqPool
	ph.hs = make([]ProtoHandler, 0)
	return ph
}

func (p PipelineHandler) Append(h ProtoHandler) {
	p.hs = append(p.hs, h)
}

func (p PipelineHandler) Build() *Handler {
	p.H = func(req proto.Message, respWriter ProtoWriter) {
		for _, h := range p.hs {
			h(req, respWriter)
		}
	}
	return &p.Handler
}
