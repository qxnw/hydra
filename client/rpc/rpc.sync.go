package rpc

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/qxnw/lib4go/rpc"
)

//AsyncRequest 发起异步Request请求
func (r *Invoker) AsyncRequest(service string, method string, header map[string]string, form map[string]string, failFast bool) rpc.IRPCResponse {
	result := NewResponse(service)
	go func() {
		data := &Result{Service: service}
		data.Status, data.Result, data.Params, data.Err = r.Request(service, method, header, form, failFast)
		result.Result <- data
	}()
	return result
}

//WaitWithFailFast 快速失败，当有一个请求返回失败后不再等待其它请求，直接返回失败
func (r *Invoker) WaitWithFailFast(callback func(string, int, string, error), timeout time.Duration, rs ...rpc.IRPCResponse) error {
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
				errChan <- &Result{Err: fmt.Errorf("rpc(%v) 请求等待超时", r)}
				results = append(results, &Result{Status: 500, Err: fmt.Errorf("rpc(%v) 请求等待超时", r)})
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
