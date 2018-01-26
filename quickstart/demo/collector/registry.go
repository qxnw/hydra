package main

//注册所有服务
func init() {
	demoService = NewDemoService()
	demoService.AddMicroService("/modify", NewOrderHandler)
}
