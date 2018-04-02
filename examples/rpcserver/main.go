package main

import (
	"github.com/qxnw/hydra/engines"
	"github.com/qxnw/hydra/examples/rpcserver/services/order"
	"github.com/qxnw/hydra/examples/rpcserver/services/user"
	"github.com/qxnw/hydra/hydra"
)

func main() {
	engines.AddServiceLoader(loader())
	app := hydra.NewApp(
		hydra.WithPlatName("hydrav4"),
		hydra.WithSystemName("collector"),
		hydra.WithServerTypes("rpc"),
		hydra.WithAutoCreateConf(true),
		hydra.WithDebug())

	app.Micro("/user/login", user.NewLoginHandler)
	app.Micro("/order/query", order.NewQueryHandler)
	app.Micro("/order/bind", order.NewBindHandler)
	app.Start()
}
