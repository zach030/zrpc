package zrpc

import "errors"

var (
	ErrShutDown = errors.New("connection is shut down")
	ServiceAlreadyExist = errors.New("service is already exist")
	NotMatchRpcArgs = errors.New("incorrect args num")
	NotFoundService = errors.New("not found service in rpc server")
	NotFoundMethod = errors.New("not found method in service")

	RpcClientConnectTimeOut = errors.New("rpc client connect server timeout")

	RpcClientCallServiceTimeOut = errors.New("rpc client call service.method failed")
)
