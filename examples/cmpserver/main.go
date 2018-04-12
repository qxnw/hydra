package main

import (
	"fmt"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/examples/cmpserver/services/order"
	"github.com/qxnw/hydra/examples/cmpserver/services/user"
	"github.com/qxnw/hydra/hydra"
)

func main() {
	app := hydra.NewApp(
		hydra.WithPlatName("hydra-20"),
		hydra.WithSystemName("collector"),
		hydra.WithServerTypes("api"),
		hydra.WithAutoCreateConf(),
		hydra.WithDebug())
	app.Initializing(func(c component.IContainer) error {
		c.Set("abc", "1")
		return nil
	})
	app.Initializing(func(c component.IContainer) error {
		c.Set("efg", "2")
		return nil
	})
	app.Closing(func(c component.IContainer) error {
		fmt.Println(c.Get("efg"), c.Get("abc"))
		return nil
	})
	app.Closing(func(c component.IContainer) error {
		fmt.Println(c.Get("abc"))
		return nil
	})
	app.Micro("/user/login", user.NewLoginHandler)
	app.Micro("/order/query", order.NewQueryHandler)
	app.Micro("/order/bind", order.NewBindHandler)

	app.Start()
}
