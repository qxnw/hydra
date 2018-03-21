package main

import (
	"github.com/qxnw/helloworld/services/order"
	"github.com/qxnw/helloworld/services/user"
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

func loader() engines.ServiceLoader {
	return func(component *component.StandardComponent, container component.IContainer) {
		component.AddMicroService("/user/login", user.NewLoginHandler)
		component.AddMicroService("/order/query", order.NewQueryHandler)
	}
}
