package context

import (
	"sync"

	"github.com/qxnw/lib4go/types"
)

var standradResponsePool *sync.Pool

func init() {
	standradResponsePool = &sync.Pool{
		New: func() interface{} {
			return &StandardResponse{baseResponse: &baseResponse{Params: make(map[string]interface{})}}
		},
	}
}

type StandardResponse struct {
	Content string
	*baseResponse
}

func GetStandardResponse() *StandardResponse {
	return standradResponsePool.Get().(*StandardResponse)
}

func (r *StandardResponse) GetContent(errs ...error) interface{} {
	if r.Content != "" {
		return r.Content
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return r.Content
}

func (r *StandardResponse) Success(v ...string) *StandardResponse {
	r.Status = 200
	if len(v) > 0 {
		r.Content = v[0]
		return r
	}
	r.Content = "SUCCESS"
	return r
}
func (r *StandardResponse) SetContent(status int, content string) *StandardResponse {
	r.Status = types.DecodeInt(status, 0, 200, status)
	r.Content = content
	return r
}
func (r *StandardResponse) SetError(status int, err error) {
	if err != nil {
		r.Status = types.DecodeInt(status, 0, 500, status)
		r.Content = err.Error()
		return
	}
	r.Status = types.DecodeInt(status, 0, 200, status)
}
func (r *StandardResponse) Set(s int, rr string, p map[string]string, err error) error {
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
func (r *StandardResponse) Close() {
	r.Content = ""
	r.Params = make(map[string]interface{})
	standradResponsePool.Put(r)
}
