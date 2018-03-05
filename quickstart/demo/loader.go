package main

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

func loader() engines.ServiceLoader {
	return func(component *component.StandardComponent, container component.IContainer) {
		component.AddMicroService("/order/query", NewQueryHandler)
	}
}
