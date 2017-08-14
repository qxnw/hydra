package cache

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/types"
)

func (s *cacheProxy) getGetParams(ctx *context.Context) (key string, err error) {
	key, err = ctx.Input.Get("key") //已包含key
	if err == nil {
		return
	}
	if types.IsEmpty(ctx.Input.Body) { //未传入body
		err = fmt.Errorf("输入参数不包含key")
		return
	}
	body := ctx.Input.Body //根据body 转换数据
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
	err = errors.New("form中未包含key标签")
	return

}

func (s *cacheProxy) get(name string, mode string, service string, ctx *context.Context) (response *context.Response, err error) {
	response = context.GetResponse()
	key, err := s.getGetParams(ctx)
	if err != nil {
		response.Failed(406)
		return
	}
	r, err := ctx.Cache.Get(key)
	if err != nil {
		response.Failed(410)
		return
	}
	response.Success(r)
	return
}
