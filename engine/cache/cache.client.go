package cache

import (
	"fmt"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/memcache"
)

func (s *cacheProxy) getMemcacheClient(ctx *context.Context) (*memcache.MemcacheClient, error) {
	cacheName, err := ctx.GetArgByName("cache")
	if err != nil {
		return nil, err
	}
	content, err := ctx.GetVarParam("cache", cacheName)
	if err != nil {
		return nil, err
	}
	_, client, err := s.dbs.SetIfAbsentCb(content, func(i ...interface{}) (interface{}, error) {
		cnf, err := conf.NewJSONConfWithJson(content, 0, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("args配置错误无法解析:%s(err:%v)", content, err)
		}
		server := cnf.Strings("server")
		if len(server) == 0 {
			return nil, fmt.Errorf("配置错误:server 和 dataBase不能为空（server:%v）", server)
		}
		mem, err := memcache.New(server)
		return mem, err
	})
	if err != nil {
		return nil, err
	}
	return client.(*memcache.MemcacheClient), err

}
