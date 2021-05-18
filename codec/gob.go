package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"zrpc/logger"
)

// codec implement by codec func
type GobCodec struct {
	conn   io.ReadWriteCloser
	buf    *bufio.Writer
	decode *gob.Decoder // gob decode API
	encode *gob.Encoder // gob encode API
}

func (g *GobCodec) Close() error {
	return g.conn.Close()
}

// decode header part
func (g *GobCodec) ReadHeader(header *Header) error {
	return g.decode.Decode(header)
}

func (g *GobCodec) ReadBody(i interface{}) error {
	return g.decode.Decode(i)
}

func (g *GobCodec) Write(header *Header, body interface{}) error {
	defer func() {
		err := g.buf.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()
	if err := g.encode.Encode(header); err != nil {
		logger.Error("gob encode header err:%v", err)
		return err
	}
	if err := g.encode.Encode(body); err != nil {
		logger.Error("gob encode body err:%v", err)
		return err
	}
	return nil
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn:   conn,
		buf:    buf,
		decode: gob.NewDecoder(conn),
		encode: gob.NewEncoder(conn),
	}
}
