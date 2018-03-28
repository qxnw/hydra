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
	GetCache(names ...string) (c cache.ICache, err error)
	GetDB(names ...string) (d db.IDB, err error)
	GetInflux(names ...string) (d influxdb.IInfluxClient, err error)
	GetQueue(names ...string) (q queue.IQueue, err error)
	Close() error
}
