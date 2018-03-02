package influx

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

func LoadService(r *component.StandardComponent, i component.IContainer) {
	r.AddMicroService("/influx/query", Query(i), "influx")
	r.AddMicroService("/influx/save", Save(i), "influx")
}
func init() {
	engines.AddServiceLoader("influx", LoadService)
}
