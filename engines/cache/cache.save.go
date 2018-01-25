package cache

import (
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

func getSaveParams(ctx *context.Context) (key string, value string, expiresAt int, err error) {
	if err = ctx.Request.Form.Check("key", "value", "expiresAt"); err != nil {
		return
	}
	key = ctx.Request.Form.GetString("key")
	value = ctx.Request.Form.GetString("value")
	expiresAt = ctx.Request.Form.GetInt("expiresAt")
	body, err := ctx.Request.Ext.GetBody()
	if err != nil {
		return
	}
	conf, err := context.NewConf(body)
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

//Save 保存缓存值
func Save(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		key, value, expiresAt, err := getSaveParams(ctx)
		if err != nil {
			return
		}
		cache, err := c.GetCache("cache")
		if err != nil {
			return
		}
		err = cache.Set(key, value, expiresAt)
		if err != nil {
			err = fmt.Errorf("cache.set错误(err:%v)", err)
			return
		}
		response.Success()
		return
	}
}
