package context

import (
	"time"

	"github.com/qxnw/lib4go/rpc"
)

type RPCInvoker interface {
	PreInit(services ...string) (err error)
	RequestFailRetry(service string, method string, header map[string]string, form map[string]string, times int) (status int, result string, params map[string]string, err error)
	Request(service string, method string, header map[string]string, form map[string]string, failFast bool) (status int, result string, param map[string]string, err error)
	AsyncRequest(service string, method string, header map[string]string, form map[string]string, failFast bool) rpc.IRPCResponse
	WaitWithFailFast(callback func(string, int, string, error), timeout time.Duration, rs ...rpc.IRPCResponse) error
}
