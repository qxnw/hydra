package context

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/memcache"
	"github.com/qxnw/lib4go/transform"
)

//ContextCache 缓存
type ContextCache struct {
	ctx *Context
	db  *db.DB
	err error
}

//ErrDataNotExist 数据不存在
var ErrDataNotExist = errors.New("查询的数据不存在")

//Reset 重置context
func (cache *ContextCache) Reset(ctx *Context) (err error) {
	cache.ctx = ctx
	cache.db, cache.err = ctx.GetDB()
	return cache.err
}

//NewContextCache 构建缓存操作对象
func NewContextCache(wx *Context, db *db.DB) *ContextCache {
	ctx := &ContextCache{}
	ctx.ctx = wx
	ctx.db = db
	return ctx
}

//GetCache 获取缓存操作对象
func (cache *ContextCache) GetCache(names ...string) (c *memcache.MemcacheClient, err error) {
	sName := "cache"
	if len(names) > 0 {
		sName = names[0]
	}
	name, ok := cache.ctx.Input.Args[sName]
	if !ok {
		return nil, fmt.Errorf("未配置cache参数(%v)", cache.ctx)
	}
	_, memCached, err := memCache.SetIfAbsentCb(name, func(input ...interface{}) (c interface{}, err error) {
		name := input[0].(string)
		conf, err := cache.ctx.Input.GetVarParam("cache", name)
		if err != nil {
			return nil, err
		}
		configMap, err := jsons.Unmarshal([]byte(conf))
		if err != nil {
			return nil, err
		}
		server, ok := configMap["server"]
		if !ok {
			err = fmt.Errorf("cache[%s]配置文件错误，未包含server节点:%s", name, conf)
			return nil, err
		}
		c, err = memcache.New(strings.Split(server.(string), ";"))
		if err != nil {
			return nil, err
		}
		return
	}, name)
	if err != nil {
		err = fmt.Errorf("初始化memcached失败:%v", err)
		return
	}
	c = memCached.(*memcache.MemcacheClient)
	return
}
func (cache *ContextCache) Set(key string, value string, expiresAt int) error {
	client, err := cache.GetCache()
	if err != nil {
		return err
	}
	return client.Set(key, value, expiresAt)
}
func (cache *ContextCache) Delay(key string, expiresAt int) error {
	client, err := cache.GetCache()
	if err != nil {
		return err
	}
	return client.Delay(key, expiresAt)
}

func (cache *ContextCache) Get(key string) (string, error) {
	client, err := cache.GetCache()
	if err != nil {
		return "", err
	}
	return client.Get(key)
}
func (cache *ContextCache) Delete(key string) error {
	client, err := cache.GetCache()
	if err != nil {
		return err
	}
	err = client.Delete(key)
	if err != nil {
		err = fmt.Errorf("缓存删除失败:%s(err:%v)", key, err)
	}
	return err
}

//GetJSON 从缓存中获取json字符串，缓存中不存在时从数据库中获取
func (cache *ContextCache) GetJSON(tpl []string, input map[string]interface{}) (cvalue string, err error) {
	err = cache.err
	if err != nil {
		return
	}
	sql, key, expireAt, err := cache.getCacheSetting(tpl)
	if err != nil {
		return
	}
	client, err := cache.GetCache()
	if err != nil {
		return
	}
	tf := transform.NewMaps(input)
	key = tf.Translate(key)
	cvalue, _ = client.Get(key)
	if cvalue != "" {
		return
	}
	data, _, _, err := cache.db.Query(sql, input)
	if err != nil {
		return
	}
	buffer, err := jsons.Marshal(&data)
	if err != nil {
		return
	}
	cvalue = string(buffer)
	errx := client.Set(key, cvalue, expireAt)
	if errx != nil {
		cache.ctx.Errorf("保存缓存数据异常：%v", errx)
	}
	return
}

//GetFirstRow 从缓存中获取首行数据，缓存中不存在时从数据中获取并保存到缓存中，数据不存在时返回ErrDataNotExist错误
func (cache *ContextCache) GetFirstRow(tpl []string, input map[string]interface{}) (data db.QueryRow, err error) {
	result, _, _, err := cache.GetDataRows(tpl, input)
	if err != nil {
		return
	}
	if len(result) > 0 {
		return result[0], nil
	}
	return nil, ErrDataNotExist
}

func (cache *ContextCache) getCacheSetting(tpl []string) (sql string, key string, expireAt int, err error) {
	if len(tpl) < 3 {
		err = fmt.Errorf("包含缓存信息的SQL模式配置有误，必须包含3个元素，SQL语句/缓存KEY/过期时间:%v", tpl)
		return
	}
	sql = tpl[0]
	key = tpl[1]
	if key == "" {
		err = fmt.Errorf("包含缓存信息的SQL模式配置有误，key不能为空:%v", tpl)
		return
	}
	expireAt, err = strconv.Atoi(tpl[2])
	if err != nil {
		err = fmt.Errorf("包含缓存信息的SQL模式配置有误，过期时间必须为数字:%v,err:%v", tpl, err)
		return
	}
	return
}

//GetDataRows 从缓存中获取数据集,缓存中不存在时从数据库中获取并保存到缓存中
func (cache *ContextCache) GetDataRows(tpl []string, input map[string]interface{}) (data []db.QueryRow, query string, params []interface{}, err error) {
	err = cache.err
	if err != nil {
		return
	}
	sql, key, expireAt, err := cache.getCacheSetting(tpl)
	if err != nil {
		return
	}

	client, err := cache.GetCache()
	if err != nil {
		return
	}
	tf := transform.NewMaps(input)
	key = tf.Translate(key)
	dstr, _ := client.Get(key)
	if dstr != "" {
		err = json.Unmarshal([]byte(dstr), &data)
		return
	}
	data, query, params, err = cache.db.Query(sql, input)
	if err != nil {
		err = fmt.Errorf("从数据库中查询数据异常:%s,%v,err:%v", sql, input, err)
		return
	}
	if len(data) == 0 {
		return
	}
	cvalue, err := jsons.Marshal(data)
	if err != nil {
		return
	}
	errx := client.Set(key, string(cvalue), expireAt)
	if errx != nil {
		cache.ctx.Errorf("数据保存到缓存中异常:%s,%s,err:%v", key, string(cvalue), errx)
	}
	return
}

var memCache cmap.ConcurrentMap

func init() {
	memCache = cmap.New(2)
}
