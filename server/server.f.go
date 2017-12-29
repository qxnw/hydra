package server

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/qxnw/lib4go/jsons"

	"github.com/qxnw/hydra/conf"
)

/*
func SaveCronTasks(tagName string, cn conf.Conf, t *Task) error {
	if tagName == "" {
		return nil
	}
	client, err := getRedisClient(tagName, cn)
	if err != nil {
		return err
	}
	_, err = client.HMSet(cn.Translate("hydra:{@domain}:{@name}:cron:tasks"), t.GetMap()).Result()
	if err != nil {
		err = fmt.Errorf("保存cron执行记录失败:err:%v", err)
		return err
	}
	return nil
}
*/
//SaveCronExecuteHistory 保存 cron执行的历史记录
func SaveCronExecuteHistory(tagName string, cn conf.Conf, t *Task) error {
	if tagName == "" {
		return nil
	}
	client, err := getRedisClient(tagName, cn)
	if err != nil {
		return err
	}
	buff, err := jsons.Marshal(t)
	if err != nil {
		return err
	}
	member := redis.Z{Member: string(buff)}
	member.Score, _ = strconv.ParseFloat(time.Now().Format("20060102150405"), 64)

	key := fmt.Sprintf(cn.Translate("hydra:@domain:@name:cron:@tag:task:%s"), t.Name)
	r, err := client.ZAdd(key, member).Result()
	if err != nil || r == 0 {
		err = fmt.Errorf("保存cron执行记录失败:c:%d,err:%v", r, err)
		return err
	}
	min := time.Now().Add(time.Hour * -10000).Format("20060102150405")
	max := time.Now().Add(time.Hour * -1).Format("20060102150405")
	_, err = client.ZRemRangeByScore(key, min, max).Result()
	return err
}
