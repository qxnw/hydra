package rpc

//IServiceRegister 服务注册接口
type IServiceRegister interface {
	Register(string, string) error
	UnRegister(string, string) error
}

func register(i IServiceRegister, services []string, address string) {
	if i == nil || len(services) == 0 {
		return
	}
	for _, v := range services {
		i.Register(v, address)
	}
}

func unRegister(i IServiceRegister, services []string, address string) {
	if i == nil || len(services) == 0 {
		return
	}
	for _, v := range services {
		i.UnRegister(v, address)
	}
}
