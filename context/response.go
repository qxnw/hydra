package context

import "github.com/qxnw/lib4go/types"

//Response 响应
type Response struct {
	Content interface{}
	Status  int
	Params  map[string]interface{}
}

func GetResponse() *Response {
	return &Response{Params: make(map[string]interface{})}
}

//IsRedirect 是否是URL转跳
func (r *Response) IsRedirect() bool {
	return r.Status == 301 || r.Status == 302 || r.Status == 303 || r.Status == 307
}
func (r *Response) Success(v ...interface{}) *Response {
	r.Status = 200
	if len(v) > 0 {
		r.Content = v[0]
		return r
	}
	r.Content = "SUCCESS"
	return r
}

//Redirect 设置页面转跳
func (r *Response) Redirect(code int, url string) *Response {
	r.Params["Status"] = code
	r.Params["Location"] = url
	return r
}
func (r *Response) AddParams(key string, v interface{}) *Response {
	r.Params[key] = v
	return r
}
func (r *Response) SetContent(status int, content interface{}) *Response {
	r.Status = types.DecodeInt(status, 0, 200, status)
	r.Content = content
	return r
}
func (r *Response) SetStatus(status int) *Response {
	r.Status = types.DecodeInt(status, 0, 200, status)
	return r
}
func (r *Response) SetError(status int, err error) *Response {
	if err != nil {
		r.Status = types.DecodeInt(status, 0, 500, status)
		return r
	}
	r.Status = types.DecodeInt(status, 0, 200, status)
	return r
}
func (r *Response) Set(s int, rr interface{}, p map[string]string, err error) error {
	r.Status = s
	r.Content = rr
	for k, v := range p {
		r.Params[k] = v
	}
	return err
}
func (r *Response) SetHeader(name string, value string) *Response {
	r.Params[name] = value
	return r
}
