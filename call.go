package zrpc

type Call struct {
	Seq           uint64
	ServiceMethod string      // format "<service>.<method>"
	Args          interface{} // arguments to the function
	Reply         interface{} // reply from the function
	Error         error       // if error occurs, it will be set
	Done          chan *Call  // Strobes when call is complete.
}

//当调用结束时，会调用 call.done() 通知调用方
func (c *Call) done() {
	c.Done <- c
}

