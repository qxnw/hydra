package main

import (
	"runtime"

	_ "github.com/go-sql-driver/mysql"

	_ "github.com/qxnw/hydra/conf/cluster"
	_ "github.com/qxnw/hydra/conf/standalone"
	"github.com/qxnw/hydra/hydra"
	_ "github.com/qxnw/hydra/servers/cron"
	_ "github.com/qxnw/hydra/servers/http"
	_ "github.com/qxnw/hydra/servers/mqc"
	_ "github.com/qxnw/hydra/servers/rpc"
	_ "github.com/qxnw/lib4go/cache/memcache"
	_ "github.com/qxnw/lib4go/cache/redis"
	_ "github.com/qxnw/lib4go/mq/redis"
	_ "github.com/qxnw/lib4go/mq/stomp"
	_ "github.com/qxnw/lib4go/mq/xmq"
	_ "github.com/qxnw/lib4go/queue"
	_ "github.com/qxnw/lib4go/queue/redis"

	_ "github.com/qxnw/hydra/engines/alarm"
	_ "github.com/qxnw/hydra/engines/cache"
	_ "github.com/qxnw/hydra/engines/file"
	_ "github.com/qxnw/hydra/engines/http"
	_ "github.com/qxnw/hydra/engines/influx"
	_ "github.com/qxnw/hydra/engines/log"
	_ "github.com/qxnw/hydra/engines/mock"
	_ "github.com/qxnw/hydra/engines/monitor"
	_ "github.com/qxnw/hydra/engines/registry"
	_ "github.com/qxnw/hydra/engines/ssm"
)

var (
	VERSION = "2.0.1"
)

func main() {
	hydra.Version = VERSION
	runtime.GOMAXPROCS(runtime.NumCPU())
	hydra := hydra.NewHydra()
	defer hydra.Close()
	hydra.Start()
}
