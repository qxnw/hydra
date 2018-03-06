package component

import (
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/cache"
	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/influxdb"
	"github.com/qxnw/lib4go/queue"
)

type IContainer interface {
	context.RPCInvoker
	GetVarParam(tp string, name string) (string, error)
	GetDomainName() string
	GetServerName() string
	GetServerType() string
	GetRegistry() registry.Registry
	GetDefaultCache() (c cache.ICache, err error)
	GetCache(name string) (c cache.ICache, err error)
	GetConf(conf interface{}) (c interface{}, err error)
	GetDefaultDB() (c *db.DB, err error)
	GetDB(name string) (d *db.DB, err error)
	GetDefaultInflux() (c *influxdb.InfluxClient, err error)
	GetInflux(name string) (d *influxdb.InfluxClient, err error)
	GetDefaultQueue() (c queue.IQueue, err error)
	GetQueue(name string) (q queue.IQueue, err error)
	Close() error
}
