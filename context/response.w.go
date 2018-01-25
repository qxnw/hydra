package context

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/qxnw/lib4go/types"
)

var webResponsePool *sync.Pool

func init() {
	webResponsePool = &sync.Pool{
		New: func() interface{} {
			return &WebResponse{baseResponse: &baseResponse{Params: make(map[string]interface{})}}
		},
	}
}

type WebResponse struct {
	Content interface{}
	ctx     *Context
	*baseResponse
}

func GetWebResponse(ctx *Context) *WebResponse {
	response := webResponsePool.Get().(*WebResponse)
	response.ctx = ctx
	return response
}

//Redirect 设置页面转跳
func (r *WebResponse) Redirect(code int, url string) *WebResponse {
	r.Params["Status"] = code
	r.Params["Location"] = url
	r.Status = code
	return r
}
func (r *WebResponse) GetContent(errs ...error) interface{} {
	if r.Content != nil {
		return r.Content
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return r.Content
}
func (r *WebResponse) Success(v ...interface{}) *WebResponse {
	r.Status = 200
	if len(v) > 0 {
		r.Content = v[0]
		return r
	}
	return r
}

//SetView 设置view
func (r *WebResponse) SetView(name string) {
	r.Params["__view"] = name
}

//NoView 设置view
func (r *WebResponse) NoView() {
	r.Params["__view"] = "NONE"
}

//SetCookie 设置cookie
func (c *WebResponse) SetCookie(name string, value string, timeout int, domain string) {
	list := make([]string, 0, 2)
	if v, ok := c.Params["Set-Cookie"]; ok {
		list = v.([]string)
	}
	list = append(list, c.getSetCookie(name, value, timeout, domain))
	c.Params["Set-Cookie"] = list
}
func (c *WebResponse) getSetCookie(name string, value string, timeout interface{}, domain string) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s=%s", name, value)
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
		if len(domain) > 0 {
			fmt.Fprintf(&b, "; Expires=%s; Max-Age=%d;path=/;domain=%s", time.Now().Add(time.Duration(maxAge)*time.Second).UTC().Format(time.RFC1123), maxAge, domain)
			return b.String()
		}
		fmt.Fprintf(&b, "; Expires=%s; Max-Age=%d;path=/", time.Now().Add(time.Duration(maxAge)*time.Second).UTC().Format(time.RFC1123), maxAge)

	case maxAge < 0:
		fmt.Fprintf(&b, "; Max-Age=0")
	}
	return b.String()
}

func (r *WebResponse) SetError(status int, err error) {
	if err != nil {
		r.Status = types.DecodeInt(status, 0, 500, status)
		r.Content = err
		return
	}
	r.Status = types.DecodeInt(status, 0, 200, status)
}
func (r *WebResponse) Set(s int, rr interface{}, p map[string]string, err error) error {
	r.Status = s
	if r.Status == 0 {
		r.Status = types.DecodeInt(err, nil, 500, 200)
	}
	for k, v := range p {
		r.Params[k] = v
	}
	r.Content = rr
	return err
}
func (r *WebResponse) Close() {
	r.Content = nil
	r.Params = make(map[string]interface{})
	webResponsePool.Put(r)
}
