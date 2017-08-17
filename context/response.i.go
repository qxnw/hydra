package context

import "github.com/qxnw/lib4go/types"

type ObjectResponse struct {
	Content interface{}
	*baseResponse
}

func GetObjectResponse() *ObjectResponse {
	return &ObjectResponse{baseResponse: &baseResponse{Params: make(map[string]interface{})}}
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

func (r *ObjectResponse) Set(s int, rr interface{}, p map[string]string, err error) error {
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
