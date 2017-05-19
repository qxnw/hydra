package sms

/*配置文件内容：{
    "appid": "8a48b5514e5298b9014e67a3f02f1411",
    "main_account": "aaf98fda42c744c90142d505bfab0135",
    "main_account_token": "5cd9cb617def42678553fe4d93b8f291",
    "soft_version": "2013-12-26",
    "sign": "{@main_account}{@main_account_token}{@timestamp}",
    "auth":"@main_account:{@timestamp}",
    "header":"Accept:application/xml\r\nContent-type:application/xml;charset=utf-8\r\nAuthorization:@auth",
    "url": "https://app.cloopen.com:8883/{@soft_version}/Accounts/{@main_account}/SMS/TemplateSMS?sig=@sign",
    "body": "<?xml version='1.0' encoding='utf-8'?><TemplateSMS><to>{@mobile}</to><appId>{@appid}</appId><templateId>65871</templateId><datas>{@data}</datas></TemplateSMS>"
}*/

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"

	"strings"

	"time"

	"github.com/qxnw/lib4go/encoding/base64"
	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/security/md5"
	"github.com/qxnw/lib4go/transform"
)

type eSMS struct {
	mobile  string
	data    string
	url     string
	body    string
	charset string
	header  map[string]string
}

func (s *smsProxy) getGetParams(ctx *context.Context) (sms *eSMS, err error) {
	if ctx.Input.Input == nil || ctx.Input.Args == nil || ctx.Input.Params == nil {
		err = fmt.Errorf("engine:cache.input,params,args不能为空:%v", ctx.Input)
		return
	}
	sms = &eSMS{header: make(map[string]string)}
	input := ctx.Input.Input.(transform.ITransformGetter)
	sms.mobile, err = input.Get("mobile")
	if err != nil || sms.mobile == "" {
		err = fmt.Errorf("接收人手机号不能为空")
		return
	}

	data, err := input.Get("data")
	if err != nil || data == "" {
		err = fmt.Errorf("短信内容(data)不能为空")
		return
	}
	datas := strings.Split(data, ";")
	for _, v := range datas {
		sms.data = fmt.Sprintf("%s<data>%s</data>", sms.data, v)
	}
	params, ok := ctx.Input.Args.(map[string]string)
	if !ok {
		err = fmt.Errorf("未设置Args参数")
		return
	}
	setting, ok := params["setting"]
	if !ok {
		err = fmt.Errorf("Args.setting配置不能为空")
		return
	}
	content, err := s.getVarParam(ctx, setting)
	if err != nil {
		return
	}

	form, err := conf.NewJSONConfWithJson(content, 0, nil)
	if err != nil {
		err = fmt.Errorf("setting[%s]配置错误，无法解析(err:%v)", content, err)
		return
	}
	form.Set("mobile", sms.mobile)
	form.Set("data", sms.data)
	form.Set("timestamp", time.Now().Format("20060102150405"))

	sign := form.String("sign")
	if sign == "" || strings.Contains(sign, "@") {
		err = fmt.Errorf("setting[%s]配置错误，sign配置错误(sign:%s)", setting, sign)
		return
	}
	form.Set("sign", strings.ToUpper(md5.Encrypt(sign)))

	sms.url = form.String("url")
	if sms.url == "" || strings.Contains(sms.url, "@") {
		err = fmt.Errorf("setting[%s]配置错误，url配置错误(url:%s)", setting, sms.url)
		return
	}
	sms.body = form.String("body")
	if sms.body == "" || strings.Contains(sms.body, "@") {
		err = fmt.Errorf("setting[%s]配置错误，body配置错误(body:%s)", setting, sms.body)
		return
	}
	auth := form.String("auth")
	if auth == "" || strings.Contains(auth, "@") {
		err = fmt.Errorf("setting[%s]配置错误，auth配置错误(auth:%s)", setting, auth)
		return
	}
	form.Set("auth", base64.Encode(auth))
	header := form.String("header")
	if header == "" || strings.Contains(header, "@") {
		err = fmt.Errorf("setting[%s]配置错误，header配置错误(header:%s)", setting, header)
		return
	}
	headers := strings.Split(header, "\r\n")
	for _, v := range headers {
		hs := strings.SplitN(v, ":", 2)
		if len(hs) != 2 {
			err = fmt.Errorf("setting[%s]配置错误，header配置错误,不是有效的键值对(header:%s)", setting, header)
			return
		}
		sms.header[hs[0]] = hs[1]
	}
	sms.charset = form.String("charset", "utf-8")
	return

}

func (s *smsProxy) getVarParam(ctx *context.Context, name string) (string, error) {
	funcVar := ctx.Ext["__func_var_get_"]
	if funcVar == nil {
		return "", errors.New("未找到__func_var_get_")
	}
	if f, ok := funcVar.(func(c string, n string) (string, error)); ok {
		return f("setting", name)
	}
	return "", errors.New("未找到__func_var_get_传入类型错误")
}

func (s *smsProxy) ytxSend(ctx *context.Context) (r string, st int, err error) {
	m, err := s.getGetParams(ctx)
	if err != nil {
		err = fmt.Errorf("engine:ytx.sms.%v", err)
		return
	}
	client := http.NewHTTPClient()
	r, st, err = client.Request("post", m.url, m.body, m.charset, m.header)
	if err != nil {
		err = fmt.Errorf("engine:ytx.sms.%v(url:%s,body:%s,header:%s)", err, m.url, m.body, m.header)
		return
	}
	return
}
