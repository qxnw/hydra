package cache

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/transform"
)

func (s *cacheProxy) getDelayParams(ctx *context.Context) (key string, expiresAt int, err error) {
	if ctx.Input.Input == nil || ctx.Input.Args == nil || ctx.Input.Params == nil {
		err = fmt.Errorf("input,params,args不能为空:%v", ctx.Input)
		return
	}
	input := ctx.Input.Input.(transform.ITransformGetter)
	key, err = input.Get("key")
	if ctx.Input.Body != nil && err != nil {
		inputMap := make(map[string]interface{})
		inputMap, err = jsons.Unmarshal([]byte(ctx.Input.Body.(string)))
		if err != nil {
			err = fmt.Errorf("body不是有效的json数据，(err:%v)", err)
			return
		}
		msm, ok := inputMap["key"]
		if !ok {
			err = errors.New("body的内容中未包含key标签")
			return
		}

		if key, ok = msm.(string); !ok {
			err = fmt.Errorf("body的内容中key标签必须为字符串:(err:%v)", msm)
			return
		}
		expires, ok := inputMap["expiresAt"]
		if !ok {
			expires = "0"
		}
		expiresAt, err = strconv.Atoi(expires.(string))
		if err != nil {
			err = fmt.Errorf("body的内容中expiresAt标签不是有效的数字:(err:%v)", err)
			return "", 0, err
		}
		return key, expiresAt, nil
	}
	key, err = input.Get("key")
	if err != nil {
		err = errors.New("form中未包含key标签")
		return
	}
	expires, err := input.Get("expiresAt")
	if err != nil {
		err = errors.New("form中未包含expiresAt标签")
		return
	}
	expiresAt, err = strconv.Atoi(expires)
	if err != nil {
		err = fmt.Errorf("form的内容中expiresAt标签不是有效的数字:(err:%v)", err)
		return "", 0, err
	}
	return
}

func (s *cacheProxy) delay(ctx *context.Context) (r string, err error) {
	key, expiresAt, err := s.getDelayParams(ctx)
	if err != nil {
		return "", err
	}
	client, err := s.getMemcacheClient(ctx)
	if err != nil {
		return "", err
	}
	err = client.Delay(key, expiresAt)
	if err != nil {
		err = fmt.Errorf("delay错误(err:%v)", err)
	}
	r = "SUCCESS"

	return
}
