package cache

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

//Delete 删除缓存
func Delete(c component.IContainer) component.ServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response context.Response, err error) {
		response = context.GetStandardResponse()
		key, err := getInputKey(ctx)
		if err != nil {
			return
		}
		cache, err := c.GetCache("cache")
		if err != nil {
			return
		}
		err = cache.Delete(key)
		if err != nil {
			return
		}
		response.SetContent(200, "success")
		return
	}
}