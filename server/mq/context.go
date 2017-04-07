package mq

type Handler interface {
	Handle(*Context)
}

type HandlerFunc func(ctx *Context)

func (h HandlerFunc) Handle(ctx *Context) {
	h(ctx)
}

type Context struct {
	taskName   string
	idx        int
	server     *MQConsumer
	params     interface{}
	handle     func(*Context) error
	Result     interface{}
	err        error
	statusCode int
}

func (ctx *Context) Next() {
	ctx.idx += 1
	ctx.invoke()
}
func (ctx *Context) invoke() {
	if ctx.idx < len(ctx.server.handlers) {
		ctx.server.handlers[ctx.idx].Handle(ctx)
	} else {
		ctx.err = ctx.handle(ctx)
	}
}
