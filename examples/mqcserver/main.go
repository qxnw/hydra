package main

import (
	"github.com/qxnw/hydra/examples/mqcserver/services/order"
	"github.com/qxnw/hydra/examples/mqcserver/services/user"
	"github.com/qxnw/hydra/hydra"
)

func main() {
	app := hydra.NewApp(
		hydra.WithPlatName("hydra-20"),
		hydra.WithSystemName("collector"),
		hydra.WithServerTypes("mqc-api-rpc"),
		hydra.WithAutoCreateConf(true),
		hydra.WithDebug())

	app.Autoflow("/order/query", order.NewQueryHandler)
	app.Autoflow("/order/bind", order.NewBindHandler)
	app.Micro("/message/send", user.NewLoginHandler)
	app.Micro("/order/bind", order.NewBindHandler)
	app.Start()
}
