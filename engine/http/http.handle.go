package http

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"time"

	"strings"

	"bytes"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/encoding/base64"
	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/security/aes"
	"github.com/qxnw/lib4go/security/des"
	"github.com/qxnw/lib4go/security/md5"
	"github.com/qxnw/lib4go/security/rsa"
	"github.com/qxnw/lib4go/transform"
)

//httpHandle get请求,input获取参数
//body
func (s *httpProxy) httpHandle(name string, mode string, service string, ctx *context.Context) (response *context.WebResponse, err error) {
	response = context.GetWebResponse(ctx)
	content, err := ctx.Input.GetVarParamByArgsName("setting", "setting")
	if err != nil {
		return
	}
	config, err := conf.NewJSONConfWithJson(content, 0, nil)
	if err != nil {
		return
	}
	hostURL, err := config.Get("url") //获取当前请求的URL
	if err != nil || hostURL == "" {
		err = fmt.Errorf("args.seting(%s) http 模块配置错误，未指定url参数(err:%v)", content, err)
		return
	}
	u, err := url.Parse(hostURL)
	if err != nil {
		err = fmt.Errorf("args.seting(%s) http 模块配置错误，url参数(%s)配置错误(err:%v)", hostURL, content, err)
		return
	}
	url := u.String()

	header, _ := config.GetSMap("header") //获取header标签，可以为空
	if header == nil {
		header = make(map[string]string)
	}
	method := config.String("method", "get") //获取请求方式，可以为空
	charset := header["Charset"]
	if charset == "" {
		charset = "utf-8"
	}

	input, err := config.GetSection("data") //获取data标签，可以为空
	if err != nil {
		input = conf.NewJSONConfWithHandle(make(map[string]interface{}), 0, nil)
		return
	}

	paraTransform := transform.NewGetter(ctx.Input.Params)
	paraTransform.Append(ctx.Input.Input)
	values, raw, err := s.GetData(u.Query(), input, paraTransform)
	if err != nil {
		return
	}
	requestData := values.Encode()
	client := http.NewHTTPClient()
	header["Cookie"] = fmt.Sprintf("hydra_sid=%s", ctx.Input.Ext["hydra_sid"])
	hc, t, err := client.Request(method, url+"?"+requestData, "", charset, header)
	if err != nil {
		response.SetStatus(t)
		return
	}
	result := make(map[string]interface{})
	result["url"] = url
	result["data"] = requestData
	result["charset"] = charset
	result["raw"] = raw
	if strings.Contains(hc, "{") || strings.Contains(hc, "[") {
		result["content"] = json.RawMessage(hc)
	} else {
		result["content"] = hc
	}
	buff, err := json.Marshal(result)
	if err != nil {
		return
	}
	response.Success(string(buff))
	return
}
func (s *httpProxy) GetData(u url.Values, data conf.Conf, trs *transform.Transform) (ua url.Values, rawStr string, err error) {
	encrypt := strings.ToLower(data.String("_encrypt", "md5"))
	hasEncrypt := false
	for _, v := range s.encrypts {
		if encrypt == v {
			hasEncrypt = true
			break
		}
	}
	if !hasEncrypt {
		err := fmt.Errorf("%s加密方式不支持，只支持:%v", encrypt, s.encrypts)
		return u, "", err
	}
	kc := data.String("_c", "=")
	kvConnect := data.String("_k", "&")
	sorted := data.String("_sorted", "true") == "true"
	hasTimestamp := data.String("_timestamp", "true") == "true"
	kvs := make([]string, 0, data.Len())
	if hasTimestamp {
		trs.Set("timestamp", time.Now().Format("20060102150405"))
	}
	data.Each(func(k string) {
		if !strings.EqualFold(k, "sign") && !strings.HasPrefix(k, "_") {
			kvs = append(kvs, fmt.Sprintf("%s", k))
			u.Add(k, trs.Translate(data.String(k)))
		}
	})
	data.Append(trs.Data)
	if data.Has("sign") {
		if sorted {
			sort.Slice(kvs, func(i, j int) bool {
				return kvs[i] < kvs[j]
			})
		}
		raw := bytes.NewBufferString("")
		for i, k := range kvs {
			raw.WriteString(k)
			raw.WriteString(kc)
			raw.WriteString(trs.Translate(data.String(k)))
			if i < len(kvs)-1 {
				raw.WriteString(kvConnect)
			}
		}
		rawStr = raw.String()
		data.Set("_raw", rawStr)
		var sign string
		switch encrypt {
		case "md5":
			u.Add("sign", md5.Encrypt(data.String("sign")))
		case "base64":
			u.Add("sign", base64.URLEncode(data.String("sign")))
		case "rsa/sha1":
			if !data.Has("_key") {
				return u, "", fmt.Errorf("rsa私钥不能为空")
			}
			if sign, err = rsa.Sign(data.String("sign"), data.String("_key"), "sha1"); err != nil {
				return u, "", err
			}
			u.Add("sign", sign)

		case "rsa/md5":
			if !data.Has("_key") {
				return u, "", fmt.Errorf("rsa私钥不能为空")
			}
			if sign, err = rsa.Sign(data.String("sign"), data.String("_key"), "md5"); err != nil {
				return u, "", err
			}
			u.Add("sign", sign)
		case "aes":
			if !data.Has("_key") {
				return u, "", fmt.Errorf("aes密钥不能为空")
			}
			if sign, err = aes.Encrypt(data.String("sign"), data.String("_key")); err != nil {
				return u, "", err
			}
			u.Add("sign", sign)
		case "des":
			if !data.Has("_key") {
				return u, "", fmt.Errorf("des密钥不能为空")
			}
			if sign, err = des.Encrypt(data.String("sign"), data.String("_key"), "ecb/pkcs5"); err != nil {
				return u, "", err
			}
			u.Add("sign", sign)
		}
	}
	return u, rawStr, nil
}
