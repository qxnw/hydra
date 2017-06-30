package sms

import (
	"fmt"
	"net/url"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/transform"
)

func (s *smsProxy) checkMustField(input conf.Conf, fields ...string) error {
	for _, f := range fields {
		if ok := input.Has(f); !ok {
			return fmt.Errorf("配置:%s不能为空:%v", f, input)
		}
	}
	return nil
}
func (s *smsProxy) checkInputField(input transform.ITransformGetter, fields ...string) error {
	for _, f := range fields {
		if _, err := input.Get(f); err != nil {
			return fmt.Errorf("输入:%s不能为空:%v", f, input)
		}
	}
	return nil
}
func (s *smsProxy) getwxSendarams(ctx *context.Context) (settings conf.Conf, err error) {
	setting, ok := ctx.GetArgs()["setting"]
	if !ok {
		err = fmt.Errorf("Args参数未配置setting属性")
		return
	}
	content, err := s.getVarParam(ctx, setting)
	if err != nil {
		err = fmt.Errorf("Args参数的属性setting节点未找到:%v", err)
		return
	}
	settings, err = conf.NewJSONConfWithJson(content, 0, nil, nil)
	if err != nil {
		err = fmt.Errorf("setting[%s]配置错误，无法解析(err:%v)", content, err)
		return
	}
	return
}
func (s *smsProxy) wxSend(ctx *context.Context) (r string, t int, err error) {
	setting, err := s.getwxSendarams(ctx)
	if err != nil {
		return
	}
	if err = s.checkMustField(setting, "host", "appId", "templateId", "data"); err != nil {
		return
	}
	if err = s.checkInputField(ctx.GetInput(), "openId"); err != nil {
		return
	}
	ctx.GetInput().Each(func(k, v string) {
		setting.Set(k, v)
	})
	u, err := url.Parse(setting.String("host"))
	if err != nil {
		err = fmt.Errorf("wx.host配置错误 %s. err=%v", setting.String("host"), err)
		return "", 500, err
	}
	values := u.Query()
	unionID, _ := ctx.GetInput().Get("openId")
	data, err := setting.GetSectionString("data")
	if err != nil {
		return
	}
	values.Set("app_id", setting.String("appId"))
	values.Set("union_id", unionID)
	values.Set("template_id", setting.String("templateId"))
	values.Set("content", data)

	u.RawQuery = values.Encode()
	client := http.NewHTTPClient()
	content, status, err := client.Get(u.String())
	if err != nil {
		err = fmt.Errorf("请求返回错误:status:%d,%s(host:%s,err:%v)", status, content, setting.String("host"), err)
		return "", 500, err
	}
	if status != 200 {
		err = fmt.Errorf("请求返回错误:status:%d,%s(host:%s)", status, content, setting.String("host"))
		return "", status, err
	}
	return content, status, err
}
