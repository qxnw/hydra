package context

import (
	"strings"

	"github.com/qxnw/lib4go/types"
)

//Response 响应
type Response interface {
	GetContent() interface{}
	GetStatus() int
	SetStatus(int)
	GetParams() map[string]interface{}
	SetParams(v map[string]interface{})
	SetParam(key string, v interface{})
	SetContent(status int, content interface{})
	IsRedirect() (string, bool)
	GetContentType() int
	GetHeaders() map[string]string
	SetHeader(name string, value string)
	SetHeaders(map[string]string)
	SetJWTBody(data interface{})
	GetError() error
	IsSuccess() bool
	Close()
}

type baseResponse struct {
	Status int
	Params map[string]interface{}
}

func (r *baseResponse) GetStatus() int {
	return r.Status
}

func (r *baseResponse) GetParams() map[string]interface{} {
	return r.Params
}
func (r *baseResponse) SetParams(v map[string]interface{}) {
	r.Params = v
}
func (r *baseResponse) SetParam(key string, v interface{}) {
	r.Params[key] = v
}

func (r *baseResponse) IsSuccess() bool {
	return r.Status < 500
}

func (r *baseResponse) SetStatus(status int) {
	r.Status = types.DecodeInt(status, 0, 200, status)
}

func (r *baseResponse) SetJWTBody(data interface{}) {
	r.Params["__jwt_"] = data
}

//SetJsonContentType 将content type设置为application/json; charset=UTF-8
func (r *baseResponse) SetJSONContentType() {
	r.Params["Content-Type"] = "application/json; charset=UTF-8"
}

//SetXMLContentType 将content type设置为application/xml; charset=UTF-8
func (r *baseResponse) SetXMLContentType() {
	r.Params["Content-Type"] = "text/xml; charset=UTF-8"
}

//SetPlainContentType 将content type设置为text/plain; charset=UTF-8
func (r *baseResponse) SetPlainContentType() {
	r.Params["Content-Type"] = "text/plain; charset=UTF-8"
}

//GetContentType  0：自动 1:json 2:xml 3:plain
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

//SetHeader 设置http头
func (r *baseResponse) SetHeader(name string, value string) {
	r.Params[name] = value
}
func (r *baseResponse) SetHeaders(h map[string]string) {
	for k, v := range h {
		r.Params[k] = v
	}
}

//GetHeaders 获取http头配置
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
