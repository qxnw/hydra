package http

import (
	"errors"
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

func (s *httpProxy) httpHandle(ctx *context.Context) (r string, t int, err error) {
	if ctx.Input.Input == nil || ctx.Input.Args == nil || ctx.Input.Params == nil {
		err = fmt.Errorf("input,params,args不能为空:%v", ctx.Input)
		return
	}

	args, ok := ctx.Input.Args.(map[string]string)
	if !ok {
		err = fmt.Errorf("args类型错误必须为map[string]string:%v", ctx.Input)
		return
	}
	setting, ok := args["setting"]
	if !ok {
		err = fmt.Errorf("args配置错误，未指定setting参数的值:%v", args)
		return
	}
	content, err := s.getVarParam(ctx, setting)
	if err != nil {
		return
	}
	config, err := conf.NewJSONConfWithJson(content, 0, nil)
	if err != nil {
		return
	}
	hostURL, err := config.Get("url") //获取当前请求的URL
	if err != nil || hostURL == "" {
		err = fmt.Errorf("args.seting(%s) http 模块配置错误，未指定url参数:%v(err:%v)", setting, content, err)
		return
	}
	u, err := url.Parse(hostURL)
	if err != nil {
		err = fmt.Errorf("args.seting(%s) http 模块配置错误，url参数(%s)配置错误:%v(err:%v)", setting, hostURL, content, err)
		return
	}
	url := u.String()

	header, _ := config.GetSMap("header") //获取header标签，可以为空
	if header == nil {
		header = make(map[string]string)
	}
	method := config.String("method", "get") //获取请求方式，可以为空
	charset := header["charset"]
	if charset == "" {
		charset = "utf-8"
	}

	input, err := config.GetSection("data") //获取data标签，可以为空
	if err != nil {
		input = conf.NewJSONConfWithHandle(make(map[string]interface{}), 0, nil)
		return
	}
	input.Append(ctx.Input.Input.(transform.ITransformGetter))
	input.Append(ctx.Input.Params.(transform.ITransformGetter))
	values, err := s.GetData(u.Query(), input)
	if err != nil {
		return
	}
	requestData := values.Encode()
	client := http.NewHTTPClient()
	return client.Request(method, url, requestData, charset, header)

}
func (s *httpProxy) GetData(u url.Values, data conf.Conf) (ua url.Values, err error) {
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
		return u, err
	}
	kc := data.String("_c", "=")
	kvConnect := data.String("_k", "&")
	sorted := data.String("_sorted", "true") == "true"
	hasTimestamp := data.String("_timestamp", "true") == "true"
	kvs := make([]string, 0, data.Len())
	data.Each(func(k string) {
		if !strings.EqualFold(k, "sign") && !strings.HasPrefix(k, "_") {
			kvs = append(kvs, fmt.Sprintf("%s", k))
			u.Add(k, data.String(k))
		}
	})

	if data.Has("sign") {
		if hasTimestamp {
			data.Set("timestamp", time.Now().Format("20160102150405"))
			kvs = append(kvs, "timestamp")
		}
		if sorted {
			sort.Slice(kvs, func(i, j int) bool {
				return kvs[i] < kvs[j]
			})
		}
		raw := bytes.NewBufferString("")
		for i, k := range kvs {
			raw.WriteString(k)
			raw.WriteString(kc)
			raw.WriteString(data.String(k))
			if i < len(kvs) {
				raw.WriteString(kvConnect)
			}
		}
		data.Set("_raw", raw.String())
		var sign string
		switch encrypt {
		case "md5":
			u.Add("sign", md5.Encrypt(data.String("sign")))
		case "base64":
			u.Add("sign", base64.URLEncode(data.String("sign")))
		case "rsa/sha1":
			if !data.Has("_key") {
				return u, fmt.Errorf("rsa私钥不能为空")
			}
			if sign, err = rsa.Sign(data.String("sign"), data.String("_key"), "sha1"); err != nil {
				return u, err
			}
			u.Add("sign", sign)

		case "rsa/md5":
			if !data.Has("_key") {
				return u, fmt.Errorf("rsa私钥不能为空")
			}
			if sign, err = rsa.Sign(data.String("sign"), data.String("_key"), "md5"); err != nil {
				return u, err
			}
			u.Add("sign", sign)
		case "aes":
			if !data.Has("_key") {
				return u, fmt.Errorf("aes密钥不能为空")
			}
			if sign, err = aes.Encrypt(data.String("sign"), data.String("_key")); err != nil {
				return u, err
			}
			u.Add("sign", sign)
		case "des":
			if !data.Has("_key") {
				return u, fmt.Errorf("des密钥不能为空")
			}
			if sign, err = des.Encrypt(data.String("sign"), data.String("_key")); err != nil {
				return u, err
			}
			u.Add("sign", sign)
		}
	}
	return u, nil
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
