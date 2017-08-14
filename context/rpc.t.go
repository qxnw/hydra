package context

import (
	"time"

	"github.com/qxnw/lib4go/rpc"
)

type TRPC struct {
	Status int
	Result string
	Params map[string]string
	Error  error
}

type TRPCResponse struct {
	service string
	rpc     *TRPC
}

func (t *TRPCResponse) Wait(timeout time.Duration) (int, string, error) {
	return t.rpc.Status, t.rpc.Result, t.rpc.Error

}
func (t *TRPCResponse) GetResult() chan rpc.IRPCResult {
	c := make(chan rpc.IRPCResult)
	c <- &TRPCResult{response: t}
	return c

}

type TRPCResult struct {
	response *TRPCResponse
}

func (t *TRPCResult) GetService() string {
	return t.response.service
}
func (t *TRPCResult) GetStatus() int {
	return t.response.rpc.Status
}
func (t *TRPCResult) GetResult() string {
	return t.response.rpc.Result
}
func (t *TRPCResult) GetErr() error {
	return t.response.rpc.Error
}

func (r *TRPC) RequestFailRetry(service string, input map[string]string, times int) (status int, result string, params map[string]string, err error) {
	return r.Status, r.Result, r.Params, r.Error
}
func (r *TRPC) Request(service string, input map[string]string, failFast bool) (status int, result string, param map[string]string, err error) {
	return r.Status, r.Result, r.Params, r.Error
}
func (r *TRPC) AsyncRequest(service string, input map[string]string, failFast bool) rpc.IRPCResponse {
	return &TRPCResponse{rpc: r, service: service}
}
func (r *TRPC) WaitWithFailFast(callback func(string, int, string, error), timeout time.Duration, rs ...rpc.IRPCResponse) error {
	return r.Error
}
func (r *TRPC) PreInit(services ...string) error {
	return r.Error
}
