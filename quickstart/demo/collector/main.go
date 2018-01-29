package main

//go build -buildmode=plugin
import (
	"energy/coupon-services/conf"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

//DemoService 服务组件示例
type DemoService struct {
	*component.StandardComponent
}

//NewDemoService 初始化服务组
func NewDemoService(c component.IContainer) *DemoService {
	s := &DemoService{}
	s.StandardComponent = component.NewStandardComponent("demo-service", c)
	return s
}

//Handling 每次handle执行前执行
func (cs *DemoService) Handling(name string, mode string, service string, ctx *context.Context) (rs context.Response, err error) {
	//检查用户权限
	return nil, nil
}

//LoadServices 加载组件
func (cs *DemoService) LoadServices() error {
	cs.registerService()
	err := conf.Init(cs.Container)
	if err != nil {
		return err
	}
	return cs.StandardComponent.LoadServices()
}

//GetComponent 获取GetComponent
func GetComponent(container component.IContainer) (component.IComponent, error) {
	return NewDemoService(container), nil
}
