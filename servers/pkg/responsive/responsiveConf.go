package responsive

import (
	"fmt"
	"strings"

	xconf "github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/conf"
)

//ResponsiveBaseConf 响应式配置
type ResponsiveBaseConf struct {
	Nconf xconf.Conf
	Oconf xconf.Conf
}

//NewResponsiveBaseConf 创建响应式配置
func NewResponsiveBaseConf(oconf xconf.Conf, nconf xconf.Conf) *ResponsiveBaseConf {
	return &ResponsiveBaseConf{
		Nconf: nconf,
		Oconf: oconf,
	}
}

//IsChanged 配置是否已经发生变化
func (s *ResponsiveBaseConf) IsChanged() bool {
	return s.Oconf.GetVersion() != s.Nconf.GetVersion()
}

//IsStoped 服务器已停止
func (s *ResponsiveBaseConf) IsStoped() bool {
	return strings.EqualFold(s.Nconf.String("status"), servers.ST_STOP)
}

//GetMetric 获取metric配置
func (s *ResponsiveBaseConf) GetMetric() (enable bool, host string, dataBase string, userName string, password string, cron string, err error) {
	//设置metric服务器监控数据
	metric, err := s.Nconf.GetNodeWithSectionName("metric", "#@path/metric")
	if err != nil {
		if !s.Nconf.Has("#@path/metric") {
			err = conf.ErrNoSetting
			return
		}
		err = fmt.Errorf("metric未配置或配置有误:%+v", err)
		return false, "", "", "", "", "", err
	}
	enable = true
	host = metric.String("host")
	dataBase = metric.String("dataBase")
	userName = metric.String("userName")
	password = metric.String("password")
	cron = metric.String("cron", "@every 1m")
	enable, _ = metric.Bool("enable", true)
	if host == "" || dataBase == "" {
		err = fmt.Errorf("metric配置错误:host 和 dataBase不能为空(`host:%s，dataBase:%s)", host, dataBase)
		return false, "", "", "", "", "", err
	}
	if !strings.Contains(host, "://") {
		host = "http://" + host
	}
	return enable, host, dataBase, userName, password, cron, nil

}

//ISetMetricHandler 设置任务列表
type ISetMetricHandler interface {
	SetMetric(host string, dataBase string, userName string, password string, cron string) error
	StopMetric() error
}

//SetMetric 设置metric服务器监控
func (s *ResponsiveBaseConf) SetMetric(set ISetMetricHandler) (hasSet bool, err error) {
	enable, host, dataBase, userName, password, span, err := s.GetMetric()
	if err == conf.ErrNoSetting || !enable {
		err = set.StopMetric()
		hasSet = false
		return
	}
	if err != nil {
		return false, err
	}
	err = set.SetMetric(host, dataBase, userName, password, span)
	hasSet = err == nil
	return
}
