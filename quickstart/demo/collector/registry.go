package main

import (
	"github.com/qxnw/collector/services/activity"
	"github.com/qxnw/collector/services/order"
	"github.com/qxnw/collector/services/product"
)

//注册所有服务
func (s *DemoService) registerService() {
	s.AddMicroService("/order/query", order.NewOrderHandler)
	s.AddMicroService("/product", product.NewProductHandler)
	s.AddMicroService("/activity/create", activity.NewActivityHandler)
}
