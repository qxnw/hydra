package engines

import (
	"fmt"
	"strings"
	"time"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/types"
)

//RPCProxy rpc 代理服务
func (r *ServiceEngine) RPCProxy() component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (r interface{}) {
		header, _ := ctx.Request.Http.GetHeader()
		cookie, _ := ctx.Request.Http.GetCookies()
		for k, v := range cookie {
			header[k] = v
		}
		input := ctx.Request.Ext.GetBodyMap()
		timeout := ctx.Request.Setting.GetInt("timeout", 3)
		response := ctx.RPC.AsyncRequest(service, strings.ToUpper(ctx.Request.Ext.GetMethod()), header, input, true)
		status, result, params, err := response.Wait(time.Second * time.Duration(timeout))
		if err != nil {
			err = fmt.Errorf("rpc.proxy %v(%d)", err, status)
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
