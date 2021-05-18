package service

import (
	"go/ast"
	"reflect"
	"sync/atomic"
	"zrpc/logger"
)

type MethodType struct {
	method    reflect.Method // 方法
	ArgType   reflect.Type   // 入参类型
	ReplyType reflect.Type   // 返回值类型
	numCalls  uint64         // 调用次数
}

func (m *MethodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *MethodType) NewArgv() reflect.Value {
	var argv reflect.Value
	// arg may be a pointer type, or a value type
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *MethodType) NewReplyv() reflect.Value {
	// reply must be a pointer type
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}

type Service struct {
	Name   string                 // service名
	Typ    reflect.Type           // 结构体类型
	Self   reflect.Value          // 结构体自身
	Method map[string]*MethodType // 方法集合
}

func NewService(s interface{}) *Service {
	srv := new(Service)
	// 将service自身放入结构体
	srv.Self = reflect.ValueOf(s)
	// 获取service的name
	srv.Name = reflect.Indirect(srv.Self).Type().Name()
	// 类型
	srv.Typ = reflect.TypeOf(s)
	// 判断是否是对外暴露的
	if !ast.IsExported(srv.Name) {
		logger.Error("rpc server err, not exported service")
	}
	// 根据反射将方法都注册
	srv.registerMethods()

	return srv
}

func (s *Service) registerMethods() {
	s.Method = make(map[string]*MethodType, 0)

	// 得到当前service下的所有方法数，遍历注册到map
	for i := 0; i < s.Typ.NumMethod(); i++ {
		// 获取method
		method := s.Typ.Method(i)
		methodType := method.Type
		if methodType.NumIn() != 3 || methodType.NumOut() != 1 {
			continue
		}
		if methodType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		// 分别获取方法的入参和返回值
		argType, replyType := methodType.In(1), methodType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.Method[method.Name] = &MethodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}

		logger.Info("rpc server register methods, service Name:" + s.Name + " ,Method Name:" + method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

// 用反射完成函数的调用
func (s *Service) Call(m *MethodType, arg, reply reflect.Value) error {
	// 将方法的使用次数自增
	atomic.AddUint64(&m.numCalls, 1)
	// 取出方法函数
	f := m.method.Func
	// 调用函数，入参:service,arg,reply
	returnValues := f.Call([]reflect.Value{s.Self, arg, reply})
	// 返回值错误判断
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
