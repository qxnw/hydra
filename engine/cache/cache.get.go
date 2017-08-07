package cache

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/types"
)

func (s *cacheProxy) getGetParams(ctx *context.Context) (key string, err error) {
	key, err = ctx.GetInput().Get("key")
	if err == nil {
		return
	}
	body, _ := ctx.GetBody()
	if err != nil && !types.IsEmpty(body) {
		inputMap := make(map[string]interface{})
		inputMap, err = jsons.Unmarshal([]byte(body))
		if err != nil {
			err = fmt.Errorf("body不是有效的json数据，[%v](err:%v)", body, err)
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
		return
	}
	err = errors.New("form中未包含key标签")
	return

}

func (s *cacheProxy) get(ctx *context.Context) (r string, t int, err error) {
	key, err := s.getGetParams(ctx)
	if err != nil {
		t = 406
		return
	}
	client, err := s.getMemcacheClient(ctx)
	if err != nil {
		return
	}
	r, err = client.Get(key)
	if err != nil {
		t = 410
	}
	return
}
