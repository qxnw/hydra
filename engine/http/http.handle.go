package http

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/transform"
)

func (s *httpProxy) httpHandle(service string, ctx *context.Context) (r string, t int, err error) {
	if ctx.Input.Input == nil || ctx.Input.Args == nil {
		err = fmt.Errorf("input,args不能为空:%v", ctx.Input)
		return
	}
	params, ok := ctx.Input.Args.(map[string]string)
	if !ok {
		err = fmt.Errorf("args类型错误必须为map[string]string:%v", ctx.Input)
		return
	}

	if _, ok := params["host"]; !ok {
		err = fmt.Errorf("args配置错误，未指定host参数的值:%v", params)
		return
	}
	u, err := url.Parse(params["host"] + "/" + strings.Trim(service, "/"))
	if err != nil {
		return
	}
	url := u.String()
	values := u.Query()
	input, ok := ctx.Input.Input.(transform.ITransformGetter)
	if !ok {
		err = fmt.Errorf("input类型错误必须为transform.ITransformGetter:%v", ctx.Input)
		return
	}
	input.Each(func(k, v string) {
		values.Set(k, v)
	})

	for k, v := range params {
		if k != "host" && k != "method" && k != "header" {
			values.Set(k, v)
		}
	}
	header := make(map[string]string)
	if v, ok := params["header"]; ok {
		hds, err := s.getVarParam(ctx, v)
		if err != nil {
			return "", 500, err
		}
		confs, err := conf.NewJSONConfWithJson(hds, 0, nil)
		if err != nil {
			return "", 500, err
		}
		confs.Each(func(k string) {
			header[k] = confs.String(k)
		})
	}

	method := params["method"]
	if method == "" {
		method = "get"
	}
	content := values.Encode()
	if len(values) == 0 && ctx.Input.Body != nil {
		content = ctx.Input.Body.(string)
	}
	charset := header["charset"]
	if charset == "" {
		charset = "utf-8"
	}
	client := http.NewHTTPClient()
	return client.Request(method, url, content, charset, header)

}

func (s *httpProxy) getVarParam(ctx *context.Context, name string) (string, error) {
	funcVar := ctx.Ext["__func_var_get_"]
	if funcVar == nil {
		return "", errors.New("未找到__func_var_get_")
	}
	if f, ok := funcVar.(func(c string, n string) (string, error)); ok {
		return f("header", name)
	}
	return "", errors.New("未找到__func_var_get_传入类型错误")
}
