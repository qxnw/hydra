package main

import (
	"runtime"
	//_ "github.com/mattn/go-oci8"
	_ "github.com/qxnw/hydra/conf/cluster"
	_ "github.com/qxnw/hydra/conf/standalone"
	_ "github.com/qxnw/hydra/engine/alarm"
	_ "github.com/qxnw/hydra/engine/cache"
	_ "github.com/qxnw/hydra/engine/email"
	_ "github.com/qxnw/hydra/engine/file"
	_ "github.com/qxnw/hydra/engine/goplugin"
	_ "github.com/qxnw/hydra/engine/http"
	_ "github.com/qxnw/hydra/engine/influx"
	_ "github.com/qxnw/hydra/engine/registry"
	_ "github.com/qxnw/hydra/engine/rpc_proxy"
	_ "github.com/qxnw/hydra/engine/script"
	_ "github.com/qxnw/hydra/engine/sms"
	"github.com/qxnw/hydra/hydra"
	_ "github.com/qxnw/hydra/server/api"
	_ "github.com/qxnw/hydra/server/cron"
	_ "github.com/qxnw/hydra/server/mq"
	_ "github.com/qxnw/hydra/server/rpc"
	_ "github.com/qxnw/hydra/service/discovery"
	_ "github.com/qxnw/hydra/service/register"
	_ "github.com/qxnw/lib4go/mq/kafka"
	_ "github.com/qxnw/lib4go/mq/stomp"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	hydra := hydra.NewHydra()
	hydra.Install()
	defer hydra.Close()
	hydra.Start()
}
