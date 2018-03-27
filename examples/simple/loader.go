package main

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
	"github.com/qxnw/hydra/examples/simple/services/order"
	"github.com/qxnw/hydra/examples/simple/services/user"
)

func loader() engines.ServiceLoader {
	return func(component *component.StandardComponent, container component.IContainer) error {
		component.AddMicroService("/user/login", user.NewLoginHandler)
		component.AddMicroService("/order/query", order.NewQueryHandler)
		component.AddMicroService("/order/bind", order.NewBindHandler)
		return nil
	}
}
