package context

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type ContextHTTP struct {
	ctx *Context
}

//Reset 重置context
func (c *ContextHTTP) Reset(ctx *Context) {
	c.ctx = ctx
}

//GetHTTPRequest 获和取http.request对象
func (c *ContextHTTP) GetHTTPRequest() (request *http.Request, err error) {
	r := c.ctx.Input.Ext["__func_http_request_"]
	if r == nil {
		return nil, errors.New("未找到__func_http_request_")
	}
	if f, ok := r.(*http.Request); ok {
		return f, nil
	}
	return nil, errors.New("未找到__func_http_request_传入类型错误")
}

//Redirect 设置页面转跳
func (c *ContextHTTP) Redirect(code int, url string, param map[string]interface{}) {
	param["Status"] = code
	param["Location"] = url
}

func (c *ContextHTTP) getSetCookie(name string, value string, timeout interface{}, domain string) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s=%s", name, value)

	//fix cookie not work in IE

	var maxAge int64
	switch v := timeout.(type) {
	case int:
		maxAge = int64(v)
	case int32:
		maxAge = int64(v)
	case int64:
		maxAge = v
	}

	switch {
	case maxAge > 0:
		fmt.Fprintf(&b, "; Expires=%s; Max-Age=%d;path=/;domain=%s", time.Now().Add(time.Duration(maxAge)*time.Second).UTC().Format(time.RFC1123), maxAge, domain)
	case maxAge < 0:
		fmt.Fprintf(&b, "; Max-Age=0")
	}
	return b.String()
}

//SetCookie 设置cookie
func (c *ContextHTTP) SetCookie(name string, value string, timeout int, domain string, param map[string]interface{}) {
	if param["Set-Cookie"] == nil {
		param["Set-Cookie"] = make([]string, 0, 2)
	}
	list := param["Set-Cookie"].([]string)
	list = append(list, c.getSetCookie(name, value, timeout, domain))
	param["Set-Cookie"] = list
}

//GetCookieString 获取cookie字符串
func (c *ContextHTTP) GetCookieString(name string) string {
	if s, err := c.GetCookie(name); err == nil {
		return s
	}
	return ""
}

//GetCookie 从http.request中获取cookie
func (c *ContextHTTP) GetCookie(name string) (string, error) {
	request, err := c.GetHTTPRequest()
	if err != nil {
		return "", err
	}
	cookie, err := request.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
