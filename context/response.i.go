package context

import "github.com/qxnw/lib4go/types"

type ObjectReponse struct {
	Content interface{}
	*baseResponse
}

func GetObjectReponse() *ObjectReponse {
	return &ObjectReponse{baseResponse: &baseResponse{Params: make(map[string]interface{})}}
}

func (r *ObjectReponse) GetContent(errs ...error) interface{} {
	if r.Content != nil {
		return r.Content
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return r.Content
}

func (r *ObjectReponse) Success(v ...interface{}) *ObjectReponse {
	r.Status = 200
	if len(v) > 0 {
		r.Content = v[0]
		return r
	}
	return r
}
func (r *ObjectReponse) SetContent(status int, content interface{}) *ObjectReponse {
	r.Status = types.DecodeInt(status, 0, 200, status)
	r.Content = content
	return r
}

func (r *ObjectReponse) Set(s int, rr interface{}, p map[string]string, err error) error {
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
