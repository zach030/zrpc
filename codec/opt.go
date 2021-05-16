package codec

import (
	"errors"
)

const (
	ZRpcMagicNumber = 0x3bef5c
)

type Option struct {
	MagicNumber int
	CodecType   string
}

var DefaultOpt = &Option{
	MagicNumber: ZRpcMagicNumber,
	CodecType:   GobType,
}

func ParseOptions(opts ...*Option)(*Option,error){
	if len(opts)==0||opts[0]==nil{
		return DefaultOpt,nil
	}
	if len(opts)!=1{
		return nil,errors.New("more than one options")
	}

	opt := opts[0]

	opt.MagicNumber = ZRpcMagicNumber
	if opt.CodecType==""{
		opt.CodecType = DefaultOpt.CodecType
	}
	return opt,nil
}