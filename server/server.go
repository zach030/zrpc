package server

import (
	"encoding/json"
	"gin-master/gin-master"
	"io"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"
	"zrpc"
	"zrpc/codec"
	"zrpc/logger"
	"zrpc/service"
)

// provide method to call
type Server struct {
	engine     *gin.Engine
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) RegisterService(recv interface{}) error {
	srv := service.NewService(recv)
	if _, exist := s.serviceMap.LoadOrStore(srv.Name, srv); exist {
		return zrpc.ServiceAlreadyExist
	}
	return nil
}

var DefaultServer = NewServer()

func Accept(l net.Listener) {
	DefaultServer.Accept(l)
}

func Register(srv interface{}) error {
	return DefaultServer.RegisterService(srv)
}

func (s *Server) selectService(serviceMethod string) (srv *service.Service, method *service.MethodType, err error) {
	param := strings.Split(serviceMethod, ".")
	if len(param) != 2 {
		err = zrpc.NotMatchRpcArgs
		return
	}
	serviceName, methodName := param[0], param[1]
	srvi, ok := s.serviceMap.Load(serviceName)
	if !ok {
		err = zrpc.NotFoundService
		return
	}
	srv = srvi.(*service.Service)
	method = srv.Method[methodName]
	if method == nil {
		err = zrpc.NotFoundMethod
	}
	return
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
		logger.Info("rpc server detect conn, start serve conn...")
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
	logger.Info("rpc server successfully parse option, start to codec request...")
	s.serveCodec(codecFunc(conn), &opt)
}

// 一个连接存在多个请求(header+body)，需要等到全部请求处理后退出
func (s *Server) serveCodec(cc codec.Codec, opt *codec.Option) {
	sending := new(sync.Mutex) // make sure to send a complete response
	wg := new(sync.WaitGroup)  // wait until all request are handled
	for {
		// todo 1 read-request
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break // it's not possible to recover, so close the connection
			}
			req.Header.Error = err.Error()
			s.sendResponse(cc, req.Header, &InvalidRequest{}, sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(cc, req, sending, wg, opt)
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

	req.Srv, req.MType, err = s.selectService(req.Header.ServiceMethod)
	if err != nil {
		return req, err
	}

	//todo 通过反射得出参数的类型
	req.argv = req.MType.NewArgv()
	req.replyv = req.MType.NewReplyv()

	// make sure that argvi is a pointer, ReadBody need a pointer as parameter
	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}

	err = s.readRequestBody(cc, argvi)
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

func (s *Server) readRequestBody(cc codec.Codec, body interface{}) error {
	if err := cc.ReadBody(body); err != nil {
		logger.Error("read request header failed,err:%v", err)
		return err
	}
	return nil
}

type InvalidRequest struct{}

//todo 在此处实现RPC的函数调用过程
func (s *Server) handleRequest(cc codec.Codec, req *Request, sending *sync.Mutex, wg *sync.WaitGroup, opt *codec.Option) {
	defer wg.Done()

	called := make(chan struct{}, 1)
	sent := make(chan struct{}, 1)

	go func() {
		err := req.Srv.Call(req.MType, req.argv, req.replyv)
		called <- struct{}{}
		if err != nil {
			req.Header.Error = err.Error()
			s.sendResponse(cc, req.Header, InvalidRequest{}, sending)
			sent <- struct{}{}
			return
		}
		s.sendResponse(cc, req.Header, req.replyv.Interface(), sending)
		sent <- struct{}{}
	}()

	if opt.HandleTimeout == 0 {
		called <- struct{}{}
		sent <- struct{}{}
		return
	}

	select {
	case <-time.After(opt.HandleTimeout):
		logger.Error(zrpc.ServerHandleRequestTimeOut.Error())
		s.sendResponse(cc, req.Header, &InvalidRequest{}, sending)
		return
	case <-called:
		<-sent
	}

}

// 需要加锁，不能并发
func (s *Server) sendResponse(cc codec.Codec, header *codec.Header, body interface{}, lock *sync.Mutex) {
	lock.Lock()
	defer lock.Unlock()
	if err := cc.Write(header, body); err != nil {
		logger.Error("write response to client failed,err:%v", err)
	}
}
