package context

import (
	"errors"
	"net/http"
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
func (c *ContextHTTP) GetHostName() (string, error) {
	request, err := c.GetHTTPRequest()
	if err != nil {
		return "", err
	}
	return request.Host, nil
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
