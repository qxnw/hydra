package main

import (
	"github.com/qxnw/collector/services/activity"
	"github.com/qxnw/collector/services/product"
)

//注册所有服务
func (s *DemoService) registerService() {
	s.AddAutoflowService("/order/query", product.NewQueryHandler)
	s.AddPageService("/product", product.NewProductHandler)
	s.AddMicroService("/order/success", product.NewProductHandler)
	s.AddPageService("/order/success", product.NewProductHandler)
	s.AddPageService("/activity/create", activity.NewActivityHandler)
}
