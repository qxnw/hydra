package mock

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

func LoadService(r *component.StandardComponent, c component.IContainer) {
	r.AddMicroService("/mock/raw/request", RawRequest(c), "mock")
	r.AddAutoflowService("/mock/raw/request", RawRequest(c), "mock")
}
func init() {
	engines.AddLoader("mock", LoadService)
}
