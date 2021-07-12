package giao

type Server interface {
	Listen(network, address string) error
	Serve() error
	ListenAndServe(network, address string) error
	RegWithId(id int, handler *Handler) Server
	Shutdown() error
	Flush()
}

type Client interface {
	Connect(network, address string) (Client, error)
	Serve() error
	RegWithId(id int, handler *Handler) Client
	Go(id int, req Msg) error
	Shutdown() error
	Flush() error
}

type MultiConnClient interface {
	Connect(network, address string) (MultiConnClient, error)
	RegWithId(id int, handler *Handler) MultiConnClient
	Go(id int, req Msg) error
	Serve() chan error
	Broadcast(id int, req Msg) chan error
	SetSelector(selector Selector)
	Shutdown() chan error
}

type Selector interface {
	AddSession(session Session)
	GetAllSession() []Session
	Select() Session
}

type Session interface {
	GetId() uint64
	Write(handlerId int, msg Msg) error
}

type Pool interface {
	Get() interface{}
	Put(interface{})
}
