package rpc

import (
	"fmt"
	"io"
	"time"

	"github.com/lunny/log"
	"github.com/qxnw/hydra/client/rpc/balancer"
	"github.com/qxnw/hydra/server/rpc/pb"

	"os"

	"strings"

	"errors"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/naming"
)

//Logger 日志组件
type Logger interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
}

//Client client
type RPCClient struct {
	address string
	conn    *grpc.ClientConn
	*clientOption
	client        pb.RPCClient
	longTicker    *time.Ticker
	hasRunChecker bool
	IsConnect     bool
	isClose       bool
}

type clientOption struct {
	connectionTimeout time.Duration
	log               Logger
	balancer          balancer.CustomerBalancer
	service           string
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
func WithLogger(log Logger) ClientOption {
	return func(o *clientOption) {
		o.log = log
	}
}

//WithRoundRobinBalancer 设置轮询负载均衡器
func WithRoundRobinBalancer(r naming.Resolver, service string, timeout time.Duration, limit map[string]int) ClientOption {
	return func(o *clientOption) {
		o.service = service
		o.balancer = balancer.RoundRobin(service, r, limit)
	}
}

//WithLocalFirstBalancer 设置本地优先负载均衡器
func WithLocalFirstBalancer(r naming.Resolver, service string, local string, limit map[string]int) ClientOption {
	return func(o *clientOption) {
		o.service = service
		o.balancer = balancer.FickFirst(service, local, r, limit)
	}
}

//WithBalancer 设置负载均衡器
func WithBalancer(service string, lb balancer.CustomerBalancer) ClientOption {
	return func(o *clientOption) {
		o.service = service
		o.balancer = lb
	}
}

//NewClient 创建客户端
func NewRPCClient(address string, opts ...ClientOption) *RPCClient {
	client := &RPCClient{address: address, clientOption: &clientOption{connectionTimeout: time.Second * 3}}
	for _, opt := range opts {
		opt(client.clientOption)
	}
	if client.log == nil {
		client.log = NewLogger(os.Stdout)
	}
	grpclog.SetLogger(client.log)
	client.connect()
	return client
}

//Connect 连接服务器，如果当前无法连接系统会定时自动重连
func (c *RPCClient) connect() (b bool) {
	if c.IsConnect {
		return
	}
	var err error
	if c.balancer == nil {
		c.conn, err = grpc.Dial(c.address, grpc.WithInsecure(), grpc.WithTimeout(c.connectionTimeout))
	} else {
		ctx, _ := context.WithTimeout(context.Background(), c.connectionTimeout)
		c.conn, err = grpc.DialContext(ctx, c.address, grpc.WithInsecure(), grpc.WithBalancer(c.balancer))
	}
	if err != nil {
		c.IsConnect = false
		return c.IsConnect
	}
	c.client = pb.NewRPCClient(c.conn)
	//检查是否已连接到服务器
	response, er := c.client.Heartbeat(context.Background(), &pb.HBRequest{Ping: 0})
	c.IsConnect = er == nil && response.Pong == 0
	return c.IsConnect
}

//Request 发送请求
func (c *RPCClient) Request(service string, input map[string]string, failFast bool) (status int, result string, err error) {
	if !strings.HasPrefix(service, c.service) {
		return 500, "", fmt.Errorf("服务:%s调用失败", service)
	}
	//metadata.NewContext(context.Background(), metadata.Pairs(kvs...))
	response, err := c.client.Request(context.Background(), &pb.RequestContext{Service: service, Args: input},
		grpc.FailFast(failFast))
	if err != nil {
		return
	}
	status = int(response.Status)
	result = response.GetResult()
	return
}

//Query 发送请求
func (c *RPCClient) Query(service string, input map[string]string, failFast bool) (status int, result string, err error) {
	if !strings.HasPrefix(service, c.service) {
		return 500, "", fmt.Errorf("服务:%s调用失败", service)
	}
	response, err := c.client.Query(context.Background(), &pb.RequestContext{Service: service, Args: input},
		grpc.FailFast(failFast))
	if err != nil {
		return
	}
	status = int(response.Status)
	result = response.GetResult()
	return
}

//Update 发送请求
func (c *RPCClient) Update(service string, input map[string]string, failFast bool) (status int, err error) {
	if !strings.HasPrefix(service, c.service) {
		return 500, fmt.Errorf("服务:%s调用失败", service)
	}
	response, err := c.client.Update(context.Background(), &pb.RequestContext{Service: service, Args: input},
		grpc.FailFast(failFast))
	if err != nil {
		return
	}
	status = int(response.Status)
	return
}

//Insert 发送请求
func (c *RPCClient) Insert(service string, input map[string]string, failFast bool) (status int, err error) {
	if !strings.HasPrefix(service, c.service) {
		return 500, fmt.Errorf("服务:%s调用失败", service)
	}
	response, err := c.client.Insert(context.Background(), &pb.RequestContext{Service: service, Args: input},
		grpc.FailFast(failFast))
	if err != nil {
		return
	}
	status = int(response.Status)
	return
}

//Delete 发送请求
func (c *RPCClient) Delete(service string, input map[string]string, failFast bool) (status int, err error) {
	if !strings.HasPrefix(service, c.service) {
		return 500, fmt.Errorf("服务:%s调用失败", service)
	}
	response, err := c.client.Delete(context.Background(), &pb.RequestContext{Service: service, Args: input},
		grpc.FailFast(failFast))
	if err != nil {
		return
	}
	status = int(response.Status)
	return

}

//UpdateLimiter 修改限流规则
func (c *RPCClient) UpdateLimiter(limit map[string]int) error {
	if c.balancer != nil {
		c.balancer.UpdateLimiter(limit)
		return nil
	}
	return errors.New("未指定balancer")
}

//UpdateAssign 修改指定条件规则
func (c *RPCClient) UpdateAssign(limit map[string][]string) {

}

//logInfof 日志记录
func (c *RPCClient) logInfof(format string, msg ...interface{}) {
	if c.log == nil {
		return
	}
	c.log.Printf(format, msg...)
}

//Close 关闭连接
func (c *RPCClient) Close() {
	c.isClose = true
	if c.longTicker != nil {
		c.longTicker.Stop()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

//NewLogger 创建日志组件
func NewLogger(out io.Writer) Logger {
	l := log.New(out, "[grpc client] ", log.Ldefault())
	l.SetOutputLevel(log.Ldebug)
	return &nLogger{Logger: l}
}

type nLogger struct {
	*log.Logger
}

func (n *nLogger) Fatalln(args ...interface{}) {
	n.Fatal(args...)
}
