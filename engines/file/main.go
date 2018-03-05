package file

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

//LoadService 加载服务
func LoadService(r *component.StandardComponent, i component.IContainer) {
	r.AddMicroService("/file/upload", FileUpload(), "file")
}

func init() {
	engines.AddLoader("file", LoadService)
}
