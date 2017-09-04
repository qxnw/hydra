package rpc

import (
	"fmt"
	"strings"

	"github.com/qxnw/lib4go/transform"
)

//检查路由配置，匹配服务
//1. name:/order/request,match:/order/request service:/order/request ====> reg:/order/request
//2.name:/order/:a,match:/order/request service /order/request ======>reg:/order/request
//3. name:/order/:a match:/order/@a service  /order/request / /order/query ======>reg:/order/request /order/query
func (s *RPCServer) getServiceList(all []string) []string {
	r := make([]string, 0, 0)
	serviceMap := make(map[string]string)
	for _, router := range s.apiRouters {
		rservices := getServices(router.Path, router.service, all...)
		for _, service := range rservices {
			if _, ok := serviceMap[service]; !ok {
				serviceMap[service] = service
				r = append(r, service)
			}
		}
	}
	serviceMap = nil
	return r
}
func getServices(path, service string, all ...string) []string {
	if !strings.Contains(path, ":") {
		return []string{path}
	}
	if !strings.Contains(service, "@") {
		return getDepthOneService(path, service)
	}
	return getDepthTwoService(path, service, all...)
}
func getDepthTwoService(path, service string, services ...string) []string {
	srvs := make([]string, 0, 0)
	rpath := strings.Replace(path, ":", "@", -1)
	rservice := strings.Replace(service, "@", ":", -1)
	nrouter := NewRouter()
	nrouter.Route([]string{"REQUEST"}, rservice, func(*Context) {})
	for _, srvName := range services {
		r, params := nrouter.Match(srvName, "REQUEST")
		if r != nil && len(params) > 0 {
			trans := transform.New()
			params.Each(func(k, v string) {
				trans.Set(k[1:], v)
			})
			rservice := trans.TranslateAll(rpath, false)
			if strings.Contains(rservice, "@") {
				continue
			}
			srvs = append(srvs, rservice)
		}
	}
	return srvs
}

func getDepthOneService(path string, services ...string) []string {
	srvs := make([]string, 0, 2)
	nrouter := NewRouter()
	nrouter.Route([]string{"REQUEST"}, path, func(*Context) {})
	for _, service := range services {
		r, params := nrouter.Match(service, "REQUEST")
		fmt.Println("match:", r, params, path, service)
		if r != nil && len(params) > 0 {
			fmt.Println("match:", path, service, params)
			trans := transform.New()
			params.Each(func(k string, v string) {
				trans.Set(k[1:], v)
			})
			rservice := trans.TranslateAll(strings.Replace(path, ":", "@", -1), false)
			fmt.Println("r.service:", rservice)
			if strings.Contains(path, "@") {
				continue
			}
			srvs = append(srvs, rservice)
		}
	}
	return srvs
}
