package main

import (
	"runtime"

	_ "github.com/mattn/go-oci8"
	_ "github.com/qxnw/hydra/conf/cluster"
	_ "github.com/qxnw/hydra/conf/standalone"
	_ "github.com/qxnw/hydra/engine/cache"
	_ "github.com/qxnw/hydra/engine/collect"
	_ "github.com/qxnw/hydra/engine/file"
	_ "github.com/qxnw/hydra/engine/goplugin"
	_ "github.com/qxnw/hydra/engine/http"
	_ "github.com/qxnw/hydra/engine/influx"
	_ "github.com/qxnw/hydra/engine/log"
	_ "github.com/qxnw/hydra/engine/mock"
	_ "github.com/qxnw/hydra/engine/registry"
	_ "github.com/qxnw/hydra/engine/rpc_proxy"
	_ "github.com/qxnw/hydra/engine/ssm"
	"github.com/qxnw/hydra/hydra"
	_ "github.com/qxnw/hydra/server/api"
	_ "github.com/qxnw/hydra/server/cron"
	_ "github.com/qxnw/hydra/server/mq"
	_ "github.com/qxnw/hydra/server/rpc"
	_ "github.com/qxnw/hydra/server/web"
	_ "github.com/qxnw/lib4go/mq/stomp"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	hydra := hydra.NewHydra()
	defer hydra.Close()
	hydra.Start()
}
