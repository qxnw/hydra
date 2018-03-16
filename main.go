package main

import (
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
	"github.com/qxnw/hydra/hydra"
)

var (
	VERSION = "2.0.1"
)

//dev-whj 分支
func main() {

	//dev 分支修改
	hydra.Version = VERSION
	hydra := hydra.NewHydra()
	defer hydra.Close()
	hydra.Start()
}
