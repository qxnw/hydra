package conf

import (
	"errors"
	"fmt"

	"github.com/qxnw/lib4go/concurrent/cmap"

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
	metadata      cmap.ConcurrentMap
}

//GetFullName 获取服务名称
func (s *ServerConf) GetFullName() string {
	return fmt.Sprintf("%s.%s(%s)", s.Name, s.Type, s.Cluster)
}
func (s *ServerConf) GetMetadata(key string) interface{} {
	v, _ := s.metadata.Get(key)
	return v
}
func (s *ServerConf) SetMetadata(key string, v interface{}) {
	s.metadata.Set(key, v)
}
func (s *ServerConf) Get(key string) string {
	return s.nconf.String(key)
}
func (s *ServerConf) GetHealthChecker() string {
	return s.HealthChecker
}

//GetString 获取指定节点的值
func (s *ServerConf) GetString(name string, d ...string) string {
	if s.nconf != nil {
		return s.nconf.String(name, d...)
	}
	if len(d) > 0 {
		return d[0]
	}
	return ""

}

//GetInt 获取指定节点的值
func (s *ServerConf) GetInt(name string, d ...int) int {
	if s.nconf != nil {
		v, _ := s.nconf.Int(name, d...)
		return v
	}
	if len(d) > 0 {
		return d[0]
	}
	return 0

}
func NewConf(domain string, name string, typeName string, tag string, registryNode string, mask string, timeout int) *ServerConf {
	return &ServerConf{
		Domain:     domain,
		Name:       name,
		Type:       typeName,
		Cluster:    tag,
		ServerNode: registryNode,
		IP:         net.GetLocalIPAddress(mask),
		Timeout:    timeout,
		metadata:   cmap.New(8),
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
		HealthChecker: nconf.Translate("@healthChecker"),
		ServerNode:    nconf.Translate("/{@domain}/@name/@type/@tag/servers"),
		ServiceNode:   nconf.Translate("/@domain/services/@type"),
		IP:            net.GetLocalIPAddress(nconf.String("mask")),
		metadata:      cmap.New(8),
	}
	c.Timeout, _ = nconf.Int("timeout", 3)
	if c.Timeout == 0 {
		c.Timeout = 3
	}
	return c
}

//ErrNoSetting 未配置
var ErrNoSetting = errors.New("未配置")

//ErrNoChanged 配置未变化
var ErrNoChanged = errors.New("配置未变化")

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

type CircuitBreaker struct {
	ForceBreak      bool                `json:"force-break"`
	Enable          bool                `json:"enable"`
	SwitchWindow    int                 `json:"swith-window"`
	CircuitBreakers map[string]*Breaker `json:"circuit-breakers"`
}
type Breaker struct {
	URL              string `json:"url"`
	RequestPerSecond int    `json:"request-per-second"`
	FailedPercent    int    `json:"failed-request"`
	RejectPerSecond  int    `json:"reject-per-second"`
}
