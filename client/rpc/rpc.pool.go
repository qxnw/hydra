package rpc

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/qxnw/lib4go/pool"
)

var counter int32

//RPCClientPool client 缓存池
type RPCClientPool struct {
	p       pool.IPool
	timeout time.Duration
	address string
}

//NewRPCClientPool 构建LUA引擎池
func NewRPCClientPool(address string, maxSize int, timeout time.Duration, opts ...ClientOption) (pl *RPCClientPool, err error) {
	pl = &RPCClientPool{timeout: timeout, address: address}
	pl.p, err = pool.New(&pool.PoolConfigOptions{
		InitialCap:  1,
		MaxCap:      maxSize,
		IdleTimeout: pl.timeout,
		Factory: func() (interface{}, error) {
			pl.printCounter(1)
			return NewRPCClient(address, opts...)
		},
		Close: func(v interface{}) error {
			pl.printCounter(-1)
			client := v.(*RPCClient)
			client.Close()
			return nil
		},
	})
	return
}
func (p *RPCClientPool) printCounter(v int32) {
	atomic.AddInt32(&counter, v)
	if v > 0 {
		fmt.Println("+", atomic.LoadInt32(&counter), p.address)
	} else {
		fmt.Println("-", atomic.LoadInt32(&counter), p.address)
	}
}

//GetClient 获取一个可用的rpc client
func (p *RPCClientPool) GetClient() (*RPCClient, error) {
	c, err := p.p.Get()
	if err != nil {
		return nil, err
	}
	defer p.p.Put(c)
	client := c.(*RPCClient)
	if client.CanUse() {
		return client, nil
	}
	return p.GetClient()
}

//Close 关闭所有连接池
func (p *RPCClientPool) Close() {
	p.p.Release()
}
