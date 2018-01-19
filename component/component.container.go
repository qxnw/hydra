package component

import (
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
)

type IContainer interface {
	context.RPCInvoker
	GetVarParam(tp string, name string) (string, error)
	GetDomainName() string
	GetServerName() string
	GetRegistry() registry.Registry
}
