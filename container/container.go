package container

import "github.com/qxnw/hydra/context"

type Response struct {
}

//Container 容器接口
type Container interface {
	PreLoad(service string)
	Execute(service string, context *context.Context) *Response
}
