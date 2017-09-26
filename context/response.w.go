package context

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/qxnw/lib4go/encoding/base64"
	"github.com/qxnw/lib4go/security/xsrf"
	"github.com/qxnw/lib4go/types"
	"github.com/qxnw/lib4go/utility"
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

//AddAuthTimes 累加授权次数
func (r *WebResponse) AddAuthTimes(names ...string) (err error) {
	name := "__auth_times"
	if len(names) > 0 {
		name = names[0]
	}
	ckv := r.ctx.HTTP.GetCookieString(name)
	times := types.ToInt(ckv, 0) + 1
	host, err := r.ctx.HTTP.GetHostName()
	if err != nil {
		return
	}
	r.SetCookie(name, strconv.Itoa(times), 0, host)
	if times > 3 {
		err = fmt.Errorf("授权次数超过限制次数3次:%s", r.GetSession())
		return
	}
	return
}

//ClearAuthTimes ...
func (r *WebResponse) ClearAuthTimes(names ...string) error {
	name := "__auth_times"
	if len(names) > 0 {
		name = names[0]
	}
	host, err := r.ctx.HTTP.GetHostName()
	if err != nil {
		return err
	}
	r.SetCookie(name, "0", 0, host)
	return nil
}

//RedirectToWXAuth 跳转微信授权
func (r *WebResponse) RedirectToWXAuth(wxAuthURL, appid, notifyURL string) error {
	//记录当前地址
	request, err := r.ctx.HTTP.GetHTTPRequest()
	if err != nil {
		panic(err)
	}
	host, err := r.ctx.HTTP.GetHostName()
	if err != nil {
		return err
	}
	r.SetCookie("_pre_url", request.Proto+request.RequestURI, 0, host)
	url, err := r.GetAuthURL(wxAuthURL, appid, notifyURL)
	if err != nil {
		return err
	}
	r.Redirect(302, url)
	return nil
}

//GetAuthURL 获取用户授权地址
func (r *WebResponse) GetAuthURL(wxAuthURL, appid, notifyURL string) (string, error) {
	host, err := r.ctx.HTTP.GetHostName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s?appid=%s&redirect_uri=http://%s%s&response_type=code&scope=snsapi_base&state=#wechat:redirect",
		wxAuthURL,
		appid,
		host,
		notifyURL), nil
}

//MakeXSRFToken
func (c *WebResponse) MakeXSRFToken(xsrfSecret string) string {
	t, _ := base64.URLDecode(utility.GetGUID())
	token := xsrf.CreateXSRFToken(xsrfSecret, t)
	return token
}
func (r *WebResponse) AuthorizeLogin(appid string, code string, authLoginName string) (sid string, err error) {
	status, result, _, err := r.ctx.RPC.Request(authLoginName, map[string]string{
		"app_id": appid,
		"state":  utility.GetGUID(),
		"code":   code,
	}, true)
	if err != nil || status != 200 {
		return "", fmt.Errorf("处理用户授权信息失败:%s,%s,%d:%v", authLoginName, code, status, err)
	}

	conf, err := NewConf(result)
	if err != nil {
		return "", fmt.Errorf("用户信息转换为json异常,%s,err:%v", result, err)
	}
	if err = conf.Has("guid"); err != nil {
		return "", fmt.Errorf("返回的数据中没有guid(session):%s", result)
	}
	return conf.GetString("guid")
}

//FetchUserSession 用微信code换取session
func (r *WebResponse) Login(appid string, authLoginName string) (err error) {
	if err = r.ctx.Input.CheckInput("code"); err != nil {
		err = fmt.Errorf("微信授权返回code为空")
		return
	}
	sid, err := r.AuthorizeLogin(appid, r.ctx.Input.GetString("code"), authLoginName)
	if err != nil || sid == "" {
		err = fmt.Errorf("获取用户session错误:sid:%s,err:%v", sid, err)
		return
	}
	cURL := r.ctx.HTTP.GetCookieString("_pre_url")
	host, err := r.ctx.HTTP.GetHostName()
	if err != nil {
		return err
	}
	r.SetCookie("_pre_url", "", 0, host)
	err = r.SetSession(sid)
	if err != nil {
		return
	}
	r.Redirect(302, cURL)
	return
}

//SetSessionID 设置session
func (r *WebResponse) SetSession(sid string) error {
	host, err := r.ctx.HTTP.GetHostName()
	if err != nil {
		return err
	}
	r.SetCookie("sid", sid, 3600*24*15, host)
	return nil
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

func (c *WebResponse) IsLogin() (err error) {
	sid := c.GetSession()
	if sid == "" {
		err = errors.New("sid为空，判定为session过期或新用户，需要跳转授权")
		return
	}
	return
}

//GetSessionID 获取当session
func (c *WebResponse) GetSession() string {
	return c.ctx.HTTP.GetCookieString("sid")
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
