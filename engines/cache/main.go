package cache

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

//LoadService 加载服务
func LoadService(r *component.StandardComponent, i component.IContainer) {
	r.AddMicroService("/cache/memcached/get", Get(i), "cache")
	r.AddMicroService("/cache/memcached/save", Save(i), "cache")
	r.AddMicroService("/cache/memcached/del", Delete(i), "cache")
	r.AddMicroService("/cache/memcached/delay", Delay(i), "cache")
}
func init() {
	engines.AddServiceLoader("cache", LoadService)
}
