package main

//go build -buildmode=plugin
import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
)

//Container 提交组件上下文
var Container component.IContainer

//DemoService 服务组件示例
type DemoService struct {
	*component.StandardComponent
}

var demoService *DemoService

//NewDemoService 初始化服务组
func NewDemoService() *DemoService {
	c := &DemoService{}
	c.StandardComponent = component.NewStandardComponent("demo-service")
	return c
}

//Handling 每次handle执行前执行
func (cs *DemoService) Handling(name string, mode string, service string, ctx *context.Context) (rs context.Response, err error) {
	//检查用户权限
	return nil, nil
}

//GetComponent 获取GetComponent
func GetComponent(container component.IContainer) (component.IComponent, error) {
	Container = container
	return demoService, nil
}
