package engines

import (
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

//RPCProxy rpc 代理服务
func (r *ServiceEngine) RPCProxy() component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		input := make(map[string]string)
		ctx.Request.Form.Each(func(k string, v string) {
			input[k] = v
		})
		status, result, params, err := r.Invoker.Request(name, input, true)
		if err != nil {
			err = fmt.Errorf("rpc执行错误status：%d,result:%v,err:%v", status, result, err)
		}
		response.Set(status, result, params, err)
		return response, err
	}
}
