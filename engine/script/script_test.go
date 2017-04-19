package script

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/encoding"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/ut"
	"github.com/qxnw/lib4go/utility"
	"github.com/qxnw/lua4go"
	"github.com/qxnw/lua4go/bind"
)

func TestScript1(t *testing.T) {
	p := newScriptWorker()
	err := os.MkdirAll("./hydra/merchant.api/rpc/script/", 0777)
	ut.Expect(t, err, nil)
	svs, err := p.Start("./hydra", "merchant.api", "rpc")
	ut.Expect(t, err, nil)
	ut.Expect(t, len(svs), 0)
	os.RemoveAll("./hydra/merchant.api/rpc/script/")
}

func TestScript2(t *testing.T) {
	p := newScriptWorker()
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

//检查基本输入参数
func TestScript3(t *testing.T) {
	formData := map[string]string{
		"id":   "100",
		"name": "colin",
	}
	paramData := map[string]string{
		"action":     "request",
		"controller": "order",
	}
	tfForm := transform.NewMap(formData)
	tfParams := transform.NewMap(paramData)
	body := `<?xml version="1.0" encoding="UTF-8"?>
<recipe>
</recipe>`
	buf := []byte(body)
	ctx := context.GetContext()
	ctx.Ext["hydra_sid"] = utility.GetGUID()
	ctx.Ext["__txt_body_"] = string(buf)
	ctx.Ext["__func_param_getter_"] = tfParams
	ctx.Ext["__func_args_getter_"] = tfForm
	rservice := tfForm.Translate(tfParams.Translate("./t/@controller.@action.lua"))
	rArgs := tfForm.Translate(tfParams.Translate("db=@action/@name"))

	ctx.Input.Input = tfForm.Data
	ctx.Input.Body = string(buf)
	ctx.Input.Params = tfParams.Data
	var err error
	ctx.Input.Args, err = utility.GetMapWithQuery(rArgs)

	log := logger.GetSession("test", ctx.Ext["hydra_sid"].(string))
	defer log.Close()
	fmt.Println(ctx.Input.ToJson())
	input := lua4go.NewContextWithLogger(ctx.Input.ToJson(), ctx.Ext, log)
	vm := lua4go.NewLuaVM(bind.NewDefault(), 1, 100, time.Second*300)
	result, m, err := vm.Call(rservice, input)
	ut.Expect(t, err, nil)
	ut.Expect(t, len(result), 1)
	ut.Expect(t, len(m), 2)
	ut.Expect(t, m["Content-type"], "text/plain")
	ut.Expect(t, m["Charset"], "gbk")

	data, err := jsons.Unmarshal([]byte(result[0]))
	ut.Expect(t, err, nil)
	ut.Expect(t, data["body"], body)
	ut.RefuteSkip(t, data["args"].(map[string]interface{}), nil)
	ut.Expect(t, data["args"].(map[string]interface{})["db"], "request/colin")

	ut.RefuteSkip(t, data["input"].(map[string]interface{}), nil)
	ut.Expect(t, data["input"].(map[string]interface{})["id"], "100")

	ut.RefuteSkip(t, data["params"].(map[string]interface{}), nil)
	ut.Expect(t, data["params"].(map[string]interface{})["action"], "request")

}

func TestScript4(t *testing.T) {
	formData := map[string]string{
		"id":   "100",
		"name": "colin",
	}
	paramData := map[string]string{
		"action":     "query",
		"controller": "order",
	}
	tfForm := transform.NewMap(formData)
	tfParams := transform.NewMap(paramData)
	body := `<?xml version="1.0" encoding="UTF-8"?>
<recipe>
</recipe>`
	buf := []byte(body)
	ctx := context.GetContext()
	ctx.Ext["hydra_sid"] = utility.GetGUID()
	ctx.Ext["__txt_body_"] = string(buf)
	ctx.Ext["__func_param_getter_"] = tfParams
	ctx.Ext["__func_args_getter_"] = tfForm
	rservice := tfForm.Translate(tfParams.Translate("./t/@controller.@action.lua"))
	rArgs := tfForm.Translate(tfParams.Translate("db=@action/@name"))

	ctx.Input.Input = tfForm.Data
	ctx.Input.Body = string(buf)
	ctx.Input.Params = tfParams.Data
	var err error
	ctx.Input.Args, err = utility.GetMapWithQuery(rArgs)

	log := logger.GetSession("test", ctx.Ext["hydra_sid"].(string))
	defer log.Close()
	fmt.Println(ctx.Input.ToJson())
	input := lua4go.NewContextWithLogger(ctx.Input.ToJson(), ctx.Ext, log)
	vm := lua4go.NewLuaVM(bind.NewDefault(), 1, 100, time.Second*300)
	result, m, err := vm.Call(rservice, input)
	ut.Expect(t, err, nil)
	ut.Expect(t, len(result), 1)
	ut.Expect(t, len(m), 2)
	ut.Expect(t, m["Location"], "/order/request")

}

func TestScript5(t *testing.T) {
	formData := map[string]string{
		"id":   "100",
		"name": "colin",
	}
	paramData := map[string]string{
		"action":     "request",
		"controller": "order",
	}
	tfForm := transform.NewMap(formData)
	tfParams := transform.NewMap(paramData)
	body := `<?xml version="1.0" encoding="UTF-8"?>
<recipe>
</recipe>`
	buf := []byte(body)
	ctx := context.GetContext()
	ctx.Ext["hydra_sid"] = utility.GetGUID()
	ctx.Ext["__txt_body_"] = string(buf)
	ctx.Ext["__func_param_getter_"] = tfParams
	ctx.Ext["__func_args_getter_"] = tfForm
	ctx.Ext["__func_body_get_"] = func(c string) (string, error) {
		return encoding.Convert(buf, c)
	}
	rservice := tfForm.Translate(tfParams.Translate("./t/@controller.@action.lua"))
	rArgs := tfForm.Translate(tfParams.Translate("db=@action/@name"))

	ctx.Input.Input = tfForm.Data
	ctx.Input.Body = string(buf)
	ctx.Input.Params = tfParams.Data
	var err error
	ctx.Input.Args, err = utility.GetMapWithQuery(rArgs)

	log := logger.GetSession("test", ctx.Ext["hydra_sid"].(string))
	defer log.Close()
	fmt.Println(ctx.Input.ToJson())
	input := lua4go.NewContextWithLogger(ctx.Input.ToJson(), ctx.Ext, log)
	vm := lua4go.NewLuaVM(bind.NewDefault(), 1, 100, time.Second*300)
	result, m, err := vm.Call(rservice, input)
	ut.Expect(t, err, nil)
	ut.Expect(t, len(result), 1)
	ut.Expect(t, len(m), 2)
	ut.Refute(t, result[0], "")

}

func TestScript6(t *testing.T) {
	formData := map[string]string{
		"id":   "100",
		"name": "colin",
	}
	paramData := map[string]string{
		"action":     "body",
		"controller": "order",
	}
	tfForm := transform.NewMap(formData)
	tfParams := transform.NewMap(paramData)
	body := `<?xml version="1.0" encoding="UTF-8"?>
<recipe>
</recipe>`
	buf := []byte(body)
	ctx := context.GetContext()
	ctx.Ext["hydra_sid"] = utility.GetGUID()
	ctx.Ext["__txt_body_"] = string(buf)
	ctx.Ext["__func_param_getter_"] = tfParams
	ctx.Ext["__func_args_getter_"] = tfForm
	ctx.Ext["__func_body_get_"] = func(c string) (string, error) {
		return encoding.Convert(buf, c)
	}
	rservice := tfForm.Translate(tfParams.Translate("./t/@controller.@action.lua"))
	rArgs := tfForm.Translate(tfParams.Translate("db=@action/@name"))

	ctx.Input.Input = tfForm.Data
	ctx.Input.Body = string(buf)
	ctx.Input.Params = tfParams.Data
	var err error
	ctx.Input.Args, err = utility.GetMapWithQuery(rArgs)

	log := logger.GetSession("test", ctx.Ext["hydra_sid"].(string))
	defer log.Close()
	fmt.Println(ctx.Input.ToJson())
	input := lua4go.NewContextWithLogger(ctx.Input.ToJson(), ctx.Ext, log)
	vm := lua4go.NewLuaVM(bind.NewDefault(), 1, 100, time.Second*300)
	result, _, err := vm.Call(rservice, input)
	ut.Expect(t, err, nil)
	ut.Expect(t, len(result), 1)
	ut.Expect(t, result[0], body)
}
