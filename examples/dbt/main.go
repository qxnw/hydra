package main

import (
	"github.com/qxnw/hydra/examples/dbt/services/order"
	"github.com/qxnw/hydra/hydra"
)

func main() {
	app := hydra.NewApp(
		hydra.WithPlatName("hydrav-db"),
		hydra.WithSystemName("collector"),
		hydra.WithServerTypes("api"),
		hydra.WithAutoCreateConf(true),
		hydra.WithDebug())
	app.Micro("/order/query", order.NewQueryHandler)
	app.Micro("/order/request", order.NewRequestHandler)
	app.Start()
}
