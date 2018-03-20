package micro

type option struct {
	registryAddr string
	name         string
	isDebug      bool
}

//Option 配置选项
type Option func(*option)

//WithRegistry 设置注册中心地址
func WithRegistry(addr string) Option {
	return func(o *option) {
		o.registryAddr = addr
	}
}

//WithName 设置系统名称
func WithName(name string) Option {
	return func(o *option) {
		o.name = name
	}
}

//WithDebug 设置dubug模式
func WithDebug() Option {
	return func(o *option) {
		o.isDebug = true
	}
}

//WithProduct 设置产品模式
func WithProduct() Option {
	return func(o *option) {
		o.isDebug = false
	}
}
