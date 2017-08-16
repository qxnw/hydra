package context

type WebReponse struct {
	Content interface{}
	*baseResponse
}

func GetWebResponse() *WebReponse {
	return &WebReponse{baseResponse: &baseResponse{Params: make(map[string]interface{})}}
}

//Redirect 设置页面转跳
func (r *WebReponse) Redirect(code int, url string) *WebReponse {
	r.Params["Status"] = code
	r.Params["Location"] = url
	return r
}
func (r *WebReponse) GetContent(errs ...error) interface{} {
	if r.Content != nil {
		return r.Content
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return r.Content
}
func (r *WebReponse) Success(v ...string) *WebReponse {
	r.Status = 200
	if len(v) > 0 {
		r.Content = v[0]
		return r
	}
	r.Content = "SUCCESS"
	return r
}
