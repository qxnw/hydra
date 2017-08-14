package cache

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/types"
)

func (s *cacheProxy) getDelayParams(ctx *context.Context) (key string, expiresAt int, err error) {
	key, err = ctx.Input.Get("key")
	if err == nil { //key不为空
		expiresAt, err = ctx.Input.GetInt("expiresAt")
		if err != nil {
			err = fmt.Errorf("form的内容中expiresAt标签不是有效的数字:(err:%v)", err)
			return "", 0, err
		}
		return
	}
	if types.IsEmpty(ctx.Input.Body) { //未传入body
		err = fmt.Errorf("输入参数不包含key")
		return
	}
	inputMap := make(map[string]interface{}) //根据body转换参数
	inputMap, err = jsons.Unmarshal([]byte(ctx.Input.Body))
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

func (s *cacheProxy) delay(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
	response = context.GetResponse()
	key, expiresAt, err := s.getDelayParams(ctx)
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
