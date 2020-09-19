package errors

import "errors"

var ErrWriteToClosedConn = errors.New("write to closed conn")
