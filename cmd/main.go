package main

import (
	"encoding/json"
	"fmt"
	_ "log"
	"net"
	"time"
	"zrpc/codec"
	"zrpc/logger"
	"zrpc/server"
)

func startServer(addr chan string) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		logger.Error("listen err:%v", err)
	}
	logger.Info("start rpc server on:%v", l.Addr())

	addr <- l.Addr().String()

	server.Accept(l)
}

func main() {
	addr := make(chan string)
	go startServer(addr)

	conn, _ := net.Dial("tcp", <-addr)
	defer conn.Close()

	time.Sleep(time.Second)

	// send opt
	_ = json.NewEncoder(conn).Encode(codec.DefaultOpt)
	cc := codec.NewGobCodec(conn)

	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
			Error:         "",
		}
		_ = cc.Write(h, fmt.Sprintf("zrpc req:%d", h.Seq))

		_ = cc.ReadHeader(h)

		var reply string

		_ = cc.ReadBody(&reply)

		logger.Info("get resp from server:%v", reply)
	}
}
