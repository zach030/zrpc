package codec

import (
	"errors"
	"time"
)

const (
	ZRpcMagicNumber = 0x3bef5c
)

type Option struct {
	MagicNumber    int
	CodecType      string
	ConnectTimeout time.Duration // 建立连接超时
	HandleTimeout  time.Duration // 处理连接请求超时
}

var DefaultOpt = &Option{
	MagicNumber:    ZRpcMagicNumber,
	CodecType:      GobType,
	ConnectTimeout: time.Second * 10,
}

func ParseOptions(opts ...*Option) (*Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOpt, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("more than one options")
	}

	opt := opts[0]

	opt.MagicNumber = ZRpcMagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOpt.CodecType
	}
	return opt, nil
}
