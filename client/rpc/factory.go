package rpc

import (
	"time"

	"github.com/qxnw/hydra/client/rpc/balancer"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"google.golang.org/grpc/naming"
)

//RPCClientFactory rpc client factory
type RPCClientFactory struct {
	cache   cmap.ConcurrentMap
	address string
	opts    []ClientOption

	lb balancer.CustomerBalancer
	*facotryOption
}

type facotryOption struct {
	logger       Logger
	timerout     time.Duration
	resolverType int
	balancerType int
	servers      string
	local        string
}

const (
	ZKResolver = iota + 1
	FileResolver
)
const (
	RoundRobin = iota + 1
	LocalFirst
)

//FactoryOption 客户端配置选项
type FactoryOption func(*facotryOption)

//WithFactoryLogger 设置日志记录器
func WithFactoryLogger(log Logger) FactoryOption {
	return func(o *facotryOption) {
		o.logger = log
	}
}

//WithRoundRobin 设置为轮询负载
func WithRoundRobin() FactoryOption {
	return func(o *facotryOption) {
		o.balancerType = RoundRobin
	}
}

//WithLocalFirst 设置为轮询负载
func WithLocalFirst(local string) FactoryOption {
	return func(o *facotryOption) {
		o.balancerType = LocalFirst
		o.local = local
	}
}

//WithZKBalancer 设置基于Zookeeper服务发现的负载均衡器
func WithZKBalancer(servers string, timeout time.Duration) FactoryOption {
	return func(o *facotryOption) {
		o.servers = servers
		o.timerout = timeout
		o.resolverType = ZKResolver
	}
}

//WithFileBalancer 设置本地优先负载均衡器
func WithFileBalancer(f string) FactoryOption {
	return func(o *facotryOption) {
		o.servers = f
		o.resolverType = FileResolver
	}
}

//NewRPCClientFactory new rpc client factory
func NewRPCClientFactory(address string, opts ...FactoryOption) (f *RPCClientFactory) {
	f = &RPCClientFactory{
		address:       address,
		cache:         cmap.New(),
		facotryOption: &facotryOption{},
	}
	for _, opt := range opts {
		opt(f.facotryOption)
	}
	return
}

//PreInitClient 预初始化客户端
func (r *RPCClientFactory) PreInitClient(services ...string) (err error) {
	for _, v := range services {
		_, err = r.Get(v)
		if err != nil {
			return
		}
	}
	return
}

//Get 获取rpc client
func (r *RPCClientFactory) Get(service string) (c *RPCClient, err error) {
	_, client, err := r.cache.SetIfAbsentCb(service, func(i ...interface{}) (interface{}, error) {
		opts := make([]ClientOption, 0, 0)
		opts = append(opts, WithLogger(r.logger))
		if r.resolverType > 0 {
			var rs naming.Resolver
			switch r.resolverType {
			case ZKResolver:
				rs = balancer.NewZKResolver(service, r.local, time.Second)
			}
			if rs != nil {
				switch r.balancerType {
				case RoundRobin:
					opts = append(opts, WithRoundRobinBalancer(rs, service, time.Second, map[string]int{}))
				case LocalFirst:
					opts = append(opts, WithLocalFirstBalancer(rs, service, r.local, map[string]int{}))
				}
			}
		}
		return NewRPCClient(r.address, opts...), nil
	})
	if err != nil {
		return
	}
	c = client.(*RPCClient)
	return
}
