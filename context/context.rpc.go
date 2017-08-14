package context

import (
	"fmt"

	"github.com/qxnw/lib4go/jsons"
)

//ContextRPC MQ操作实例
type ContextRPC struct {
	ctx *Context
}

//Reset 重置context
func (cr *ContextRPC) Reset(ctx *Context) {
	cr.ctx = ctx
}

//PreInit 预加载服务
func (cr *ContextRPC) PreInit(services ...string) error {
	return cr.ctx.rpc.PreInit()
}

//RequestFailRetry RPC请求
func (cr *ContextRPC) RequestFailRetry(service string, input map[string]string, times int) (status int, r string, param map[string]string, err error) {
	if _, ok := input["hydra_sid"]; !ok {
		input["hydra_sid"] = cr.ctx.GetSessionID()
	}
	if _, ok := input["__body"]; !ok {
		input["__body"] = cr.ctx.Input.Body
	}
	status, r, param, err = cr.ctx.rpc.RequestFailRetry(service, input, times)
	if err != nil || status != 200 {
		err = fmt.Errorf("rpc请求(%s)失败:%d,err:%v", service, status, err)
		return
	}
	return
}

//Request RPC请求
func (cr *ContextRPC) Request(service string, input map[string]string, failFast bool) (status int, r string, param map[string]string, err error) {
	if _, ok := input["hydra_sid"]; !ok {
		input["hydra_sid"] = cr.ctx.GetSessionID()
	}
	if _, ok := input["__body"]; !ok {
		input["__body"] = cr.ctx.Input.Body
	}
	status, r, param, err = cr.ctx.rpc.Request(service, input, failFast)
	if err != nil || status != 200 {
		err = fmt.Errorf("rpc请求(%s)失败:%d,err:%v", service, status, err)
		return
	}
	return
}

//RequestMap RPC请求返回结果转换为map
func (cr *ContextRPC) RequestMap(service string, input map[string]string, failFast bool) (status int, r map[string]interface{}, param map[string]string, err error) {
	if _, ok := input["hydra_sid"]; !ok {
		input["hydra_sid"] = cr.ctx.GetSessionID()
	}
	if _, ok := input["__body"]; !ok {
		input["__body"] = cr.ctx.Input.Body
	}
	status, result, _, err := cr.Request(service, input, failFast)
	if err != nil {
		return
	}
	r, err = jsons.Unmarshal([]byte(result))
	if err != nil {
		err = fmt.Errorf("rpc请求返结果不是有效的json串:%s,%v,%s,err:%v", service, input, result, err)
		return
	}
	return
}
