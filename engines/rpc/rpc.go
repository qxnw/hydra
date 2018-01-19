package rpc

import (
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

//Proxy rpc 代理服务
func Proxy() component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		input := make(map[string]string)
		ctx.Input.Input.Each(func(k string, v string) {
			input[k] = v
		})
		status, result, params, err := ctx.RPC.Request(name, input, true)
		if err != nil {
			err = fmt.Errorf("rpc执行错误status：%d,result:%v,err:%v", status, result, err)
		}
		response.Set(status, result, params, err)
		return response, err
	}
}
