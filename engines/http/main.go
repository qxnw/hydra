package http

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

func LoadService(r *component.StandardComponent, c component.IContainer) {
	r.AddMicroService("/http/redirect", Redirect(), "http")
	r.AddMicroService("/http/request", Request(c), "http")
}
func init() {
	engines.AddLoader("http", LoadService)
}
