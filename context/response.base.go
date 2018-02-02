package context

import (
	"strings"

	"github.com/qxnw/lib4go/types"
)

//Response 响应
type Response interface {
	GetContent(...error) interface{}
	GetStatus(...error) int
	GetParams() map[string]interface{}
	IsRedirect() (string, bool)
	GetContentType() int
	GetHeaders() map[string]string
	SetError(status int, err error)
	Close()
}

type baseResponse struct {
	Status int
	Params map[string]interface{}
}

//IsRedirect 是否是URL转跳
func (r *baseResponse) IsRedirect() (string, bool) {
	location, ok := r.Params["Location"]
	if !ok {
		return "", false
	}
	url, ok := location.(string)
	if !ok {
		return url, false
	}
	status := r.Params["Status"]
	return url, status == 301 || status == 302 || status == 303 || status == 307 || status == 309
}
func (r *baseResponse) GetStatus(err ...error) int {
	if len(err) > 0 {
		return types.DecodeInt(r.Status, 0, 500, r.Status)
	}
	return types.DecodeInt(r.Status, 0, 200, r.Status)
}

func (r *baseResponse) GetParams() map[string]interface{} {
	return r.Params
}

func (r *baseResponse) SetParams(key string, v interface{}) {
	r.Params[key] = v
}
func (r *baseResponse) SetStatus(status int) {
	r.Status = types.DecodeInt(status, 0, 200, status)
}
func (r *baseResponse) SetJWTBody(data interface{}) {
	r.Params["__jwt_"] = data
}
func (r *baseResponse) GetJWTBody() interface{} {
	return r.Params["__jwt_"]
}

func (r *baseResponse) SetHeader(name string, value string) {
	r.Params[name] = value
}
func (r *baseResponse) JsonContentType() {
	r.Params["Content-Type"] = "application/json; charset=UTF-8"
}
func (r *baseResponse) XMLContentType() {
	r.Params["Content-Type"] = "text/html; charset=UTF-8"
}
func (r *baseResponse) PlainContentType() {
	r.Params["Content-Type"] = "text/plain; charset=UTF-8"
}

func (r *baseResponse) GetContentType() int {
	responseType := 0
	if tp, ok := r.Params["Content-Type"].(string); ok {
		if strings.Contains(tp, "json") {
			responseType = 1
		} else if strings.Contains(tp, "xml") {
			responseType = 2
		} else if strings.Contains(tp, "plain") {
			responseType = 3
		}
	}
	return responseType
}
func (r *baseResponse) GetHeaders() map[string]string {
	header := make(map[string]string)
	for k, v := range r.Params {
		if !strings.HasPrefix(k, "__") && v != nil && k != "Status" {
			switch v.(type) {
			case []string:
				list := v.([]string)
				for _, i := range list {
					header[k] = i
				}
			default:
				header[k] = types.GetString(v)
			}
		}
	}
	return header
}
