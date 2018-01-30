package http

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/net"
)

type ServerConf struct {
	Domain               string
	Name                 string
	Type                 string
	Cluster              string
	RegistryNode         string
	IP                   string
	Hosts                []string
	Headers              map[string]string
	OnlyAllowAjaxRequest bool
	JWTAuth              *Auth
}

//GetFullName 获取服务名称
func (s *ServerConf) GetFullName() string {
	return fmt.Sprintf("%s.%s(%s)", s.Name, s.Type, s.Cluster)
}

func NewServerConf(nconf conf.Conf) *ServerConf {
	return &ServerConf{
		Domain:       nconf.String("domain"),
		Name:         nconf.String("name"),
		Type:         nconf.String("type"),
		Cluster:      nconf.String("tag"),
		RegistryNode: nconf.Translate("{@category_path}/servers/{@tag}"),
		IP:           net.GetLocalIPAddress(nconf.String("mask")),
	}
}

type Auth struct {
	Name     string
	ExpireAt int64
	Mode     string
	Secret   string
	Exclude  []string
	Enable   bool
}
type Router struct {
	Name    string
	Action  []string
	Engine  string
	Service string
	Setting string
	Handler interface{}
}
type View struct {
	ViewPath string
	Left     string
	Right    string
}
