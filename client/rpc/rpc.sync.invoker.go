package rpc

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/qxnw/lib4go/rpc"
)

type RPCResponse struct {
	Service string
	Result  chan rpc.IRPCResult
}

type RpcResult struct {
	Service string
	Status  int
	Result  string
	Params  map[string]string
	Err     error
}

func (r *RpcResult) GetService() string {
	return r.Service
}

func (r *RpcResult) GetStatus() int {
	return r.Status
}

func (r *RpcResult) GetResult() string {
	return r.Result
}

func (r *RpcResult) GetErr() error {
	return r.Err
}
func NewRPCResponse(service string) *RPCResponse {
	return &RPCResponse{Service: service, Result: make(chan rpc.IRPCResult, 1)}
}

//Wait 等待请求返回
func (r *RPCResponse) Wait(timeout time.Duration) (int, string, error) {
	select {
	case <-time.After(timeout):
		return 500, "", fmt.Errorf("rpc(%s) 请求等待超时", r.Service)
	case value := <-r.Result:
		return value.GetStatus(), value.GetResult(), value.GetErr()
	}
}
func (r *RPCResponse) GetResult() chan rpc.IRPCResult {
	return r.Result
}

//AsyncRequest 异步请求
func (r *RPCInvoker) AsyncRequest(service string, input map[string]string, failFast bool) rpc.IRPCResponse {
	result := NewRPCResponse(service)
	go func() {
		data := &RpcResult{Service: service}
		data.Status, data.Result, data.Params, data.Err = r.Request(service, input, failFast)
		result.Result <- data
	}()
	return result
}

//WaitWithFailFast 快速失败，当有一个请求返回失败后不再等待其它请求，直接返回失败
func (r *RPCInvoker) WaitWithFailFast(callback func(string, int, string, error), timeout time.Duration, rs ...rpc.IRPCResponse) error {
	errChan := make(chan rpc.IRPCResult, 1)
	allResponse := make(chan struct{}, 1)
	closeCh := make(chan struct{}, 1)
	results := make([]rpc.IRPCResult, 0, len(rs))
	max := int32(len(rs))
	var counter int32
	for _, v := range rs {
		go func(r rpc.IRPCResponse) {
			select {
			case <-closeCh:
				return
			case <-time.After(timeout):
				errChan <- &RpcResult{Err: fmt.Errorf("rpc(%v) 请求等待超时", r)}
				results = append(results, &RpcResult{Status: 500, Err: fmt.Errorf("rpc(%v) 请求等待超时", r)})
			case value := <-r.GetResult():
				if value.GetErr() != nil {
					errChan <- value
				}
				results = append(results, value)
			}
			atomic.AddInt32(&counter, 1)
		}(v)
	}
	go func() {
		select {
		case <-time.After(timeout):
			return
		case <-time.After(time.Millisecond):
			if atomic.LoadInt32(&counter) == max {
				allResponse <- struct{}{}
			}
		}

	}()
	select {
	case <-allResponse:
		for _, v := range results {
			callback(v.GetService(), v.GetStatus(), v.GetResult(), v.GetErr())
		}
		close(closeCh)
	case <-time.After(timeout):
		close(closeCh)
		callback("", 500, "", errors.New("rpc 请求等待超时"))
	case v := <-errChan:
		close(closeCh)
		callback(v.GetService(), v.GetStatus(), v.GetResult(), v.GetErr())
	}
	return nil
}
