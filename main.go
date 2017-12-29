package main

import (
	"runtime"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-oci8"
	_ "github.com/qxnw/hydra/conf/cluster"
	_ "github.com/qxnw/hydra/conf/standalone"
	_ "github.com/qxnw/hydra/engine/alarm"
	_ "github.com/qxnw/hydra/engine/cache"
	_ "github.com/qxnw/hydra/engine/file"
	_ "github.com/qxnw/hydra/engine/http"
	_ "github.com/qxnw/hydra/engine/influx"
	_ "github.com/qxnw/hydra/engine/log"
	_ "github.com/qxnw/hydra/engine/mock"
	_ "github.com/qxnw/hydra/engine/monitor"
	_ "github.com/qxnw/hydra/engine/plugin"
	_ "github.com/qxnw/hydra/engine/registry"
	_ "github.com/qxnw/hydra/engine/report"
	_ "github.com/qxnw/hydra/engine/rpc"
	_ "github.com/qxnw/hydra/engine/ssm"
	"github.com/qxnw/hydra/hydra"
	_ "github.com/qxnw/hydra/server/api"
	_ "github.com/qxnw/hydra/server/cron"
	_ "github.com/qxnw/hydra/server/mq"
	_ "github.com/qxnw/hydra/server/rpc"
	_ "github.com/qxnw/hydra/server/web"
	_ "github.com/qxnw/lib4go/cache/memcache"
	_ "github.com/qxnw/lib4go/cache/redis"
	_ "github.com/qxnw/lib4go/mq/redis"
	_ "github.com/qxnw/lib4go/mq/stomp"
	_ "github.com/qxnw/lib4go/mq/xmq"
	_ "github.com/qxnw/lib4go/queue"
	_ "github.com/qxnw/lib4go/queue/redis"
)

var (
	VERSION = "1.0.1"
)

func main() {
	hydra.Version = VERSION
	runtime.GOMAXPROCS(runtime.NumCPU())
	hydra := hydra.NewHydra()
	defer hydra.Close()
	hydra.Start()
}
