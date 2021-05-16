package codec

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
