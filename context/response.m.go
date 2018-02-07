package context

import (
	"fmt"
	"sync"

	"github.com/qxnw/lib4go/types"
)

var mapResponsePool *sync.Pool

func init() {
	mapResponsePool = &sync.Pool{
		New: func() interface{} {
			r := &MapResponse{
				baseResponse: &baseResponse{Params: make(map[string]interface{})}, Content: make(map[string]interface{})}
			return r
		},
	}
}

type MapResponse struct {
	Content map[string]interface{}
	err     error
	*baseResponse
}

func GetMapResponse() *MapResponse {
	return mapResponsePool.Get().(*MapResponse)
}

func (r *MapResponse) GetContent() interface{} {
	return r.Content
}

func (r *MapResponse) Success(v map[string]interface{}) {
	r.Status = 200
	r.Content = v
	return

}
func (r *MapResponse) SetContent(status int, content interface{}) {
	if status == 0 {
		status = r.Status
	}
	switch content.(type) {
	case HydraError:
		r.Status = types.DecodeInt(status, 0, 500, status)
		r.err = content.(HydraError).error
	case error:
		r.Status = types.DecodeInt(status, 0, 500, status)
		r.err = content.(error)
	case map[string]interface{}:
		r.Status = types.DecodeInt(status, 0, 200, status)
		r.Content = content.(map[string]interface{})
	default:
		if content == nil {
			r.Status = types.DecodeInt(status, 0, 200, status)
			return
		}
		r.Status = 500
		r.err = fmt.Errorf("MapResponse.content输入类型错误,必须为:map[string]interface{}")
	}
	return
}

func (r *MapResponse) GetError() error {
	return r.err
}

func (r *MapResponse) Close() {
	r.Content = make(map[string]interface{})
	r.Params = make(map[string]interface{})
	mapResponsePool.Put(r)
}
