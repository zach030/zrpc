package zrpc

import "errors"

var (
	ErrShutDown = errors.New("connection is shut down")
)
