package codec

import "io"

const (
	JsonType = "application/json"
	GobType  = "application/gob"
)

// Header call ("service.method", in, out)
type Header struct {
	ServiceMethod string // "Example.New" implement by Go "reflect"
	Seq           uint64 // request seq number for client
	Error         string
}

// Codec codec interface for extension
type Codec interface {
	io.Closer
	ReadHeader(header *Header) error  // parse header for codec interface
	ReadBody(interface{}) error       // to be codec by specific func
	Write(*Header, interface{}) error // write header and body
}

var NewCodecFuncMap map[string]NewCodecFunc

// NewCodecFunc new-function as a type
type NewCodecFunc func(closer io.ReadWriteCloser) Codec

func init() {
	NewCodecFuncMap = make(map[string]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
	NewCodecFuncMap[JsonType] = NewJsonCodec
}
