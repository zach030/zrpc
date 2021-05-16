package server

import (
	"reflect"
	"zrpc/codec"
)

// RPC请求结构体：header，argv（入参）, replyv（返回值）
type Request struct {
	Header *codec.Header
	argv reflect.Value
	replyv reflect.Value
}

