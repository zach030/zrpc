package main

import (
	"context"
	"fmt"
	"log"
	_ "log"
	"net"
	"sync"
	"time"
	"zrpc/client"
	"zrpc/logger"
	"zrpc/server"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(addr chan string) {
	var foo Foo
	if err := server.Register(&foo); err != nil {
		return
	}
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		logger.Error("listen err:%v", err)
	}
	logger.Info("start rpc server on:%v", l.Addr())

	addr <- l.Addr().String()

	server.Accept(l)

	logger.Info("rpc server is running...")
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)

	c, _ := client.Dial("tcp", <-addr)
	defer func() { _ = c.Close() }()

	time.Sleep(time.Second)

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{
				Num1: i,
				Num2: i + 1,
			}
			//todo 注意reply的类型 需要与method保持一致
			var reply *int
			if err := c.SyncCall(ctx, "Foo.Sum", args, &reply); err != nil {
				logger.Error("call Foo.Sum error:" + err.Error())
			} else {
				logger.Info(fmt.Sprintf("%d + %d = %d", args.Num1, args.Num2, *reply))
			}
		}(i)
	}
	wg.Wait()
}
