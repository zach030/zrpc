package codec

import "io"

// struct of rpc message header
type Header struct {
	ServiceMethod string // service name "service.method"
	SeqNo         uint64 // request no
	ErrMsg        string // error msg
}

// 实rpc框架中编解码的部分，
// interface of codec
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" // not implemented
)

// all codec map
var NewCodecFuncMap map[Type]NewCodecFunc

// 这里的map存放的value不是实例，而是构造函数
func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
	NewCodecFuncMap[JsonType] = NewJsonCodec
}