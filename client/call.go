package client

type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Error error
	Done chan *Call
}

// 调用结束时 调用此方法通知调用方
func (c *Call) done()  {
	c.Done <- c
}
