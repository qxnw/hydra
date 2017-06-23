package rpc

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/qxnw/hydra/client/rpc/balancer"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/logger"
)

//RPCInvoker rpc client factory
type RPCInvoker struct {
	cache   cmap.ConcurrentMap
	address string
	opts    []ClientOption
	domain  string
	server  string
	lb      balancer.CustomerBalancer
	*invokerOption
}

type invokerOption struct {
	logger       *logger.Logger
	timerout     time.Duration
	balancerType int
	servers      string
	localPrefix  string
}

const (
	RoundRobin = iota + 1
	LocalFirst
)

//InvokerOption 客户端配置选项
type InvokerOption func(*invokerOption)

//WithInvokerLogger 设置日志记录器
func WithInvokerLogger(log *logger.Logger) InvokerOption {
	return func(o *invokerOption) {
		o.logger = log
	}
}

//WithRoundRobin 设置为轮询负载
func WithRoundRobin() InvokerOption {
	return func(o *invokerOption) {
		o.balancerType = RoundRobin
	}
}

//WithLocalFirst 设置为轮询负载
func WithLocalFirst(prefix string) InvokerOption {
	return func(o *invokerOption) {
		o.balancerType = LocalFirst
		o.localPrefix = prefix
	}
}

//NewRPCInvoker new rpc client factory
//domain: 当前服务所在域
//server: 当前服务器名称
//addrss: 注册中心地址格式: zk://192.168.0.1166:2181或standalone://localhost
func NewRPCInvoker(domain string, server string, address string, opts ...InvokerOption) (f *RPCInvoker) {
	f = &RPCInvoker{
		domain:        domain,
		server:        server,
		address:       address,
		cache:         cmap.New(4),
		invokerOption: &invokerOption{balancerType: RoundRobin},
	}
	for _, opt := range opts {
		opt(f.invokerOption)
	}
	if f.invokerOption.logger == nil {
		f.invokerOption.logger = logger.GetSession("rpc.client", logger.CreateSession())
	}
	return
}
func (r *RPCInvoker) prepareClient(service string) (*RPCClient, error) {
	p, err := r.GetClientPool(service)
	if err != nil {
		return nil, err
	}

	client, err := p.GetClient()
	if err != nil {
		return nil, err
	}
	return client, nil
}

//Request 使用RPC调用Request函数
func (r *RPCInvoker) Request(service string, input map[string]string, failFast bool) (status int, result string, params map[string]string, err error) {
	status = 500
	client, err := r.prepareClient(service)
	if err != nil {
		return
	}
	rservice, _, _, _ := r.resolvePath(service)
	status, result, params, err = client.Request(rservice, input, failFast)
	if status != 200 {
		if err != nil {
			err = fmt.Errorf("%s err:%v", result, err)
		} else {
			err = errors.New(result)
		}
	}
	return
}

//GetClientPool 获取rpc client
//addr 支持格式:
//order.request#merchant.hydra,order.request,order.request@api.hydra,order.request@api
func (r *RPCInvoker) GetClientPool(addr string) (c *RPCClientPool, err error) {
	service, domain, server, err := r.resolvePath(addr)
	if err != nil {
		return
	}
	fullService := fmt.Sprintf("%s/services/%s%s/providers", domain, server, service)
	_, client, err := r.cache.SetIfAbsentCb(fullService, func(i ...interface{}) (interface{}, error) {
		rsrvs := i[0].(string)
		opts := make([]ClientOption, 0, 0)
		opts = append(opts, WithLogger(r.logger))
		rs := balancer.NewResolver(rsrvs, time.Second, r.localPrefix)
		opts = append(opts, WithMaxUsing(100000))
		switch r.balancerType {
		case RoundRobin:
			opts = append(opts, WithRoundRobinBalancer(rs, rsrvs, time.Second, map[string]int{}))
		case LocalFirst:
			opts = append(opts, WithLocalFirstBalancer(rs, rsrvs, r.localPrefix, map[string]int{}))
		default:
		}
		return NewRPCClientPool(r.address, 100, time.Second*300, opts...)
	}, fullService)
	if err != nil {
		return
	}
	c = client.(*RPCClientPool)
	return
}

//PreInit 预初始化客户端
func (r *RPCInvoker) PreInit(services ...string) (err error) {
	for _, v := range services {
		_, err = r.GetClientPool(v)
		if err != nil {
			return
		}
	}
	return
}

//Close 关闭当前服务
func (r *RPCInvoker) Close() {
	r.cache.RemoveIterCb(func(k string, v interface{}) bool {
		client := v.(*RPCClientPool)
		client.Close()
		return true
	})
}

//resolvePath   解析服务地址:
//domain:hydra,server:merchant_cron
//order.request#merchant_api.hydra 解析为:service: /order/request,server:merchant_api,domain:hydra
//order.request 解析为 service: /order/request,server:merchant_cron,domain:hydra
//order.request#merchant_rpc 解析为 service: /order/request,server:merchant_rpc,domain:hydra
func (r *RPCInvoker) resolvePath(address string) (service string, domain string, server string, err error) {
	raddress := strings.TrimRight(address, "#")
	addrs := strings.SplitN(raddress, "#", 2)
	if len(addrs) == 1 {
		if addrs[0] == "" {
			return "", "", "", fmt.Errorf("服务地址%s不能为空", address)
		}
		service = "/" + strings.Trim(strings.Replace(raddress, ".", "/", -1), "/")
		domain = r.domain
		server = r.server
		return
	}
	if len(addrs[0]) == 0 {
		return "", "", "", fmt.Errorf("%s错误，服务名不能为空", address)
	}
	if len(addrs[1]) == 0 {
		return "", "", "", fmt.Errorf("%s错误，服务名，域不能为空", address)
	}
	service = "/" + strings.Trim(strings.Replace(addrs[0], ".", "/", -1), "/")
	raddr := strings.SplitN(strings.TrimRight(addrs[1], "."), ".", 2)
	if len(raddr) == 2 && raddr[0] != "" && raddr[1] != "" {
		domain = raddr[1]
		server = raddr[0]
		return
	}
	if len(raddr) == 1 {
		if raddr[0] == "" {
			return "", "", "", fmt.Errorf("%s错误，服务器名称不能为空", address)
		}
		domain = r.domain
		server = raddr[0]
		return
	}
	if raddr[0] == "" && raddr[1] == "" {
		return "", "", "", fmt.Errorf(`%s错误,未指定服务器名称和域名称`, addrs[1])
	}
	domain = raddr[1]
	server = r.server
	return
}
