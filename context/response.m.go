package context

import (
	"sync"

	"github.com/qxnw/lib4go/types"
)

var mapResponsePool *sync.Pool

func init() {
	mapResponsePool = &sync.Pool{
		New: func() interface{} {
			r := &MapResponse{
				baseResponse: &baseResponse{Params: make(map[string]interface{})}, Content: make(map[string]interface{})}
			r.Params["__view"] = "NONE"
			return r
		},
	}
}

type MapResponse struct {
	Content map[string]interface{}
	*baseResponse
}

func GetMapResponse() *MapResponse {
	return mapResponsePool.Get().(*MapResponse)
}
func (r *MapResponse) GetContent(errs ...error) interface{} {
	if len(r.Content) > 0 {
		return r.Content
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return r.Content
}
func (r *MapResponse) SetError(status int, err error) {
	if err != nil {
		r.Status = types.DecodeInt(status, 0, 500, status)
		return
	}
	r.Status = types.DecodeInt(status, 0, 200, status)
}
func (r *MapResponse) Success(v ...map[string]interface{}) *MapResponse {
	r.Status = 200
	if len(v) > 0 {
		r.Content = v[0]
		return r
	}
	return r
}

func (r *MapResponse) SetContent(status int, content map[string]interface{}) *MapResponse {
	r.Status = types.DecodeInt(status, 0, 200, status)
	r.Content = content
	return r
}

func (r *MapResponse) Set(s int, rr map[string]interface{}, p map[string]string, err error) error {
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
func (r *MapResponse) Close() {
	r.Content = nil
	r.Params = make(map[string]interface{})
	r.Params["__view"] = "NONE"
	mapResponsePool.Put(r)
}
