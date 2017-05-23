package handlers

import "github.com/qxnw/hydra_plugin/plugins"

type wxNotify struct {
	mustFields []string
}

func newWXNotify() *wxNotify {
	return &wxNotify{
		mustFields: []string{"appid", "session"},
	}
}
func (n *wxNotify) initParams(ctx plugins.Context, invoker plugins.RPCInvoker) (wxCtx *wxContext, err error) {
	wxCtx, err = GetWXContext(ctx, invoker)
	if err != nil {
		return
	}
	err = wxCtx.CheckMustFields(n.mustFields...)
	return
}

func (n *wxNotify) Handle(service string, ctx plugins.Context, invoker plugins.RPCInvoker) (status int, result string, err error) {
	status = 500
	//输入化context,并检查输入参数
	wxContext, err := n.initParams(ctx, invoker)
	if err != nil {
		return
	}
	defer wxContext.Close()
	wxContext.Info("----------------接收微信通知--------------")

	//业务处理

	db, err := wxContext.GetDB()
	if err != nil {
		return
	}
	defer db.Close()
	mc, err := wxContext.GetCache()
	if err != nil {
		return
	}
	result = mc.Get("t")

	//返回结果
	status = 200
	result = "SUCCESS"
	return
}
