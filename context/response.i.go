package context

import (
	"sync"

	"github.com/qxnw/lib4go/types"
)

var objectResponsePool *sync.Pool

func init() {
	objectResponsePool = &sync.Pool{
		New: func() interface{} {
			return &ObjectResponse{baseResponse: &baseResponse{Params: make(map[string]interface{})}}
		},
	}
}

type ObjectResponse struct {
	Content interface{}
	*baseResponse
}

func GetObjectResponse() *ObjectResponse {
	return objectResponsePool.Get().(*ObjectResponse)
}

func (r *ObjectResponse) GetContent(errs ...error) interface{} {
	if r.Content != nil {
		return r.Content
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return r.Content
}

func (r *ObjectResponse) Success(v ...interface{}) *ObjectResponse {
	r.Status = 200
	if len(v) > 0 {
		r.Content = v[0]
		return r
	}
	return r
}
func (r *ObjectResponse) SetContent(status int, content interface{}) *ObjectResponse {
	r.Status = types.DecodeInt(status, 0, 200, status)
	r.Content = content
	return r
}
func (r *ObjectResponse) Close() {
	r.Content = nil
	r.Params = make(map[string]interface{})
	objectResponsePool.Put(r)
}
func (r *ObjectResponse) SetError(status int, err error) {
	if err != nil {
		r.Status = types.DecodeInt(status, 0, 500, status)
		r.Content = err
		return
	}
	r.Status = types.DecodeInt(status, 0, 200, status)
}
