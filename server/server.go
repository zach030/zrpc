package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"reflect"
	"sync"
	"zrpc/codec"
	"zrpc/logger"
)

// provide method to call
type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func Accept(l net.Listener) {
	DefaultServer.Accept(l)
}

// l 监听句柄
func (s *Server) Accept(l net.Listener) {
	// 阻塞建立连接
	for {
		conn, err := l.Accept()
		if err != nil {
			logger.Error("listener accept connection failed,err:%v", err)
			return
		}
		go s.ServeConn(conn)
	}
}

// TODO 接入ohio/gnet 对于网络服务端进行优化
func (s *Server) ServeConn(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()
	var opt codec.Option
	// 读conn数据
	// 1.opt
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		logger.Error("decode conn option err:%v", err)
		return
	}
	// check opt
	if opt.MagicNumber != codec.ZRpcMagicNumber {
		logger.Error("not match magic number with zrpc")
		return
	}
	// get codec func by parse opt
	codecFunc := codec.NewCodecFuncMap[opt.CodecType]
	if codecFunc == nil {
		logger.Error("not found specific codec type")
		return
	}
	s.ServeCodec(codecFunc(conn))
}

// 一个连接存在多个请求(header+body)，需要等到全部请求处理后退出
func (s *Server) ServeCodec(cc codec.Codec) {
	sendingLock := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.Header.Error = err.Error()
			s.sendResponse(cc, req.Header, "invalid request", sendingLock)
			continue
		}
		wg.Add(1)
		// 处理request是并发的
		go s.handleRequest(cc, req, sendingLock, wg)
	}
	wg.Wait()
	_ = cc.Close()
}

func (s *Server) readRequest(cc codec.Codec) (*Request, error) {
	header, err := s.readRequestHeader(cc)
	if err != nil {
		logger.Error("read request header failed,err :%v", err)
		return nil, err
	}
	req := &Request{Header: header}
	//todo 通过反射得出参数的类型
	req.argv = reflect.New(reflect.TypeOf(""))

	err = s.readRequestBody(cc, req.argv.Interface())
	if err != nil {
		logger.Error("read request body failed,err :%v", err)
		return nil, err
	}
	return req, err
}

func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var header codec.Header
	if err := cc.ReadHeader(&header); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			logger.Error("rpc server read header err:%v", err)
		}
		logger.Error("read request header failed,err:%v", err)
		return nil, err
	}
	return &header, nil
}

func (s Server) readRequestBody(cc codec.Codec, body interface{}) error {
	if err := cc.ReadBody(body); err != nil {
		logger.Error("read request header failed,err:%v", err)
		return err
	}
	return nil
}

//todo 在此处实现RPC的函数调用过程
func (s *Server) handleRequest(cc codec.Codec, req *Request, lock *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Info("request header:%v, arg:%v", req.Header, req.argv.Elem())

	req.replyv = reflect.ValueOf(fmt.Sprintf("zrpc call,id=%d", req.Header.Seq))

	s.sendResponse(cc, req.Header, req.replyv.Interface(), lock)
}

// 需要加锁，不能并发
func (s *Server) sendResponse(cc codec.Codec, header *codec.Header, body interface{}, lock *sync.Mutex) {
	lock.Lock()
	defer lock.Unlock()
	if err := cc.Write(header, body); err != nil {
		logger.Error("write response to client failed,err:%v", err)
	}
}
