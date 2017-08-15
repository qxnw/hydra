package context

import "github.com/qxnw/lib4go/types"

type MapResponse struct {
	Content map[string]interface{}
	*baseResponse
}

func GetMapResponse() *MapResponse {
	return &MapResponse{baseResponse: &baseResponse{Params: make(map[string]interface{})}}
}
func (r *MapResponse) GetContent() interface{} {
	return r.Content
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
