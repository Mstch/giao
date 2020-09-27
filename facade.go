package giao

type Handler struct {
	H         MsgHandler
	InputPool Pool
}

type Server interface {
	Listen(network, address string) error
	Serve() error
	ListenAndServe(network, address string) error
	RegWithId(id int, handler *Handler) Server
	Shutdown() error
}

type Client interface {
	Connect(network, address string) (Client, error)
	Serve() error
	RegWithId(id int, handler *Handler) Client
	Go(id int, req Msg) error
	Shutdown() error
}

type 	MsgHandler func(in Msg, session Session)
type Msg interface {
	Size() int
	MarshalTo(data []byte) (n int, err error)
	Unmarshal(data []byte) error
}

type Session interface {
	Get(key interface{}) (interface{}, bool)
	Set(key, value interface{})
	GetId() uint64
	Write(handlerId int, msg Msg) error
}

type Pool interface {
	Get() interface{}
	Put(interface{})
}
