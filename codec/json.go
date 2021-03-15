package codec

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
)

type JsonCodec struct {
	conn   io.ReadWriteCloser
	buf    *bufio.Writer
	decode *json.Decoder
	encode *json.Encoder
}

func NewJsonCodec(conn io.ReadWriteCloser) Codec {
	return &JsonCodec{
		conn:   conn,
		buf:    bufio.NewWriter(conn),
		decode: json.NewDecoder(conn),
		encode: json.NewEncoder(conn),
	}
}
func (j JsonCodec) Close() error {
	return j.conn.Close()
}

func (j JsonCodec) ReadHeader(header *Header) error {
	return j.decode.Decode(header)
}

func (j JsonCodec) ReadBody(body interface{}) error {
	return j.decode.Decode(body)
}

func (j JsonCodec) Write(header *Header, body interface{}) error {
	defer func() {
		err := j.buf.Flush()
		if err != nil {
			_ = j.Close()
		}
	}()
	if err := j.encode.Encode(header); err != nil {
		log.Println("rpc codec: json error encoding header,", err)
		return err
	}
	if err := j.encode.Encode(body); err != nil {
		log.Println("rpc codec: json error encoding body,", err)
		return err
	}
	return nil
}
