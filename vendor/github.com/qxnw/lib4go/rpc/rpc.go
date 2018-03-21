package rpc

import "time"

type IRPCResponse interface {
	Wait(timeout time.Duration) (int, string, error)
	GetResult() chan IRPCResult
}
type IRPCResult interface {
	GetService() string
	GetStatus() int
	GetResult() string
	GetErr() error
}
