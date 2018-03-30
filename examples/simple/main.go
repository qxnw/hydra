package main

import (
	"github.com/qxnw/hydra/examples/simple/services/order"
	"github.com/qxnw/hydra/examples/simple/services/user"
	"github.com/qxnw/hydra/hydra"
)

func main() {
	app := hydra.NewApp(
		hydra.WithPlatName("hydrav4"),
		hydra.WithSystemName("collector"),
		hydra.WithServerTypes("api"),
		hydra.WithAutoCreateConf(true),
		hydra.WithDebug())

	app.Micro("/user/login", user.NewLoginHandler)
	app.Micro("/order/query", order.NewQueryHandler)
	app.Micro("/order/bind", order.NewBindHandler)

	app.Start()
}
