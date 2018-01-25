package mock

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

func LoadService(r *component.StandardComponent, i component.IContainer) {
	r.AddMicroService("/mock/raw/request", RawRequest(), "mock")
	r.AddAutoflowService("/mock/raw/request", RawRequest(), "mock")
}
func init() {
	engines.AddServiceLoader("mock", LoadService)
}
