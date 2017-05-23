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
	wxCtx, err = getWXContext(ctx, invoker)
	if err != nil {
		return
	}
	err = wxCtx.CheckMustFields(n.mustFields...)
	return
}

func (n *wxNotify) Handle(service string, ctx plugins.Context, invoker plugins.RPCInvoker) (status int, result string, err error) {
	status = 500
	wxContext, err := n.initParams(ctx, invoker)
	if err != nil {
		return
	}
	defer wxContext.Close()
	//业务处理
	return 200, "SUCCESS", nil
}
