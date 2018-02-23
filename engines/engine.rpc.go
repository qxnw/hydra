package engines

import (
	"fmt"
	"strings"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/types"
)

//RPCProxy rpc 代理服务
func (r *ServiceEngine) RPCProxy() component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		header, err := ctx.Request.Http.GetHeader()
		if err != nil {
			response.SetContent(500, err)
			return
		}
		status, result, params, err := ctx.RPC.Request(service, strings.ToUpper(ctx.Request.Ext.GetMethod()), header, ctx.Request.Ext.GetBodyMap(), true)
		if err != nil {
			err = fmt.Errorf("rpc执行错误status：%d,result:%v,err:%v", status, result, err)
		}
		response.SetParams(types.GetIMap(params))
		if err != nil {
			response.SetContent(status, err)
			return response, err
		}
		response.SetContent(status, result)
		return response, err
	}
}
