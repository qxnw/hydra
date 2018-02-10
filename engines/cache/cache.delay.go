package cache

import (
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

func getKeyExpiresAt(ctx *context.Context) (key string, expiresAt int, err error) {
	if err = ctx.Request.Form.Check("key", "expiresAt"); err != nil {
		return
	}
	key = ctx.Request.Form.GetString(key)
	expiresAt = ctx.Request.Form.GetInt("expiresAt")
	body, err := ctx.Request.Ext.GetBody()
	if err != nil {
		return
	}
	conf, err := context.NewConf(body)
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

//Delay 延长缓存时间
func Delay(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		key, expiresAt, err := getKeyExpiresAt(ctx)
		if err != nil {
			return
		}
		cache, err := c.GetCache("cache")
		if err != nil {
			return
		}
		err = cache.Delay(key, expiresAt)
		if err != nil {
			err = fmt.Errorf("delay错误(err:%v)", err)
			return
		}
		response.SetContent(200, "success")
		return
	}
}
