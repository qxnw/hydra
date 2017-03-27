package json

import "github.com/qxnw/hydra/conf"

//JSONConfResolver zookeeper配置引擎
type JSONConfResolver struct {
}

//Resolve 从服务器获取数据
func (j *JSONConfResolver) Resolve(args ...string) (conf.ConfWatcher, error) {
	return NewJSONConfWatcher(), nil
}
func init() {
	conf.Register("json", &JSONConfResolver{})
}
