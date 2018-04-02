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
	return func(name string, mode string, service string, ctx *context.Context) (response interface{}) {
		header, _ := ctx.Request.Http.GetHeader()
		input := ctx.Request.Ext.GetBodyMap()
		status, result, params, err := ctx.RPC.Request(service, strings.ToUpper(ctx.Request.Ext.GetMethod()), header, input, true)
		if err != nil {
			err = fmt.Errorf("rpc执行错误status：%d,result:%v,err:%v", status, result, err)
		}
		ctx.Response.SetParams(types.GetIMap(params))
		if err != nil {
			ctx.Response.SetStatus(status)
			return err
		}
		ctx.Response.SetStatus(status)
		return result
	}
}
