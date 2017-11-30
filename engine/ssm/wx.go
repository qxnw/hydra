package ssm

import (
	"fmt"
	"net/url"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/transform"
)

func checkMustField(input conf.Conf, fields ...string) error {
	for _, f := range fields {
		if ok := input.Has(f); !ok {
			return fmt.Errorf("配置:%s不能为空:%v", f, input)
		}
	}
	return nil
}
func checkInputField(input transform.ITransformGetter, fields ...string) error {
	for _, f := range fields {
		if _, err := input.Get(f); err != nil {
			return fmt.Errorf("输入:%s不能为空:%v", f, input)
		}
	}
	return nil
}

//SendWXM 发送微信消息
func SendWXM(ssetting string, param map[string]string) (r string, status int, err error) {
	setting, err := conf.NewJSONConfWithJson(ssetting, 0, nil)
	if err != nil {
		err = fmt.Errorf("setting[%s]配置错误，无法解析(err:%v)", ssetting, err)
		return
	}
	for k, v := range param {
		setting.Set(k, v)
	}
	if err = checkMustField(setting, "host", "app_id", "template_id", "content"); err != nil {
		return
	}
	u, err := url.Parse(setting.String("host"))
	if err != nil {
		err = fmt.Errorf("wx.host配置错误 %s. err=%v", setting.String("host"), err)
		return "", 500, err
	}
	values := u.Query()
	data, err := setting.GetSectionString("content")
	if err != nil {
		return
	}
	color, _ := setting.GetSectionString(fmt.Sprintf("color_%s", setting.String("type", "normal")))
	values.Set("appid", setting.String("app_id"))
	values.Set("openid", setting.String("open_id"))
	values.Set("templateid", setting.String("template_id"))
	values.Set("content", setting.Translate(data))
	if color != "" {
		values.Set("color", color)
	}
	u.RawQuery = values.Encode()
	client := http.NewHTTPClient()
	urlParams := u.String()
	r, status, err = client.Get(urlParams)
	if err != nil {
		err = fmt.Errorf("请求返回错误:status:%d,%s(host:%s,err:%v)", status, ssetting, setting.String("host"), err)
		return
	}
	if status != 200 {
		err = fmt.Errorf("请求返回错误:status:%d,%s(host:%s)", status, ssetting, setting.String("host"))
		return
	}
	return
}

func (s *smsProxy) wxSend(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	content, err := ctx.Input.GetVarParamByArgsName("ssm", "wx")
	if err != nil {
		return
	}
	if err = ctx.Input.CheckInput("open_id"); err != nil {
		return
	}
	input := make(map[string]string)
	ctx.Input.Input.Each(func(k, v string) {
		input[k] = v
	})
	r, status, err := SendWXM(content, input)
	if err != nil {
		response.SetError(status, err)
		return response, err
	}
	response.SetContent(status, r)
	return
}
