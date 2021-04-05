package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

// implement of gob-type codec instance
type GobCodec struct {
	conn io.ReadWriteCloser   // connection over rpc
	buf  *bufio.Writer        // buffer for read/write
	decode  *gob.Decoder      // gob-type decode func
	encode  *gob.Encoder      // gob-type encode func
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		decode:  gob.NewDecoder(conn),
		encode:  gob.NewEncoder(buf),
	}
}

func (c *GobCodec) ReadHeader(h *Header) error {
	// decode header to store in h instance
	return c.decode.Decode(h)
}

func (c *GobCodec) ReadBody(body interface{}) error {
	// decode to store in body
	return c.decode.Decode(body)
}

func (c *GobCodec) Write(h *Header, body interface{}) (err error) {
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()
	if err := c.encode.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	if err := c.encode.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}
	return nil
}

func (c *GobCodec) Close() error {
	return c.conn.Close()
}


