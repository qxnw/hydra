package log

import (
	"fmt"
	"sync"

	"github.com/golang/snappy"
	"github.com/qxnw/hydra/rpc"
	"github.com/qxnw/lib4go/logger"
)

//RemoteLogWriter 日志writer
type RemoteLogWriter struct {
	rpcInvoker *rpc.Invoker
	logger     *logger.Logger
	service    string
	writeError bool
	closeChan  chan struct{}
	once       sync.Once
}

//NewRemoteLogWriter 创建rpc日志组件
func NewRemoteLogWriter(rpcInvoker *rpc.Invoker, logger *logger.Logger) (r *RemoteLogWriter, err error) {
	r = &RemoteLogWriter{
		rpcInvoker: rpcInvoker,
		closeChan:  make(chan struct{}),
		logger:     logger,
	}

	return r, nil
}

//Write 写入日志
func (r *RemoteLogWriter) Write(p []byte) (n int, err error) {
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
		r.logger.Errorf("写入RPC日志失败:%v", err)
		return len(p) - 1, nil
	}
	if err == nil && r.writeError {
		r.writeError = false
		r.logger.Info("写入RPC日志成功")
	}
	return len(p) - 1, nil
}

//Close 关闭日志组件
func (r *RemoteLogWriter) Close() error {
	r.once.Do(func() {
		close(r.closeChan)
		if r.rpcInvoker != nil {
			r.rpcInvoker.Close()
		}
	})
	return nil
}
