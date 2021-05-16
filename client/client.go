package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"zrpc"
	"zrpc/codec"
	"zrpc/logger"
)

// 客户端：发送请求，接受请求
type Client struct {
	cc          codec.Codec   // 约定的编解码方法
	opt         *codec.Option // 消息头的opt
	sendingLock sync.Mutex    // 保证客户端发送的消息不会混乱
	header      *codec.Header // 请求头

	currSeq     uint64           // 当前注册的call序号
	statusLock  *sync.Mutex      // 对客户端状态更新的锁
	pendingCall map[uint64]*Call // 当前正在进行中的调用
	closed      bool             // 客户端主动关闭
	shutDown    bool             // 有错误发生关闭
}

func NewClient(conn net.Conn, opt *codec.Option) (*Client, error) {
	codecFunc := codec.NewCodecFuncMap[opt.CodecType]
	if codecFunc == nil {
		return nil, fmt.Errorf("not found specific codec func")
	}
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		logger.Error("encode option failed,err:%v", err)
		return nil, err
	}
	client := &Client{
		currSeq:     1,
		cc:          codecFunc(conn),
		opt:         opt,
		pendingCall: make(map[uint64]*Call),
	}
	go client.receive()
	return client, nil
}

func Dial(network, address string, opts ...*codec.Option) (*Client, error) {
	opt, err := codec.ParseOptions(opts...)
	if err != nil {
		logger.Error("parse options failed,err:%v", err)
		return nil, err
	}
	conn, err := net.Dial(network, address)
	if err != nil {
		logger.Error("dial network failed,err:%v", err)
		return nil, err
	}
	return NewClient(conn, opt)
}

// 关闭客户端：如果已关闭则报错，存在错误关闭
func (c *Client) Close() error {
	c.statusLock.Lock()
	defer c.statusLock.Unlock()
	if c.closed {
		return zrpc.ErrShutDown
	}
	c.closed = true
	return c.cc.Close()
}

// 当前客户端是否处于可用状态（未关闭，未错误关闭）
func (c *Client) IsAvailable() bool {
	c.statusLock.Lock()
	defer c.statusLock.Unlock()
	return !c.closed && !c.shutDown
}

// 注册call，加入当前正在处理的call中
func (c *Client) registerCall(call *Call) (uint64, error) {
	c.sendingLock.Lock()
	defer c.sendingLock.Unlock()
	// 对系统当前状态的判断
	if c.closed || c.shutDown {
		return 0, zrpc.ErrShutDown
	}
	call.Seq = c.currSeq
	c.pendingCall[call.Seq] = call
	c.currSeq++
	return call.Seq, nil
}

// 移除call
func (c *Client) removeCall(seq uint64) *Call {
	c.sendingLock.Lock()
	defer c.sendingLock.Unlock()

	if c.closed || c.shutDown {
		return nil
	}
	call := c.pendingCall[seq]
	delete(c.pendingCall, seq)
	return call
}

// 当服务端或客户端发生错误的时候，将正在处理的call进行终止
func (c *Client) terminateCall(err error) {
	c.sendingLock.Lock()
	defer c.sendingLock.Unlock()
	c.statusLock.Lock()
	defer c.statusLock.Unlock()

	c.shutDown = true
	for _, call := range c.pendingCall {
		call.Error = err
		call.done()
	}

}

// 阻塞接收服务端的返回
// 1. 读消息头，失败则跳过这个消息
// 2. 从正在请求的call中取出
// 3. 判断call的状态？为空；有错误；正确
// 4. 正确的call，解析出服务端消息的body部分放到call的reply
func (c *Client) receive() {
	var err error
	for err == nil {
		var header codec.Header
		if err = c.cc.ReadHeader(&header); err != nil {
			break
		}
		call := c.removeCall(header.Seq)

		switch {
		case call == nil:
			err = c.cc.ReadBody(nil)
		case header.Error != "":
			call.Error = fmt.Errorf(header.Error)
			err = c.cc.ReadBody(nil)
			call.done()
		default:
			if err = c.cc.ReadBody(call.Reply); err != nil {
				call.Error = errors.New("client read body failed," + err.Error())
			}
			call.done()
		}
	}
	c.terminateCall(err)
}

// 发送请求到服务端
func (c *Client) send(call *Call) {
	c.sendingLock.Lock()
	defer c.sendingLock.Unlock()

	seq, err := c.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	c.header.Seq = seq
	c.header.ServiceMethod = call.ServiceMethod
	c.header.Error = ""
	if err = c.cc.Write(c.header, call.Args); err != nil {
		// 此次call写数据失败
		// 移除
		call := c.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}
