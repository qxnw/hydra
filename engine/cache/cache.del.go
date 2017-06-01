package cache

import (
	"fmt"

	"github.com/qxnw/hydra/context"
)

func (s *cacheProxy) del(ctx *context.Context) (r string, t int, err error) {
	key, err := s.getGetParams(ctx)
	if err != nil {
		return
	}
	client, err := s.getMemcacheClient(ctx)
	if err != nil {
		return
	}
	err = client.Delete(key)
	err = fmt.Errorf("delete错误(err:%v)", err)
	if err != nil {
		err = fmt.Errorf("delay错误(err:%v)", err)
	}
	r = "SUCCESS"
	return
}
