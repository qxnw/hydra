package report

import (
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"fmt"

	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/sysinfo/cpu"
	"github.com/qxnw/lib4go/sysinfo/disk"
	"github.com/qxnw/lib4go/sysinfo/memory"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

type collector struct {
	Collector []monitorItem `json:"collector"`
}

type monitorItem struct {
	Mode   string          `json:"mode"`
	Params [][]interface{} `json:"params"`
	Report string          `json:"report"`
}

func (s *collectProxy) httpCollect(ctx *context.Context, param []interface{}, report string) (r []string, err error) {
	if len(param) != 1 {
		err = fmt.Errorf("输入参数有误,http.collector,param只能包含1个参数url:%v", param)
		return
	}
	uri := fmt.Sprintf("%v", param[0])
	r = make([]string, 0, 1)
	u, err := url.Parse(uri)
	if err != nil {
		err = fmt.Errorf("请求的URL配置有误:%v", uri)
		return
	}
	client := http.NewHTTPClient()
	_, t, err := client.Get(uri)
	tf := transform.New()
	tf.Set("status", strconv.Itoa(t))
	tf.Set("host", u.Host)
	tf.Set("r", strconv.Itoa(types.DecodeInt(t, 200, 0, 1)))
	tf.Set("err", fmt.Sprintf("%v", err))

	r = append(r, tf.Translate(report))
	return r, nil
}
func (s *collectProxy) tcpCollect(ctx *context.Context, param []interface{}, report string) (r []string, err error) {
	if len(param) != 1 {
		err = fmt.Errorf("输入参数有误,tcp.collector,param未配置tcp.hostname:%v", param)
		return
	}
	host := fmt.Sprintf("%v", param[0])
	r = make([]string, 0, 1)
	conn, err := net.DialTimeout("tcp", host, time.Second)
	if err == nil {
		conn.Close()
	}
	tf := transform.NewMap(map[string]string{})
	tf.Set("host", host)
	tf.Set("err", fmt.Sprintf("%v", err))
	tf.Set("r", strconv.Itoa(types.DecodeInt(err, nil, 0, 1)))
	r = append(r, tf.Translate(report))
	return r, nil
}
func (s *collectProxy) registryCollect(ctx *context.Context, param []interface{}, report string) (r []string, err error) {
	if len(param) < 1 {
		err = fmt.Errorf("输入参数有误,tcp.collector,param未配置服务路径:%v", param)
		return
	}
	path := fmt.Sprintf("%v", param[0])
	minValue := types.ToInt(param[1], 1)
	r = make([]string, 0, 1)
	data, _, err := s.registry.GetChildren(path)
	if err != nil {
		return
	}
	rslt := 1
	if len(data) > minValue {
		rslt = 0
	}
	tf := transform.NewMap(map[string]string{})
	tf.Set("path", path)
	tf.Set("count", strconv.Itoa(len(data)))
	tf.Set("r", strconv.Itoa(rslt))
	r = append(r, tf.Translate(report))
	return r, nil
}
func (s *collectProxy) dbCollect(ctx *context.Context, param []interface{}, report string) (r []string, err error) {
	if len(param) != 2 {
		err = fmt.Errorf("输入参数有误,db.collector,param只能包含1个参数:%v", param)
		return
	}
	query := fmt.Sprintf("%v", param[0])
	db, err := s.getDB(ctx)
	if err != nil {
		err = fmt.Errorf("输入参数有误,db.collector,db参数配置有误:%v", param)
		return
	}
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
func (s *collectProxy) cpuCollect(ctx *context.Context, param []interface{}, report string) (r []string, err error) {
	maxValue, err := strconv.ParseFloat(fmt.Sprintf("%v", param[0]), 64)
	if err != nil {
		err = fmt.Errorf("参数未配置或类型有误,cpu.collector,param至少包含1个参数maxValue:%v", param)
		return
	}
	cpuInfo := cpu.GetInfo()
	result := 1
	if cpuInfo.UsedPercent < maxValue {
		result = 0
	}
	r = make([]string, 0, 1)
	tf := transform.New()
	tf.Set("idle", fmt.Sprintf("%.2f", cpuInfo.Idle))
	tf.Set("total", fmt.Sprintf("%.2f", cpuInfo.Total))
	tf.Set("percent", fmt.Sprintf("%.2f", cpuInfo.UsedPercent))
	tf.Set("r", strconv.Itoa(result))
	r = append(r, tf.Translate(report))
	return r, nil
}
func (s *collectProxy) memCollect(ctx *context.Context, param []interface{}, report string) (r []string, err error) {
	maxValue, err := strconv.ParseFloat(fmt.Sprintf("%v", param[0]), 64)
	if err != nil {
		err = fmt.Errorf("参数未配置或类型有误,mem.collector,param至少包含1个参数maxValue:%v", param)
		return
	}
	memoryInfo := memory.GetInfo()
	result := 1
	if memoryInfo.UsedPercent < maxValue {
		result = 0
	}
	r = make([]string, 0, 1)
	tf := transform.New()
	tf.Set("idle", fmt.Sprintf("%d", memoryInfo.Idle))
	tf.Set("total", fmt.Sprintf("%d", memoryInfo.Total))
	tf.Set("percent", fmt.Sprintf("%.2f", memoryInfo.UsedPercent))
	tf.Set("r", strconv.Itoa(result))
	r = append(r, tf.Translate(report))
	return r, nil
}
func (s *collectProxy) diskCollect(ctx *context.Context, param []interface{}, report string) (r []string, err error) {
	maxValue, err := strconv.ParseFloat(fmt.Sprintf("%v", param[0]), 64)
	if err != nil {
		err = fmt.Errorf("参数未配置或类型有误,disk.collector,param至少包含1个参数maxValue:%v", param)
		return
	}
	diskInfo := disk.GetInfo()
	result := 1
	if diskInfo.UsedPercent < maxValue {
		result = 0
	}
	r = make([]string, 0, 1)
	tf := transform.New()
	tf.Set("idle", fmt.Sprintf("%d", diskInfo.Idle))
	tf.Set("total", fmt.Sprintf("%d", diskInfo.Total))
	tf.Set("percent", fmt.Sprintf("%.2f", diskInfo.UsedPercent))
	tf.Set("r", strconv.Itoa(result))
	r = append(r, tf.Translate(report))
	return r, nil
}
