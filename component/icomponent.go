package component

import "github.com/qxnw/hydra/context"

type IComponent interface {
	LoadServices() error
	GetServices() []string
	Handling(name string, mode string, service string, c *context.Context) (rs context.Response, err error)
	Handled(name string, mode string, service string, c *context.Context) (rs context.Response, err error)
	Handle(name string, mode string, service string, c *context.Context) (context.Response, error)
	Close() error
}

type HandlerFunc func(name string, mode string, service string, c *context.Context) (context.Response, error)

func (h HandlerFunc) Handle(name string, mode string, service string, c *context.Context) (context.Response, error) {
	return h(name, mode, service, c)
}

type SHandlerFunc func(name string, mode string, service string, c *context.Context) (*context.StandardResponse, error)

func (h SHandlerFunc) Handle(name string, mode string, service string, c *context.Context) (*context.StandardResponse, error) {
	return h(name, mode, service, c)
}

type WHandlerFunc func(name string, mode string, service string, c *context.Context) (*context.WebResponse, error)

func (h WHandlerFunc) Handle(name string, mode string, service string, c *context.Context) (*context.WebResponse, error) {
	return h(name, mode, service, c)
}

type preHandler interface {
	PreHandle() error
}

//Handler context handler
type Handler interface {
	Handle(name string, mode string, service string, c *context.Context) (context.Response, error)
	Close() error
}
type MapHandler interface {
	Handle(name string, mode string, service string, c *context.Context) (*context.MapResponse, error)
	Close() error
}
type StandardHandler interface {
	Handle(name string, mode string, service string, c *context.Context) (*context.StandardResponse, error)
	Close() error
}
type ObjectHandler interface {
	Handle(name string, mode string, service string, c *context.Context) (*context.ObjectResponse, error)
	Close() error
}
type WebHandler interface {
	Handle(name string, mode string, service string, c *context.Context) (*context.WebResponse, error)
	Close() error
}

type ServiceFunc func(name string, mode string, service string, c *context.Context) (rs context.Response, err error)

func (h ServiceFunc) Handle(name string, mode string, service string, c *context.Context) (rs context.Response, err error) {
	return h(name, mode, service, c)
}
func (h ServiceFunc) Close() error {
	return nil
}

type MapServiceFunc func(name string, mode string, service string, c *context.Context) (rs *context.MapResponse, err error)

func (h MapServiceFunc) Handle(name string, mode string, service string, c *context.Context) (rs *context.MapResponse, err error) {
	return h(name, mode, service, c)
}
func (h MapServiceFunc) Close() error {
	return nil
}

type StandardServiceFunc func(name string, mode string, service string, c *context.Context) (rs *context.StandardResponse, err error)

func (h StandardServiceFunc) Handle(name string, mode string, service string, c *context.Context) (rs *context.StandardResponse, err error) {
	return h(name, mode, service, c)
}
func (h StandardServiceFunc) Close() error {
	return nil
}
