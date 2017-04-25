package mq

import (
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/mq"
)

type Handler interface {
	Handle(*Context)
}

type HandlerFunc func(ctx *Context)

func (h HandlerFunc) Handle(ctx *Context) {
	h(ctx)
}

type Context struct {
	msg mq.IMessage
	*logger.Logger
	idx        int
	queue      string
	server     *MQConsumer
	params     string
	handle     func(*Context) error
	Result     interface{}
	err        error
	statusCode int
}

func (ctx *Context) reset(q string, msg mq.IMessage, server *MQConsumer, params string, handle func(*Context) error) {
	ctx.idx = 0
	ctx.queue = q
	ctx.msg = msg
	ctx.server = server
	ctx.params = params
	ctx.handle = handle
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
