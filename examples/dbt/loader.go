package main

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
	"github.com/qxnw/hydra/examples/dbt/services/order"
)

func loader() engines.ServiceLoader {
	return func(component *component.StandardComponent, container component.IContainer) error {
		component.AddMicroService("/order/request", order.NewRequestHandler)
		component.AddMicroService("/order/query", order.NewQueryHandler)
		return nil
	}
}
