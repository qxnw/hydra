package cache

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/context"

	"strconv"

	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/types"
)

func (s *cacheProxy) getSaveParams(ctx *context.Context) (key string, value string, expiresAt int, err error) {

	key, err1 := ctx.Input.Get("key")
	body := ctx.Input.Body
	if err1 != nil && !types.IsEmpty(body) {
		inputMap := make(map[string]interface{})
		inputMap, err = jsons.Unmarshal([]byte(body))
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
	if err1 != nil {
		err = err1
		return
	}

	value, err = ctx.Input.Get("value")
	if err != nil {
		err = errors.New("form中未包含value标签")
		return
	}
	expiresAt, err = ctx.Input.GetInt("expiresAt")
	if err != nil {
		expiresAt = 0
	}
	return
}

func (s *cacheProxy) save(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
	response = context.GetResponse()
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
