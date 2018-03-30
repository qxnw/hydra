package main

import (
	"github.com/qxnw/hydra/examples/simple/services/order"
	"github.com/qxnw/hydra/examples/simple/services/user"
	"github.com/qxnw/hydra/hydra"
)

func AddServices(app *hydra.MicroApp) {
	app.Micro("/user/login", user.NewLoginHandler)
	app.Micro("/order/query", order.NewQueryHandler)
	app.Micro("/order/bind", order.NewBindHandler)
}
