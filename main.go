package main

import (
	"runtime"

	_ "github.com/qxnw/hydra/engine/goplugin"
	_ "github.com/qxnw/hydra/engine/rpc_proxy"
	_ "github.com/qxnw/hydra/engine/script"
	"github.com/qxnw/hydra/hydra"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	hydra := hydra.NewHydra()
	hydra.Install()
	defer hydra.Close()
	hydra.Start()

}
