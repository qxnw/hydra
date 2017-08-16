package rpc

import (
	"sync"

	"golang.org/x/net/context"

	"github.com/qxnw/hydra/server/rpc/pb"
	"github.com/qxnw/lib4go/types"
)

type process struct {
	srv *RPCServer
	p   *sync.Pool
}

func newProcess(r *RPCServer) *process {
	return &process{
		srv: r,
		p: &sync.Pool{
			New: func() interface{} {
				return &Context{}
			},
		},
	}
}

//Request 客户端处理客户端请求
func (r *process) Request(context context.Context, request *pb.RequestContext) (p *pb.ResponseContext, err error) {
	r.srv.mu.RLock()
	defer r.srv.mu.RUnlock()
	ctx := r.p.Get().(*Context)
	defer func() {
		ctx.context = nil
		ctx.request = nil
		ctx.Writer = nil
		ctx.Close()
		r.p.Put(ctx)
	}()
	ctx.server = r.srv
	ctx.reset("REQUEST", context, request)
	ctx.invoke()
	p = &pb.ResponseContext{}
	p.Status = int32(ctx.Writer.Code)
	p.Result = string(ctx.Writer.Buffer.Bytes())
	p.Params, _ = types.ToStringMap(ctx.Writer.Params)
	ctx.Writer.Reset()
	return
}

//Request 客户端处理客户端请求
func (r *process) Query(context context.Context, request *pb.RequestContext) (p *pb.ResponseContext, err error) {
	r.srv.mu.RLock()
	defer r.srv.mu.RUnlock()
	ctx := &Context{}
	ctx.server = r.srv
	ctx.reset("QUERY", context, request)
	ctx.invoke()
	p = &pb.ResponseContext{}
	p.Status = int32(ctx.Writer.Code)
	p.Result = ctx.Writer.String()
	ctx.Writer.Reset()
	return
}

//Request 客户端处理客户端请求
func (r *process) Update(context context.Context, request *pb.RequestContext) (p *pb.ResponseNoResultContext, err error) {
	r.srv.mu.RLock()
	defer r.srv.mu.RUnlock()
	ctx := &Context{}
	ctx.server = r.srv
	ctx.reset("UPDATE", context, request)
	ctx.invoke()
	p = &pb.ResponseNoResultContext{}
	p.Status = int32(ctx.Writer.Code)
	ctx.Writer.Reset()
	return
}

//Request 客户端处理客户端请求
func (r *process) Delete(context context.Context, request *pb.RequestContext) (p *pb.ResponseNoResultContext, err error) {
	r.srv.mu.RLock()
	defer r.srv.mu.RUnlock()
	ctx := &Context{}
	ctx.server = r.srv
	ctx.reset("DELETE", context, request)
	ctx.invoke()
	p = &pb.ResponseNoResultContext{}
	p.Status = int32(ctx.Writer.Code)
	ctx.Writer.Reset()
	return
}

//Request 客户端处理客户端请求
func (r *process) Insert(context context.Context, request *pb.RequestContext) (p *pb.ResponseNoResultContext, err error) {
	r.srv.mu.RLock()
	defer r.srv.mu.RUnlock()
	ctx := &Context{}
	ctx.server = r.srv
	ctx.reset("INSERT", context, request)
	ctx.invoke()
	p = &pb.ResponseNoResultContext{}
	p.Status = int32(ctx.Writer.Code)
	ctx.Writer.Reset()
	return
}

//Heartbeat 返回心跳数据
func (r *process) Heartbeat(ctx context.Context, in *pb.HBRequest) (*pb.HBResponse, error) {
	return &pb.HBResponse{Pong: in.Ping}, nil
}
