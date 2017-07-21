package context

//Response 响应
type Response struct {
	Content interface{}
	Status  int
	Params  map[string]interface{}
}

//IsRedirect 是否是URL转跳
func (r *Response) IsRedirect() bool {
	return r.Status == 301 || r.Status == 302 || r.Status == 303 || r.Status == 307
}
