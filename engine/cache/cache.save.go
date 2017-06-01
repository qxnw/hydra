package cache

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/context"

	"strconv"

	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/utility"
)

func (s *cacheProxy) getSaveParams(ctx *context.Context) (key string, value string, expiresAt int, err error) {
	if ctx.Input.Input == nil || ctx.Input.Args == nil || ctx.Input.Params == nil {
		err = fmt.Errorf("input,params,args不能为空:%v", ctx.Input)
		return
	}
	input := ctx.Input.Input.(transform.ITransformGetter)
	key, err = input.Get("key")
	if err != nil && !utility.IsStringEmpty(ctx.Input.Body) {
		inputMap := make(map[string]interface{})
		inputMap, err = jsons.Unmarshal([]byte(ctx.Input.Body.(string)))
		if err != nil {
			err = fmt.Errorf("输入的body不是有效的json数据，(err:%v)", err)
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
		tgs, ok := inputMap["value"]
		if !ok {
			err = errors.New("body的内容中未包含value标签")
			return
		}
		buf, err := jsons.Marshal(tgs)
		if err != nil {
			err = fmt.Errorf("body的内容中value标签必须为不效的json对象:(err:%v)", err)
			return "", "", 0, err
		}
		expires, ok := inputMap["expiresAt"]
		if !ok {
			expires = "0"
		}
		expiresAt, err = strconv.Atoi(expires.(string))
		if err != nil {
			err = fmt.Errorf("body的内容中expiresAt标签不是有效的数字:(err:%v)", err)
			return "", "", 0, err
		}
		value = string(buf)
		return key, value, expiresAt, nil
	}
	key, err = input.Get("key")
	if err != nil {
		err = errors.New("form中未包含key标签")
		return
	}

	value, err = input.Get("value")
	if err != nil {
		err = errors.New("form中未包含value标签")
		return
	}
	expires, err := input.Get("expiresAt")
	if err != nil {
		expires = "0"
	}
	expiresAt, err = strconv.Atoi(expires)
	if err != nil {
		err = fmt.Errorf("form的内容中expiresAt标签不是有效的数字:(err:%v)", err)
		return "", "", 0, err
	}
	return
}

func (s *cacheProxy) save(ctx *context.Context) (r string, t int, err error) {
	key, value, expiresAt, err := s.getSaveParams(ctx)
	if err != nil {
		return
	}
	client, err := s.getMemcacheClient(ctx)
	if err != nil {
		return
	}
	err = client.Set(key, value, expiresAt)
	if err != nil {
		err = fmt.Errorf("set错误(err:%v)", err)
	}
	r = "SUCCESS"
	return
}
