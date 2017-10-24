package context

import (
	"fmt"

	"github.com/qxnw/lib4go/redis"
)

//GetRedisClient 获取redis客户端
func (cc *Context) GetRedisClient(names ...string) (c *redis.Client, err error) {
	sName := "redis"
	if len(names) > 0 {
		sName = names[0]
	}
	name, ok := cc.Input.Args[sName]
	if !ok {
		return nil, fmt.Errorf("未配置redis参数(%v)", cc.Input.Args)
	}
	_, redisClient, err := memCache.SetIfAbsentCb(name, func(input ...interface{}) (c interface{}, err error) {
		name := input[0].(string)
		conf, err := cc.Input.GetVarParam("redis", name)
		if err != nil {
			return nil, err
		}
		c, err = redis.NewClientByJSON(conf)
		return
	}, name)
	if err != nil {
		err = fmt.Errorf("初始化redis失败:%v", err)
		return
	}
	c = redisClient.(*redis.Client)
	return
}
