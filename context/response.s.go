package context

import (
	"fmt"
	"sync"

	"github.com/qxnw/lib4go/types"
)

var standradResponsePool *sync.Pool

func init() {
	standradResponsePool = &sync.Pool{
		New: func() interface{} {
			r := &StandardResponse{baseResponse: &baseResponse{Params: make(map[string]interface{})}}
			return r
		},
	}
}

type StandardResponse struct {
	Content string
	err     error
	*baseResponse
}

func GetStandardResponse() *StandardResponse {
	return standradResponsePool.Get().(*StandardResponse)
}

func (r *StandardResponse) GetContent() interface{} {
	return r.Content
}

func (r *StandardResponse) Success(v string) {
	r.Status = 200
	r.Content = v
	return

}
func (r *StandardResponse) SetContent(status int, content interface{}) {
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
	case string:
		r.Status = types.DecodeInt(status, 0, 200, status)
		r.Content = content.(string)
	default:
		if content == nil {
			r.Status = types.DecodeInt(status, 0, 200, status)
			return
		}
		r.Status = 500
		r.err = fmt.Errorf("StandardResponse.content输入类型错误,必须为:string")
	}
	return
}

func (r *StandardResponse) GetError() error {
	return r.err
}

func (r *StandardResponse) Close() {
	r.Content = ""
	r.Params = make(map[string]interface{})
	standradResponsePool.Put(r)
}
