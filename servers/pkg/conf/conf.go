package conf

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/net"
)

type ServerConf struct {
	nconf         conf.Conf
	Raw           string
	Domain        string
	Name          string
	Type          string
	Cluster       string
	ServerNode    string
	ServiceNode   string
	IP            string
	Hosts         []string
	Headers       map[string]string
	HealthChecker string
	Timeout       int
	JWTAuth       *Auth
}

//GetFullName 获取服务名称
func (s *ServerConf) GetFullName() string {
	return fmt.Sprintf("%s.%s(%s)", s.Name, s.Type, s.Cluster)
}
func (s *ServerConf) Get(key string) string {
	return s.nconf.String(key)
}
func (s *ServerConf) GetHealthChecker() string {
	return s.HealthChecker
}

func NewConf(domain string, name string, typeName string, tag string, registryNode string, mask string, timeout int) *ServerConf {
	return &ServerConf{
		Domain:     domain,
		Name:       name,
		Type:       typeName,
		Cluster:    tag,
		ServerNode: registryNode,
		IP:         net.GetLocalIPAddress(mask),
		JWTAuth:    &Auth{Enable: false},
		Timeout:    timeout,
	}
}

func NewConfBy(nconf conf.Conf) *ServerConf {
	c := &ServerConf{
		Raw:           nconf.GetContent(),
		nconf:         nconf,
		Domain:        nconf.String("domain"),
		Name:          nconf.String("name"),
		Type:          nconf.String("type"),
		Cluster:       nconf.String("tag"),
		HealthChecker: nconf.Translate("@healthChecker}"),
		ServerNode:    nconf.Translate("{@category_path}/servers"),
		ServiceNode:   nconf.Translate("/@domain/services/@type"),
		IP:            net.GetLocalIPAddress(nconf.String("mask")),
		JWTAuth:       &Auth{Enable: false},
	}
	c.Timeout, _ = nconf.Int("timeout", 3)
	if c.Timeout == 0 {
		c.Timeout = 3
	}
	return c
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
	Path  string
	Left  string
	Right string
	Files []string
}
type Queue struct {
	Name        string
	Queue       string
	Engine      string
	Service     string
	Setting     string
	Concurrency int
	Handler     interface{}
}
type Task struct {
	Name    string      `json:"name"`
	Cron    string      `json:"cron"`
	Input   string      `json:"input,omitempty"`
	Body    string      `json:"body,omitempty"`
	Engine  string      `json:"engine,omitempty"`
	Service string      `json:"service"`
	Setting string      `json:"setting,omitempty"`
	Next    string      `json:"next"`
	Last    string      `json:"last"`
	Handler interface{} `json:"handler,omitempty"`
}
