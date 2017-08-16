package context

import "github.com/qxnw/lib4go/types"

type StandardReponse struct {
	Content string
	*baseResponse
}

func GetStandardResponse() *StandardReponse {
	return &StandardReponse{baseResponse: &baseResponse{Params: make(map[string]interface{})}}
}

func (r *StandardReponse) GetContent(errs ...error) interface{} {
	if r.Content != "" {
		return r.Content
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return r.Content
}

func (r *StandardReponse) Success(v ...string) *StandardReponse {
	r.Status = 200
	if len(v) > 0 {
		r.Content = v[0]
		return r
	}
	r.Content = "SUCCESS"
	return r
}
func (r *StandardReponse) SetContent(status int, content string) *StandardReponse {
	r.Status = types.DecodeInt(status, 0, 200, status)
	r.Content = content
	return r
}

func (r *StandardReponse) Set(s int, rr string, p map[string]string, err error) error {
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
