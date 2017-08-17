package cache

import (
	"fmt"

	"github.com/qxnw/hydra/context"
)

func (s *cacheProxy) getSaveParams(ctx *context.Context) (key string, value string, expiresAt int, err error) {
	if err = ctx.Input.Has("key", "value", "expiresAt"); err == nil {
		key = ctx.Input.GetString("key")
		value = ctx.Input.GetString("value")
		expiresAt, err = ctx.Input.GetInt("expiresAt")
		return
	}
	conf, err := context.NewConf(ctx.Input.Body)
	if err != nil {
		return
	}
	if err = conf.Has("key", "value", "expiresAt"); err != nil {
		err = fmt.Errorf("body中:%v", err)
		return
	}
	key, _ = conf.GetString("key")
	value, _ = conf.GetString("value")
	expiresAt, err = conf.GetInt("expiresAt")
	return
}

func (s *cacheProxy) save(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
	response = context.GetStandardResponse()
	key, value, expiresAt, err := s.getSaveParams(ctx)
	if err != nil {
		return
	}
	err = ctx.Cache.Set(key, value, expiresAt)
	if err != nil {
		err = fmt.Errorf("cache.set错误(err:%v)", err)
		return
	}
	response.Success()
	return
}
