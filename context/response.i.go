package context

import (
	"sync"

	"github.com/qxnw/lib4go/types"
)

var objectResponsePool *sync.Pool

func init() {
	objectResponsePool = &sync.Pool{
		New: func() interface{} {
			r := &ObjectResponse{baseResponse: &baseResponse{Params: make(map[string]interface{})}}
			return r
		},
	}
}

type ObjectResponse struct {
	Content interface{}
	err     error
	*baseResponse
}

func GetObjectResponse() *ObjectResponse {
	return objectResponsePool.Get().(*ObjectResponse)
}

func (r *ObjectResponse) GetContent() interface{} {
	return r.Content
}

func (r *ObjectResponse) Success(v interface{}) {
	r.Status = 200
	r.Content = v
	return

}
func (r *ObjectResponse) SetContent(status int, content interface{}) {
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
	default:
		r.Status = types.DecodeInt(status, 0, 200, status)
	}
	r.Status = types.DecodeInt(status, 0, 200, status)
	r.Content = content
	return
}

func (r *ObjectResponse) GetError() error {
	return r.err
}
func (r *ObjectResponse) Close() {
	r.Content = nil
	r.Params = make(map[string]interface{})
	objectResponsePool.Put(r)
}
