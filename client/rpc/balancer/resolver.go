package balancer

import (
	"time"

	"github.com/qxnw/hydra/registry"

	"fmt"

	"google.golang.org/grpc/naming"
)

//Resolver 服务解析器
type Resolver struct {
	timeout    time.Duration
	service    string
	sortPrefix string
}

//NewResolver 返回服务解析器
func NewResolver(service string, timeout time.Duration, sortPrefix string) naming.Resolver {
	return &Resolver{timeout: timeout, service: service, sortPrefix: sortPrefix}
}

// Resolve to resolve the service from zookeeper, target is the dial address of zookeeper
// target example: "zk://192.168.0.159:2181,192.168.0.154:2181"
func (v *Resolver) Resolve(target string) (naming.Watcher, error) {
	r, err := registry.NewRegistryWithAddress(target, nil)
	if err != nil {
		return nil, fmt.Errorf("rpc.client.resolver target err:%v", err)
	}
	return &Watcher{client: r, service: v.service, sortPrefix: v.sortPrefix, closeCh: make(chan struct{})}, nil
}
