package context

import "github.com/qxnw/lib4go/cache"

type TCache struct {
	Values []string
	Value  string
	Err    error
	N      int64
}

// Get 根据key获取memcache中的数据
func (c *TCache) Get(key string) (string, error) {
	return c.Value, c.Err
}

//Decrement 增加变量的值
func (c *TCache) Decrement(key string, delta int64) (n int64, err error) {
	return c.N, c.Err
}

//Increment 减少变量的值
func (c *TCache) Increment(key string, delta int64) (n int64, err error) {
	return c.N, c.Err
}

//Gets 获取多条数据
func (c *TCache) Gets(key ...string) (r []string, err error) {
	return c.Values, c.Err
}

// Add 添加数据到memcache中,如果memcache存在，则报错
func (c *TCache) Add(key string, value string, expiresAt int) error {
	return c.Err
}

// Set 更新数据到memcache中，没有则添加
func (c *TCache) Set(key string, value string, expiresAt int) error {
	return c.Err
}

// Set 更新数据到memcache中，没有则添加
func (c *TCache) Exists(key string) bool {
	return false
}

// Delete 删除memcache中的数据
func (c *TCache) Delete(key string) error {
	return c.Err
}

// Delay 延长数据在memcache中的时间
func (c *TCache) Delay(key string, expiresAt int) error {
	return c.Err
}

// DeleteAll 删除所有缓存数据
func (c *TCache) DeleteAll() error {
	return c.Err
}

type tcacheResolver struct {
	cache *TCache
}

func (s *tcacheResolver) Resolve(address []string, c string) (cache.ICache, error) {
	return s.cache, nil
}

func RegisterTCache(name string, t *TCache) {
	cache.Register(name, &tcacheResolver{cache: t})
}
