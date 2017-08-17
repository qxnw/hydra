package context

type WebResponse struct {
	Content interface{}
	*baseResponse
}

func GetWebResponse() *WebResponse {
	return &WebResponse{baseResponse: &baseResponse{Params: make(map[string]interface{})}}
}

//Redirect 设置页面转跳
func (r *WebResponse) Redirect(code int, url string) *WebResponse {
	r.Params["Status"] = code
	r.Params["Location"] = url
	return r
}
func (r *WebResponse) GetContent(errs ...error) interface{} {
	if r.Content != nil {
		return r.Content
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return r.Content
}
func (r *WebResponse) Success(v ...string) *WebResponse {
	r.Status = 200
	if len(v) > 0 {
		r.Content = v[0]
		return r
	}
	r.Content = "SUCCESS"
	return r
}
