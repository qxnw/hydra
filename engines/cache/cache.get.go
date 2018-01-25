package cache

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

func getInputKey(ctx *context.Context) (key string, err error) {
	if key, err = ctx.Request.Form.Get("key"); err != nil {
		return
	}
	body, err := ctx.Request.Ext.GetBody()
	if err != nil {
		err = errors.New("输入参数中未包含key")
		return
	}
	conf, err := context.NewConf(body)
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

//Get 获取缓存值
func Get(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		key, err := getInputKey(ctx)
		if err != nil {
			response.SetStatus(406)
			return
		}
		cache, err := c.GetCache("cache")
		if err != nil {
			return
		}
		r, err := cache.Get(key)
		if err != nil {
			response.SetStatus(410)
			return
		}
		response.Success(r)
		return
	}
}
