package rpclog

import (
	"fmt"

	"github.com/golang/snappy"
	"github.com/qxnw/hydra/rpc"
)

type rpcWriter struct {
	rpcInvoker *rpc.Invoker
	writeError bool
	service    string
}

func newRPCWriter(service string, invoker *rpc.Invoker) (r *rpcWriter) {
	return &rpcWriter{service: service, rpcInvoker: invoker}
}
func (r *rpcWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	p[0] = byte('[')
	p = append(p, byte(']'))
	dst := snappy.Encode(nil, p)
	str := fmt.Sprintf("%s", string(dst))
	_, _, _, err = r.rpcInvoker.Request(r.service, "GET", map[string]string{}, map[string]string{
		"__body": str,
	}, true)
	if err != nil && !r.writeError {
		r.writeError = true
		return len(p) - 1, nil
	}
	if err == nil && r.writeError {
		r.writeError = false
	}
	return len(p) - 1, nil
}
func (r *rpcWriter) Close() error {
	if r.rpcInvoker != nil {
		r.rpcInvoker.Close()
	}
	return nil
}
