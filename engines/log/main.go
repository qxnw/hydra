package log

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

func LoadService(r *component.StandardComponent, i component.IContainer) {
	r.AddMicroService("/log/info", WriteInfoLog(), "log")
	r.AddMicroService("/log/error", WriteErrorLog(), "log")
}
func init() {
	engines.AddLoader("log", LoadService)
}
