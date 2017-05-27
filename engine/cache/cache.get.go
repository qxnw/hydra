package cache

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/transform"
)

func (s *cacheProxy) getGetParams(ctx *context.Context) (key string, err error) {
	if ctx.Input.Input == nil || ctx.Input.Args == nil || ctx.Input.Params == nil {
		err = fmt.Errorf("engine:cache.input,params,args不能为空:%v", ctx.Input)
		return
	}
	input := ctx.Input.Input.(transform.ITransformGetter)
	key, err = input.Get("key")
	if err == nil {
		return
	}
	if ctx.Input.Body != nil && err != nil {
		inputMap := make(map[string]interface{})
		inputMap, err = jsons.Unmarshal([]byte(ctx.Input.Body.(string)))
		if err != nil {
			err = fmt.Errorf("engine:cache.body不是有效的json数据，(err:%v)", err)
			return
		}
		msm, ok := inputMap["key"]
		if !ok {
			err = errors.New("engine:cache.body的内容中未包含key标签")
			return
		}

		if key, ok = msm.(string); !ok {
			err = fmt.Errorf("engine:cache.body的内容中key标签必须为字符串:(err:%v)", msm)
			return
		}
		return
	}
	err = errors.New("engine:cache.form中未包含key标签")
	return

}

func (s *cacheProxy) get(ctx *context.Context) (r string, err error) {
	key, err := s.getGetParams(ctx)
	if err != nil {
		return "", err
	}
	client, err := s.getMemcacheClient(ctx)
	if err != nil {
		return "", err
	}
	r, err = client.Get(key)
	return
}
