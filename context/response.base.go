package context

import (
	"strings"

	"github.com/qxnw/lib4go/types"
)

//Response 响应
type Response interface {
	GetContent() interface{}
	GetStatus(error) int
	GetParams() map[string]interface{}
	IsRedirect() bool
	GetContentType() int
	GetHeaders() map[string]string
}

type baseResponse struct {
	Status int
	Params map[string]interface{}
}

//IsRedirect 是否是URL转跳
func (r *baseResponse) IsRedirect() bool {
	location, ok := r.Params["Location"]
	if !ok {
		return false
	}
	_, ok = location.(string)
	if !ok {
		return false
	}
	return r.Status == 301 || r.Status == 302 || r.Status == 303 || r.Status == 307 || r.Status == 309
}
func (r *baseResponse) GetStatus(err error) int {
	if err != nil {
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

func (r *baseResponse) SetError(status int, err error) {
	if err != nil {
		r.Status = types.DecodeInt(status, 0, 500, status)
		return
	}
	r.Status = types.DecodeInt(status, 0, 200, status)
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
	responseType := 1
	if tp, ok := r.Params["Content-Type"].(string); ok {
		if strings.Contains(tp, "xml") {
			responseType = 2
		} else if strings.Contains(tp, "plain") {
			responseType = 0
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
