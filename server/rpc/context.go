// Copyright 2015 The WebServer Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"reflect"

	hydra "github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server/rpc/pb"
	"github.com/qxnw/lib4go/logger"
	"golang.org/x/net/context"
)

type HandlerFunc func(ctx *Context)

func (h HandlerFunc) Handle(ctx *Context) {
	h(ctx)
}

type Handler interface {
	Handle(*Context)
}
type Writer struct {
	Code int
	*bytes.Buffer
	isWritten bool
}

type Context struct {
	server *RPCServer
	*logger.Logger

	idx      int
	Writer   *Writer
	context  context.Context
	request  *pb.RequestContext
	route    *Route
	params   hydra.Params
	callArgs []reflect.Value
	matched  bool
	method   string
	stage    byte
	action   interface{}
	Result   interface{}
}

func (ctx *Context) reset(method string, context context.Context, request *pb.RequestContext) {
	ctx.method = method
	ctx.context = context
	ctx.request = request
	ctx.Writer = &Writer{Code: 0, Buffer: bytes.NewBufferString("")}
	ctx.idx = 0
	ctx.stage = 0
	ctx.route = nil
	ctx.params = nil
	ctx.callArgs = nil
	ctx.matched = false
	ctx.action = nil
	ctx.Result = nil
	session_id := logger.CreateSession()
	if sid, ok := request.Args["hydra_sid"]; ok {
		session_id = sid
	}
	ctx.Logger = logger.GetSession(request.Service, session_id)
}

func (ctx *Context) HandleError() {
	ctx.server.ErrHandler.Handle(ctx)
}
func (ctx *Context) GetStatusCode() int {
	switch ctx.Result.(type) {
	case *StatusResult:
		return ctx.Result.(*StatusResult).Code
	case AbortError:
		return ctx.Result.(AbortError).Code()
	}
	return 204
}
func (ctx *Context) Req() *pb.RequestContext {
	return ctx.request
}
func (ctx *Context) Service() string {
	return ctx.Req().Service
}
func (ctx *Context) Method() string {
	return ctx.method
}
func (ctx *Context) IP() string {
	return ctx.Req().Args["ip"]
}
func (ctx *Context) Params() *hydra.Params {
	ctx.newAction()
	return &ctx.params
}

func (ctx *Context) Action() interface{} {
	ctx.newAction()
	return ctx.action
}

func (ctx *Context) ActionValue() reflect.Value {
	ctx.newAction()
	return ctx.callArgs[0]
}
func (ctx *Context) Written() bool {
	return ctx.Writer.Len() > 0
}
func (ctx *Context) WriteString(content string) (int, error) {
	return io.WriteString(ctx.Writer, content)
}
func (ctx *Context) Write(p []byte) (n int, err error) {
	return ctx.Writer.Write(p)
}
func (ctx *Context) WriteHeader(code int) {
	ctx.Writer.Code = code
}
func (ctx *Context) newAction() {
	if !ctx.matched {
		reqPath := removeStick(ctx.Req().Service)
		ctx.route, ctx.params = ctx.server.Match(reqPath, ctx.method)
		if ctx.route != nil {
			vc := ctx.route.newAction()
			ctx.action = vc.Interface()
			switch ctx.route.routeType {
			case FuncCtxRoute:
				ctx.callArgs = []reflect.Value{reflect.ValueOf(ctx)}
			default:
				panic("routeType error")
			}
		}
		ctx.matched = true
	}
}

// WARNING: don't invoke this method on action
func (ctx *Context) Next() {
	ctx.idx += 1
	ctx.invoke()
}

func (ctx *Context) execute() {
	ctx.newAction()
	// route is matched
	if ctx.action != nil {
		if len(ctx.route.handlers) > 0 && ctx.stage == 0 {
			ctx.idx = 0
			ctx.stage = 1
			ctx.invoke()
			return
		}

		var ret []reflect.Value
		switch fn := ctx.route.raw.(type) {
		case func(*Context):
			fn(ctx)
		default:
			ret = ctx.route.method.Call(ctx.callArgs)
		}

		if len(ret) == 1 {
			ctx.Result = ret[0].Interface()
		} else if len(ret) == 2 {
			if code, ok := ret[0].Interface().(int); ok {
				ctx.Result = &StatusResult{code, ret[1].Interface(), JsonResponse}
			}
		}
		// not route matched
	} else {
		if !ctx.Written() {
			ctx.NotFound()
		}
	}
}

func (ctx *Context) invoke() {
	if ctx.stage == 0 {
		if ctx.idx < len(ctx.server.handlers) {
			ctx.server.handlers[ctx.idx].Handle(ctx)
		} else {
			ctx.execute()
		}
	} else if ctx.stage == 1 {
		if ctx.idx < len(ctx.route.handlers) {
			ctx.route.handlers[ctx.idx].Handle(ctx)
		} else {
			ctx.execute()
		}
	}
}

func toHTTPError(err error) (msg string, httpStatus int) {
	if os.IsNotExist(err) {
		return "404 page not found", http.StatusNotFound
	}
	if os.IsPermission(err) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}
func (ctx *Context) ServiceTooManyRequests() {
	ctx.Abort(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
}

func (ctx *Context) Unauthorized() {
	ctx.Abort(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
}

// NotFound writes a 404 HTTP response
func (ctx *Context) NotFound(message ...string) {
	if len(message) == 0 {
		ctx.Abort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	ctx.Abort(http.StatusNotFound, message[0])
}

// Abort is a helper method that sends an HTTP header and an optional
// body. It is useful for returning 4xx or 5xx errors.
// Once it has been called, any return value from the handler will
// not be written to the response.
func (ctx *Context) Abort(status int, body ...string) {
	ctx.Result = Abort(status, body...)
	ctx.HandleError()
}

type Contexter interface {
	SetContext(*Context)
}

type Ctx struct {
	*Context
}

func (c *Ctx) SetContext(ctx *Context) {
	c.Context = ctx
}

func Contexts() HandlerFunc {
	return func(ctx *Context) {
		if action := ctx.Action(); action != nil {
			if a, ok := action.(Contexter); ok {
				a.SetContext(ctx)
			}
		}
		ctx.Next()
	}
}
