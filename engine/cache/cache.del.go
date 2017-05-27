package cache

import (
	"fmt"

	"github.com/qxnw/hydra/context"
)

func (s *cacheProxy) del(ctx *context.Context) (r string, err error) {
	key, err := s.getGetParams(ctx)
	if err != nil {
		return "", err
	}
	client, err := s.getMemcacheClient(ctx)
	if err != nil {
		return "", err
	}
	err = client.Delete(key)
	err = fmt.Errorf("delete错误(err:%v)", err)
	if err != nil {
		err = fmt.Errorf("delay错误(err:%v)", err)
	}
	r = "SUCCESS"
	return
}
