package main

import (
	"github.com/qxnw/hydra/examples/webserver/services/order"
	"github.com/qxnw/hydra/examples/webserver/services/user"
	"github.com/qxnw/hydra/hydra"
)

func main() {
	app := hydra.NewApp(
		hydra.WithPlatName("hydra-20"),
		hydra.WithSystemName("collector"),
		hydra.WithServerTypes("web"),
		hydra.WithAutoCreateConf(),
		hydra.WithDebug())

	app.Page("/user/login", user.NewLoginHandler)
	app.Page("/order/query", order.NewQueryHandler)
	app.Micro("/order/bind", order.NewBindHandler)

	app.Start()
}
