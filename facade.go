package giao

import (
	"github.com/gogo/protobuf/proto"
	"sync"
)

type Handler struct {
	H       func(req proto.Message, respWriter ProtoWriter)
	ReqPool *sync.Pool
}
type Server interface {
	RegStruct(handlerStruct interface{})
	RegStructWithId(id int, handlerStruct interface{})
	RegFuncWithId(id int, handler *Handler)
	StartServe() error
	Shutdown() error
}

type Client interface {
	RegFuncWithId(id int, handler *Handler)
	StartServe() error
	ACall(id int, req proto.Message) error
	Call(id int, req proto.Message) (resp proto.Message)
	Shutdown() error
}

type ProtoWriter func(handlerId int, msg proto.Message) error

type PB interface {
	Size() int
	MarshalTo(data []byte) (n int, err error)
	Unmarshal(data []byte) error
}
