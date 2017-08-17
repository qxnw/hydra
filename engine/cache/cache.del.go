package cache

import "github.com/qxnw/hydra/context"

//Handle(name string, mode string, service string, ctx *Context) (*Response, error)
func (s *cacheProxy) del(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	key, err := s.getInputKey(ctx)
	if err != nil {
		return
	}
	err = ctx.Cache.Delete(key)
	if err != nil {
		return
	}
	response.Success()
	return
}
