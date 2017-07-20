package ssm

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
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"

	"strings"

	"time"

	"github.com/qxnw/lib4go/encoding/base64"
	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/security/md5"
)

type eSMS struct {
	mobile  string
	data    string
	url     string
	body    string
	charset string
	header  map[string]string
}

func (s *smsProxy) getYtxParams(ctx *context.Context) (sms *eSMS, err error) {
	sms = &eSMS{header: make(map[string]string)}
	sms.mobile, err = ctx.GetInput().Get("mobile")
	if err != nil || sms.mobile == "" {
		err = fmt.Errorf("接收人手机号不能为空")
		return
	}

	data, err := ctx.GetInput().Get("data")
	if err != nil || data == "" {
		err = fmt.Errorf("短信内容(data)不能为空")
		return
	}
	datas := strings.Split(data, ";")
	for _, v := range datas {
		sms.data = fmt.Sprintf("%s<data>%s</data>", sms.data, v)
	}
	content, err := ctx.GetVarParamByArgsName("setting", "setting")
	if err != nil {
		return
	}

	form, err := conf.NewJSONConfWithJson(content, 0, nil, nil)
	if err != nil {
		err = fmt.Errorf("setting[%s]配置错误，无法解析(err:%v)", content, err)
		return
	}
	form.Set("mobile", sms.mobile)
	form.Set("data", sms.data)
	form.Set("timestamp", time.Now().Format("20060102150405"))

	sign := form.String("sign")
	if sign == "" || strings.Contains(sign, "@") {
		err = fmt.Errorf("args.setting配置错误，sign配置错误(sign:%s)(%s)", sign, content)
		return
	}
	form.Set("sign", strings.ToUpper(md5.Encrypt(sign)))

	sms.url = form.String("url")
	if sms.url == "" || strings.Contains(sms.url, "@") {
		err = fmt.Errorf("args.setting配置错误，url配置错误(url:%s)(%s)", sms.url, content)
		return
	}
	sms.body = form.String("body")
	if sms.body == "" || strings.Contains(sms.body, "@") {
		err = fmt.Errorf("args.setting配置错误，body配置错误(body:%s)(%s)", sms.body, content)
		return
	}
	auth := form.String("auth")
	if auth == "" || strings.Contains(auth, "@") {
		err = fmt.Errorf("args.setting配置错误，auth配置错误(auth:%s)(%s)", auth, content)
		return
	}
	form.Set("auth", base64.Encode(auth))
	header := form.String("header")
	if header == "" || strings.Contains(header, "@") {
		err = fmt.Errorf("args.setting配置错误，header配置错误(header:%s)(%s)", header, content)
		return
	}
	headers := strings.Split(header, "\r\n")
	for _, v := range headers {
		hs := strings.SplitN(v, ":", 2)
		if len(hs) != 2 {
			err = fmt.Errorf("args.setting配置错误，header配置错误,不是有效的键值对(header:%s)", header)
			return
		}
		sms.header[hs[0]] = hs[1]
	}
	sms.charset = form.String("charset", "utf-8")
	return

}

func (s *smsProxy) ytxSend(ctx *context.Context) (r string, st int, err error) {
	m, err := s.getYtxParams(ctx)
	if err != nil {
		return
	}
	client := http.NewHTTPClient()
	r, st, err = client.Request("post", m.url, m.body, m.charset, m.header)
	if err != nil {
		err = fmt.Errorf("%v(url:%s,body:%s,header:%s)", err, m.url, m.body, m.header)
		return
	}
	return
}
