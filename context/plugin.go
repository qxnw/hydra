package context

import (
	"time"

	"github.com/qxnw/lib4go/rpc"
)

type RPCInvoker interface {
	PreInit(services ...string) (err error)
	RequestFailRetry(service string, input map[string]string, times int) (status int, result string, params map[string]string, err error)
	Request(service string, input map[string]string, failFast bool) (status int, result string, param map[string]string, err error)
	AsyncRequest(service string, input map[string]string, failFast bool) rpc.IRPCResponse
	WaitWithFailFast(callback func(string, int, string, error), timeout time.Duration, rs ...rpc.IRPCResponse) error
}
