package cache

import (
	"fmt"

	"github.com/qxnw/hydra/context"
)

func (s *cacheProxy) getKeyExpiresAt(ctx *context.Context) (key string, expiresAt int, err error) {
	if err = ctx.Input.Has("key", "expiresAt"); err == nil {
		key, _ = ctx.Input.Get(key)
		expiresAt, err = ctx.Input.GetInt("expiresAt")
		return
	}

	conf, err := context.NewConf(ctx.Input.Body)
	if err != nil {
		err = fmt.Errorf("body不是有效的json数据:%v", err)
		return
	}
	if key, err = conf.GetString("key"); err != nil {
		err = fmt.Errorf("body的内容中:%s", err)
		return
	}
	if expiresAt, err = conf.GetInt("expiresAt"); err != nil {
		err = fmt.Errorf("body的内容中:(err:%v)", err)
		return
	}
	return key, expiresAt, nil
}

func (s *cacheProxy) delay(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
	response = context.GetResponse()
	key, expiresAt, err := s.getKeyExpiresAt(ctx)
	if err != nil {
		return
	}
	err = ctx.Cache.Delay(key, expiresAt)
	if err != nil {
		err = fmt.Errorf("delay错误(err:%v)", err)
		return
	}
	response.Success()
	return
}
