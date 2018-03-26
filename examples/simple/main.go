package main

import (
	"github.com/qxnw/hydra/engines"
	"github.com/qxnw/hydra/hydra"
)

func main() {
	engines.AddServiceLoader(loader())
	app := hydra.NewApp(
		hydra.WithPlatName("hydrav4"),
		hydra.WithSystemName("collector"),
		hydra.WithServerTypes("api"),
		hydra.WithAutoCreateConf(true),
		hydra.WithDebug())
	app.Start()
}
