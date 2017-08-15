package ssm

import (
	"fmt"
	"net/url"

	"encoding/json"

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
	content, err := ctx.Input.GetVarParamByArgsName("setting", "setting")
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
func (s *smsProxy) wxSend(name string, mode string, service string, ctx *context.Context) (response *context.StandardReponse, err error) {
	response =context.GetStandardResponse()
	content, err := ctx.Input.GetVarParamByArgsName("setting", "setting")
	if err != nil {
		err = fmt.Errorf("Args参数的属性setting节点未找到:%v", err)
		return
	}
	m := make(map[string]interface{})
	err = json.Unmarshal([]byte(content), &m)
	if err != nil {
		return
	}

	setting, err := conf.NewJSONConfWithJson(content, 0, nil, nil)
	if err != nil {
		err = fmt.Errorf("setting[%s]配置错误，无法解析(err:%v)", content, err)
		return
	}

	if err = s.checkMustField(setting, "service"); err != nil {
		return
	}
	if err = ctx.Input.CheckInput("openId"); err != nil {
		return
	}
	raw, err := setting.GetSectionString(setting.String("alarm", "alarm"))
	if err != nil {
		err = fmt.Errorf("notify配置文件未配置:%s节点", ctx.Input.GetString("alarm", "alarm"))
		return
	}

	ctx.Input.Input.Each(func(k, v string) {
		setting.Set(k, v)
	})
	service = setting.String("service")
	if err != nil {
		return
	}
	err = response.Set(ctx.RPC.Request(service, map[string]string{
		"__raw__": setting.Translate(raw),
	}, true))
	return
}
func (s *smsProxy) wxSend1(name string, mode string, service string, ctx *context.Context) (response *context.StandardReponse, err error) {
	response =context.GetStandardResponse()
	setting, err := s.getwxSendarams(ctx)
	if err != nil {
		return
	}
	if err = s.checkMustField(setting, "host", "appId", "templateId", "data"); err != nil {
		return
	}
	if err = ctx.Input.CheckInput("openId"); err != nil {
		return
	}
	ctx.Input.Input.Each(func(k, v string) {
		setting.Set(k, v)
	})
	u, err := url.Parse(setting.String("host"))
	if err != nil {
		err = fmt.Errorf("wx.host配置错误 %s. err=%v", setting.String("host"), err)
		response.SetStatus(500)
		return response, err
	}
	values := u.Query()
	unionID, _ := ctx.Input.Get("openId")
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
		response.SetStatus(500)
		return
	}
	response.SetContent(status, content)
	if status != 200 {
		err = fmt.Errorf("请求返回错误:status:%d,%s(host:%s)", status, content, setting.String("host"))
		response.SetStatus(status)
		return
	}
	return
}

func (s *smsProxy) wxSend0(name string, mode string, service string, ctx *context.Context) (response *context.StandardReponse, err error) {
	response =context.GetStandardResponse()
	setting, err := s.getwxSendarams(ctx)
	if err != nil {
		return
	}
	if err = s.checkMustField(setting, "host", "appId", "templateId", "data"); err != nil {
		return
	}
	if err = ctx.Input.CheckInput("openId"); err != nil {
		return
	}
	ctx.Input.Input.Each(func(k, v string) {
		setting.Set(k, v)
	})
	u, err := url.Parse(setting.String("host"))
	if err != nil {
		err = fmt.Errorf("wx.host配置错误 %s. err=%v", setting.String("host"), err)
		response.SetStatus(500)
		return response, err
	}
	values := u.Query()
	unionID, _ := ctx.Input.Get("openId")
	data, err := setting.GetSectionString("data")
	if err != nil {
		return
	}
	values.Set("app_id", setting.String("appId"))
	values.Set("open_id", unionID)
	values.Set("template_id", setting.String("templateId"))
	values.Set("content", data)

	u.RawQuery = values.Encode()
	client := http.NewHTTPClient()
	urlParams := u.String()
	fmt.Println("http.get.", urlParams)
	content, status, err := client.Get(urlParams)
	if err != nil {
		err = fmt.Errorf("请求返回错误:status:%d,%s(host:%s,err:%v)", status, content, setting.String("host"), err)
		return
	}
	response.SetContent(status, content)
	if status != 200 {
		err = fmt.Errorf("请求返回错误:status:%d,%s(host:%s)", status, content, setting.String("host"))
		return
	}
	return
}
