package http

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

func LoadService(r *component.StandardComponent, i component.IContainer) {
	r.AddMicroService("/http/redirect", Redirect(), "http")
	r.AddMicroService("/http/request", Request(), "http")
}
func init() {
	engines.AddServiceLoader("http", LoadService)
}
