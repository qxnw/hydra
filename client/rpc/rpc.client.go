package rpc

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/qxnw/hydra/client/rpc/balancer"
	"github.com/qxnw/hydra/server/rpc/pb"
	"github.com/qxnw/lib4go/logger"

	"errors"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

//RPCClient client
type RPCClient struct {
	address string //RPC Server Address 或 registry target
	conn    *grpc.ClientConn
	*clientOption
	client        pb.RPCClient
	longTicker    *time.Ticker
	hasRunChecker bool
	using         int32
	IsConnect     bool
	isClose       bool
}

type clientOption struct {
	connectionTimeout time.Duration
	log               *logger.Logger
	balancer          balancer.CustomerBalancer
	resolver          balancer.ServiceResolver
	service           string
	maxUsing          int32
}

//ClientOption 客户端配置选项
type ClientOption func(*clientOption)

//WithConnectionTimeout 网络连接超时时长
func WithConnectionTimeout(t time.Duration) ClientOption {
	return func(o *clientOption) {
		o.connectionTimeout = t
	}
}

//WithLogger 设置日志记录器
func WithLogger(log *logger.Logger) ClientOption {
	return func(o *clientOption) {
		o.log = log
	}
}

//WithRoundRobinBalancer 设置轮询负载均衡器
func WithRoundRobinBalancer(r balancer.ServiceResolver, service string, timeout time.Duration, limit map[string]int) ClientOption {
	return func(o *clientOption) {
		o.service = service
		o.resolver = r
		o.balancer = balancer.RoundRobin(service, r, limit, o.log)
	}
}

//WithLocalFirstBalancer 设置本地优先负载均衡器
func WithLocalFirstBalancer(r balancer.ServiceResolver, service string, local string, limit map[string]int) ClientOption {
	return func(o *clientOption) {
		o.service = service
		o.resolver = r
		o.balancer = balancer.LocalFirst(service, local, r, limit)
	}
}

//WithBalancer 设置负载均衡器
func WithBalancer(service string, lb balancer.CustomerBalancer) ClientOption {
	return func(o *clientOption) {
		o.service = service
		o.balancer = lb
	}
}

//WithMaxUsing 设置最大使用次数
func WithMaxUsing(max int) ClientOption {
	return func(o *clientOption) {
		o.maxUsing = int32(max)
	}
}

//NewRPCClient 创建客户端
func NewRPCClient(address string, opts ...ClientOption) (*RPCClient, error) {
	client := &RPCClient{address: address, clientOption: &clientOption{connectionTimeout: time.Second * 3}}
	for _, opt := range opts {
		opt(client.clientOption)
	}
	if client.log == nil {
		client.log = logger.GetSession("rpc.client", logger.CreateSession())
	}
	grpclog.SetLogger(client.log)
	err := client.connect()
	if err != nil {
		err = fmt.Errorf("rpc.client连接到服务器失败:%s(err:%v)", address, err)
		return nil, err
	}
	return client, err
}

//Connect 连接服务器，如果当前无法连接系统会定时自动重连
func (c *RPCClient) connect() (err error) {
	if c.IsConnect {
		return nil
	}
	if c.balancer == nil {
		c.conn, err = grpc.Dial(c.address, grpc.WithInsecure(), grpc.WithTimeout(c.connectionTimeout))
	} else {
		ctx, _ := context.WithTimeout(context.Background(), c.connectionTimeout)
		c.conn, err = grpc.DialContext(ctx, c.address, grpc.WithInsecure(), grpc.WithBalancer(c.balancer))
	}
	if err != nil {
		c.IsConnect = false
		return
	}

	c.client = pb.NewRPCClient(c.conn)
	//检查是否已连接到服务器
	/*response, err := c.client.Heartbeat(context.Background(), &pb.HBRequest{Ping: 0})
	c.IsConnect = err == nil && response.Pong == 0
	if err != nil {
		err = fmt.Errorf("发送心跳失败:%v", err)
		return
	}*/
	return nil
}

//Request 发送请求
func (c *RPCClient) Request(service string, input map[string]string, failFast bool) (status int, result string, param map[string]string, err error) {
	atomic.AddInt32(&c.using, 1)
	defer atomic.AddInt32(&c.using, -1)
	response, err := c.client.Request(context.Background(), &pb.RequestContext{Service: service, Args: input},
		grpc.FailFast(failFast))
	if err != nil {
		status = 500
		return
	}
	status = int(response.Status)
	result = response.GetResult()
	param = response.Params
	return
}

//UpdateLimiter 修改限流规则
func (c *RPCClient) UpdateLimiter(limit map[string]int) error {
	if c.balancer != nil {
		c.balancer.UpdateLimiter(limit)
		return nil
	}
	return errors.New("rpc.client.未指定balancer")
}

func (c *RPCClient) canUse() bool {
	atomic.AddInt32(&c.using, 1)
	defer atomic.AddInt32(&c.using, -1)
	return c.maxUsing == 0 || atomic.AddInt32(&c.using, 1) < c.maxUsing

}

//Close 关闭连接
func (c *RPCClient) Close() {
	c.isClose = true
	if c.longTicker != nil {
		c.longTicker.Stop()
	}
	if c.resolver != nil {
		c.resolver.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}

}
