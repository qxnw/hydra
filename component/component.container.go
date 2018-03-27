package component

import (
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/lib4go/cache"
	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/influxdb"
	"github.com/qxnw/lib4go/queue"
)

type IContainer interface {
	context.RPCInvoker

	conf.ISystemConf
	conf.IVarConf

	GetRegistry() registry.IRegistry
	GetDefaultCache() (c cache.ICache, err error)
	GetCache(name string) (c cache.ICache, err error)
	GetDefaultDB() (c db.IDB, err error)
	GetDB(name string) (d db.IDB, err error)
	GetDefaultInflux() (c influxdb.IInfluxClient, err error)
	GetInflux(name string) (d influxdb.IInfluxClient, err error)
	GetDefaultQueue() (c queue.IQueue, err error)
	GetQueue(name string) (q queue.IQueue, err error)
	Close() error
}
