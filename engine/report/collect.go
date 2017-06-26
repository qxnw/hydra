package report

import (
	"net"
	"net/url"
	"strconv"

	"fmt"

	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/sysinfo/cpu"
	"github.com/qxnw/lib4go/sysinfo/memory"
	"github.com/qxnw/lib4go/transform"
)

func (s *collectProxy) httpCollect(uri string, result []string, report string) (r []string, err error) {
	r = make([]string, 0, 1)
	client := http.NewHTTPClient()
	_, t, err := client.Get(uri)
	if err == nil && t == 200 {
		return
	}
	u, err := url.Parse(uri)
	if err != nil {
		err = fmt.Errorf("请求的URL配置有误:%v", uri)
		return
	}
	tf := transform.New()
	tf.Set("status", strconv.Itoa(t))
	tf.Set("host", u.Host)
	tf.Set("err", fmt.Sprintf("%v", err))

	r = append(r, tf.Translate(report))
	return r, nil
}
func (s *collectProxy) tcpCollect(host string, report string) (r []string, err error) {
	r = make([]string, 0, 1)
	conn, err := net.Dial("tcp", host)
	if err == nil {
		conn.Close()
		return r, nil
	}
	tf := transform.NewMap(map[string]string{})
	tf.Set("host", host)
	tf.Set("err", fmt.Sprintf("%v", err))
	r = append(r, tf.Translate(report))
	return r, nil
}
func (s *collectProxy) registryCollect(path string, minValue int, report string) (r []string, err error) {
	r = make([]string, 0, 1)
	data, _, err := s.registry.GetChildren(path)
	if err != nil {
		return
	}
	if len(data) > minValue {
		return
	}
	tf := transform.NewMap(map[string]string{})
	tf.Set("path", path)
	r = append(r, tf.Translate(report))
	return r, nil
}
func (s *collectProxy) dbCollect(db *db.DB, query string, minValue int, report string) (r []string, err error) {
	rows, _, _, err := db.Query(query, map[string]interface{}{})
	if err != nil {
		err = fmt.Errorf("从数据库获取监控数据失败:%v", err)
		return
	}
	if len(rows) == 0 {
		return
	}
	r = make([]string, 0, len(rows))
	for _, row := range rows {
		tf := transform.New()
		for k, v := range row {
			tf.Set(k, fmt.Sprintf("%v", v))
		}
		r = append(r, tf.Translate(report))
	}
	return r, nil

}
func (s *collectProxy) cpuCollect(query string, maxValue float64, report string) (r []string, err error) {
	cpuInfo := cpu.GetInfo()
	if cpuInfo.UsedPercent < maxValue {
		return
	}
	r = make([]string, 0, 1)
	tf := transform.New()
	tf.Set("idle", fmt.Sprintf("%.2f", cpuInfo.Idle))
	tf.Set("total", fmt.Sprintf("%.2f", cpuInfo.Total))
	tf.Set("used", fmt.Sprintf("%.2f", cpuInfo.UsedPercent))
	r = append(r, tf.Translate(report))
	return r, nil
}
func (s *collectProxy) memCollect(query string, maxValue float64, report string) (r []string, err error) {
	memoryInfo := memory.GetInfo()
	if memoryInfo.UsedPercent < maxValue {
		return
	}
	r = make([]string, 0, 1)
	tf := transform.New()
	tf.Set("idle", fmt.Sprintf("%d", memoryInfo.Idle))
	tf.Set("total", fmt.Sprintf("%d", memoryInfo.Total))
	tf.Set("used", fmt.Sprintf("%.2f", memoryInfo.UsedPercent))
	r = append(r, tf.Translate(report))
	return r, nil
}
