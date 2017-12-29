package server

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/redis"
)

//getRedisClient 获取redis客户端
func getRedisClient(name string, cn conf.Conf) (c *redis.Client, err error) {

	_, redisClient, err := redisCache.SetIfAbsentCb(name, func(input ...interface{}) (c interface{}, err error) {
		name := input[0].(string)
		cnf, err := cn.GetRawNodeWithValue(fmt.Sprintf("#/@domain/var/redis/%s", name), false)
		if err != nil {
			return "", err
		}
		c, err = redis.NewClientByJSON(string(cnf))
		return
	}, name)
	if err != nil {
		err = fmt.Errorf("初始化redis失败:%v", err)
		return
	}
	c = redisClient.(*redis.Client)
	return
}

var redisCache cmap.ConcurrentMap

func init() {
	redisCache = cmap.New(2)
}
