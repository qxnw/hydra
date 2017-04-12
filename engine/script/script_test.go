package script

import (
	"os"
	"testing"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/ut"
	"github.com/qxnw/lib4go/utility"
)

func TestScript1(t *testing.T) {
	p := newScriptPlugin()
	err := os.MkdirAll("./hydra/merchant.api/rpc/script/", 0777)
	ut.Expect(t, err, nil)
	svs, err := p.Start("./hydra", "merchant.api", "rpc")
	ut.Expect(t, err, nil)
	ut.Expect(t, len(svs), 0)
	os.RemoveAll("./hydra/merchant.api/rpc/script/")
}

func TestScript2(t *testing.T) {
	p := newScriptPlugin()
	err := os.MkdirAll("./hydra/merchant.api/rpc/script/order_request/", 0777)
	ut.Expect(t, err, nil)
	f, err := os.OpenFile("./hydra/merchant.api/rpc/script/order_request/request.lua", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	ut.Expect(t, err, nil)
	f.WriteString(`
		function main()
		 return "hello"
		end
		`)
	f.Close()
	svs, err := p.Start("./hydra", "merchant.api", "rpc")
	ut.Expect(t, err, nil)
	ut.Expect(t, len(svs), 1)
	ut.Expect(t, len(p.services), 1)
	ut.Expect(t, p.services["order_request.request"], "./hydra/merchant.api/rpc/script/order_request/request.lua")
	os.RemoveAll("./hydra/")

	ctx := context.GetContext()
	ctx.Ext["hydra_sid"] = utility.GetGUID()
	r, err := p.Handle("order_request", "request", "order_request", ctx)
	ut.Expect(t, err, nil)
	ut.Expect(t, r.Status, 200)
	ut.Expect(t, r.Content, "hello")
}
