package standalone

import (
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/logger"
)

//JSONConfResolver zookeeper配置引擎
type JSONConfResolver struct {
}

//Resolve 从服务器获取数据
func (j *JSONConfResolver) Resolve(adapter string, domain string, tag string, log *logger.Logger, servers []string) (conf.ConfWatcher, error) {
	return NewJSONConfWatcher(domain, tag), nil
}
func init() {
	conf.Register("standalone", &JSONConfResolver{})
}
