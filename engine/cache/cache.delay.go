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
	body, _ := ctx.GetBody()
	key, err = ctx.GetInput().Get("key")
	if err != nil && !types.IsEmpty(body) {
		inputMap := make(map[string]interface{})
		inputMap, err = jsons.Unmarshal([]byte(body))
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
	key, err = ctx.GetInput().Get("key")
	if err != nil {
		err = errors.New("form中未包含key标签")
		return
	}
	expires, err := ctx.GetInput().Get("expiresAt")
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

func (s *cacheProxy) delay(ctx *context.Context) (r string, t int, err error) {
	key, expiresAt, err := s.getDelayParams(ctx)
	if err != nil {
		return
	}
	client, err := s.getMemcacheClient(ctx)
	if err != nil {
		return
	}
	err = client.Delay(key, expiresAt)
	if err != nil {
		err = fmt.Errorf("delay错误(err:%v)", err)
	}
	r = "SUCCESS"

	return
}
