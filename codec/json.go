package codec

import (
	"bufio"
	"encoding/json"
	"io"
	"zrpc/logger"
)

// codec implement by json func
type JsonCodec struct {
	conn   io.ReadWriteCloser
	buf    *bufio.Writer
	decode *json.Decoder
	encode *json.Encoder
}

func (c JsonCodec) Close() error {
	return c.conn.Close()
}

func (c JsonCodec) ReadHeader(header *Header) error {
	return c.decode.Decode(header)
}

func (c JsonCodec) ReadBody(body interface{}) error {
	return c.decode.Decode(body)
}

func (c JsonCodec) Write(header *Header, i interface{}) error {
	defer func() {
		err := c.buf.Flush()
		if err != nil {
			logger.Error("flush buf failed,err:%v", err)
			return
		}
	}()

	if err := c.encode.Encode(header); err != nil {
		logger.Error("json encode header failed,err:%v",err)
		return err
	}
	if err := c.encode.Encode(i); err != nil {
		logger.Error("json encode body failed,err:%v",err)
		return err
	}

	return nil
}

func NewJsonCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return JsonCodec{
		conn:   conn,
		buf:    buf,
		decode: json.NewDecoder(conn),
		encode: json.NewEncoder(conn),
	}
}
