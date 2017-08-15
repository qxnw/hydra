package cache

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/context"
)

func (s *cacheProxy) getInputKey(ctx *context.Context) (key string, err error) {
	if err = ctx.Input.Has("key"); err == nil {
		key, _ = ctx.Input.Get("key")
		return
	}
	if ctx.Input.Body == "" {
		err = errors.New("输入参数中未包含key")
		return
	}
	conf, err := context.NewConf(ctx.Input.Body)
	if err != nil {
		err = fmt.Errorf("body不是有效的json数据:%v", err)
		return
	}
	if key, err = conf.GetString("key"); err != nil {
		err = fmt.Errorf("body的内容中:%v", err)
		return
	}
	return
}

func (s *cacheProxy) get(name string, mode string, service string, ctx *context.Context) (response *context.StandardReponse, err error) {
	response = context.GetStandardResponse()
	key, err := s.getInputKey(ctx)
	if err != nil {
		response.SetStatus(406)
		return
	}
	r, err := ctx.Cache.Get(key)
	if err != nil {
		response.SetStatus(410)
		return
	}
	response.Success(r)
	return
}
