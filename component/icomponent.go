package component

import "github.com/qxnw/hydra/context"

//IComponent 提供一组或多组服务的组件
type IComponent interface {
	AddCustomerService(service string, h interface{}, groupNames ...string)
	AddTagPageService(service string, h interface{}, tag string, pages ...string)
	AddPageService(service string, h interface{}, pages ...string)
	AddAutoflowService(service string, h interface{}, tags ...string)
	AddMicroService(service string, h interface{}, tags ...string)
	IsMicroService(service string) bool
	IsAutoflowService(service string) bool
	IsPageService(service string) bool
	IsCustomerService(group string, service string) bool
	GetFallbackHandlers() map[string]interface{}
	AddFallbackHandlers(map[string]interface{})
	LoadServices() error

	GetGroupServices(group string) []string
	GetTagServices(tag string) []string
	GetServices() []string

	GetGroups(service string) []string
	GetPages(service string) []string
	GetTagPages(service string, tagName string) []string
	GetTags(service string) []string
	CheckTag(service string, tagName string) bool

	Fallback(name string, engine string, service string, c *context.Context) (rs context.Response, err error)
	Handling(name string, mode string, service string, c *context.Context) (rs context.Response, err error)
	Handled(name string, mode string, service string, c *context.Context) (rs context.Response, err error)
	Handle(name string, mode string, service string, c *context.Context) (context.Response, error)
	Close() error
}

//Handler context handler
type Handler interface {
	Handle(name string, mode string, service string, c *context.Context) (context.Response, error)
	Close() error
}

type GetHandler interface {
	GetHandle(name string, mode string, service string, c *context.Context) (context.Response, error)
}
type PostHandler interface {
	PostHandle(name string, mode string, service string, c *context.Context) (context.Response, error)
}
type DeleteHandler interface {
	DeleteHandle(name string, mode string, service string, c *context.Context) (context.Response, error)
}
type PutHandler interface {
	PutHandle(name string, mode string, service string, c *context.Context) (context.Response, error)
}

//FallbackHandler context handler
type FallbackHandler interface {
	Fallback(name string, mode string, service string, c *context.Context) (context.Response, error)
}

//GetFallbackHandler context handler
type GetFallbackHandler interface {
	GetFallback(name string, mode string, service string, c *context.Context) (context.Response, error)
}

//PostFallbackHandler context handler
type PostFallbackHandler interface {
	PostFallback(name string, mode string, service string, c *context.Context) (context.Response, error)
}

//PutFallbackHandler context handler
type PutFallbackHandler interface {
	PutFallback(name string, mode string, service string, c *context.Context) (context.Response, error)
}

//DeleteFallbackHandler context handler
type DeleteFallbackHandler interface {
	DeleteFallback(name string, mode string, service string, c *context.Context) (context.Response, error)
}

type FallbackServiceFunc func(name string, mode string, service string, c *context.Context) (rs context.Response, err error)

func (h FallbackServiceFunc) Fallback(name string, mode string, service string, c *context.Context) (rs context.Response, err error) {
	return h(name, mode, service, c)
}

type ServiceFunc func(name string, mode string, service string, c *context.Context) (rs context.Response, err error)

func (h ServiceFunc) Handle(name string, mode string, service string, c *context.Context) (rs context.Response, err error) {
	return h(name, mode, service, c)
}
func (h ServiceFunc) Close() error {
	return nil
}
