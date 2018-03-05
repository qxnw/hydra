package registry

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/engines"
)

func loadService(r *component.StandardComponent, i component.IContainer) {
	r.AddMicroService("/registry/backup", Backup(i), "registry")
	r.AddMicroService("/registry/get/value", GetNodeValue(i), "registry")
	r.AddMicroService("/registry/get/children", GetChildrenNodes(i), "registry")
	r.AddMicroService("/registry/create/path", CreatePersistentNode(i), "registry")
	r.AddMicroService("/registry/create/ephemeral/path", CreateEphemeralNode(i), "registry")
	r.AddMicroService("/registry/create/sequence/path", CreateSEQNode(i), "registry")
	r.AddMicroService("/registry/update/value", UpdateNodeValue(i), "registry")
	r.AddMicroService("/registry/domain/copy", Copy(i), "registry")
	r.AddAutoflowService("/registry/backup", Backup(i), "registry")

}
func init() {
	engines.AddLoader("registry", loadService)
}
