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
	Listen(network, address string) error
	RegWithId(id int, handler *Handler) Server
	Shutdown() error
}

type Client interface {
	Connect(network, address string)(Client, error)
	Serve() error
	RegWithId(id int, handler *Handler) Client
	Go(id int, req proto.Message) error
	Shutdown() error
}

type ProtoWriter func(handlerId int, msg proto.Message) error

type PB interface {
	Size() int
	MarshalTo(data []byte) (n int, err error)
	Unmarshal(data []byte) error
}
