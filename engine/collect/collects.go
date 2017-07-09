package collect

import (
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/qxnw/hydra/context"

	"fmt"

	"github.com/qxnw/lib4go/influxdb"
	xnet "github.com/qxnw/lib4go/net"
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
	Mode   string        `json:"mode"`
	Params []interface{} `json:"params"`
}

func (s *collectProxy) checkAndSave(mode string, db *influxdb.InfluxClient, tf *transform.Transform, t int, span int) (result string, err error) {
	result = "NONEED"
	query := tf.Translate(s.queryMap[mode])
	value, err := db.QueryMaps(query)
	if err != nil {
		return
	}
	if t == 0 {
		//上次无消息，则不上报
		if len(value) == 0 || len(value[0]) == 0 {
			return
		}
		//上次消息是成功不上报
		if len(value) > 0 && len(value[0]) > 0 && types.GetString(value[0][0]["value"]) == "0" {
			return
		}
		//其它情况，上次消息是失败则上报
	} else {
		//上次消息是失败，但记录时间小于5分钟，则不上报
		if len(value) > 0 && len(value[0]) > 0 && types.GetString(value[0][0]["value"]) == "1" {
			//fmt.Println("time:", value[0][0]["time"], value)
			lastTime, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", fmt.Sprintf("%v", value[0][0]["time"]))
			if err != nil {
				return result, err
			}
			if time.Now().Sub(lastTime).Minutes() < float64(span) {
				return result, nil
			}
		}
	}
	err = db.SendLineProto(tf.TranslateAll(s.reportMap[mode], true))
	if err != nil {
		return "", err
	}
	return "SUCCESS", nil
}
func (s *collectProxy) httpCollect(ctx *context.Context, param []interface{}, db *influxdb.InfluxClient) (rlt string, err error) {
	if len(param) != 1 {
		err = fmt.Errorf("输入参数有误,http.collector,param只能包含1个参数url:%v", param)
		return
	}
	uri := fmt.Sprintf("%v", param[0])
	u, err := url.Parse(uri)
	if err != nil {
		err = fmt.Errorf("请求的URL配置有误:%v", uri)
		return
	}
	client := http.NewHTTPClient()
	_, t, err := client.Get(uri)
	result := types.DecodeInt(t, 200, 0, 1)
	tf := transform.New()
	tf.Set("host", u.Host)
	tf.Set("value", strconv.Itoa(result))
	tf.Set("level", types.GetMapValue("level", ctx.GetArgs(), "1"))
	tf.Set("group", types.GetMapValue("group", ctx.GetArgs(), "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("msg", tf.Translate(types.GetMapValue("msg", ctx.GetArgs(), "-")))
	timeSpan := types.ToInt(ctx.GetArgs()["span"], 5)
	return s.checkAndSave("http", db, tf, result, timeSpan)
}
func (s *collectProxy) tcpCollect(ctx *context.Context, param []interface{}, db *influxdb.InfluxClient) (rlt string, err error) {
	if len(param) != 1 {
		err = fmt.Errorf("输入参数有误,tcp.collector,param未配置tcp.hostname:%v", param)
		return
	}
	host := fmt.Sprintf("%v", param[0])
	conn, err := net.DialTimeout("tcp", host, time.Second)
	if err == nil {
		conn.Close()
	}
	result := types.DecodeInt(err, nil, 0, 1)
	tf := transform.NewMap(map[string]string{})
	tf.Set("host", host)
	tf.Set("value", strconv.Itoa(result))
	tf.Set("level", types.GetMapValue("level", ctx.GetArgs(), "1"))
	tf.Set("group", types.GetMapValue("group", ctx.GetArgs(), "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("msg", tf.Translate(types.GetMapValue("msg", ctx.GetArgs(), "-")))
	timeSpan := types.ToInt(ctx.GetArgs()["span"], 5)
	return s.checkAndSave("tcp", db, tf, result, timeSpan)
}
func (s *collectProxy) registryCollect(ctx *context.Context, param []interface{}, db *influxdb.InfluxClient) (rlt string, err error) {
	if len(param) < 1 {
		err = fmt.Errorf("输入参数有误,registry.collector,param未配置服务路径:%v", param)
		return
	}
	path := fmt.Sprintf("%v", param[0])
	minValue := types.ToInt(param[1], 1)
	data, _, err := s.registry.GetChildren(path)
	if err != nil {
		return
	}
	result := 0
	if len(data) < minValue {
		result = 1
	}
	tf := transform.NewMap(map[string]string{})
	tf.Set("host", path)
	tf.Set("value", strconv.Itoa(result))
	tf.Set("count", strconv.Itoa(len(data)))
	tf.Set("level", types.GetMapValue("level", ctx.GetArgs(), "1"))
	tf.Set("group", types.GetMapValue("group", ctx.GetArgs(), "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("msg", tf.Translate(types.GetMapValue("msg", ctx.GetArgs(), "-")))
	timeSpan := types.ToInt(ctx.GetArgs()["span"], 5)
	return s.checkAndSave("registry", db, tf, result, timeSpan)
}
func (s *collectProxy) dbCollect(ctx *context.Context, db *influxdb.InfluxClient) (rlt string, err error) {
	sql, ok := ctx.GetArgs()["sql"]
	if !ok {
		err = fmt.Errorf("args中未配置sql字段")
		return
	}
	sql, err = s.getVarParam(ctx, "sql", sql)
	if err != nil || sql == "" {
		err = fmt.Errorf("var.sql参数未配置")
		return
	}
	title, ok := ctx.GetArgs()["title"]
	if !ok {
		err = fmt.Errorf("args未配置title字段")
		return
	}
	msg, ok := ctx.GetArgs()["msg"]
	if !ok {
		err = fmt.Errorf("args未配置msg字段")
		return
	}
	smax, ok1 := ctx.GetArgs()["max"]
	smin, ok2 := ctx.GetArgs()["min"]
	if !ok1 && !ok2 {
		err = fmt.Errorf("args未配置max或min")
		return
	}
	max := 0
	min := 0
	if ok1 {
		max, err = strconv.Atoi(smax)
		if err != nil {
			err = fmt.Errorf("args未配置max参数必须是数字:%v", err)
			return
		}
	}
	if ok2 {
		min, err = strconv.Atoi(smin)
		if err != nil {
			err = fmt.Errorf("args未配置min参数必须是数字:%v", err)
			return
		}
	}
	sdb, err := s.getDB(ctx)
	if err != nil {
		err = fmt.Errorf("args数据库db配置有错误:%v", err)
		return
	}
	data, _, _, err := sdb.Scalar(sql, map[string]interface{}{})
	if err != nil {
		err = fmt.Errorf("数据查询出错:sql:%v,err:%v", sql, err)
		return
	}
	if data == nil {
		rlt = "NONEED"
		return
	}
	value, err := strconv.Atoi(fmt.Sprintf("%v", data))
	if err != nil {
		err = fmt.Errorf("sql:%s返回结果不是有效的数字", sql)
		return
	}
	result := 1
	if ((min > 0 && value >= min) || min == 0) && ((max > 0 && value < max) || max == 0) {
		result = 0
	}

	tf := transform.NewMap(map[string]string{})
	tf.Set("host", title)
	tf.Set("value", strconv.Itoa(result))
	tf.Set("count", strconv.Itoa(value))
	tf.Set("level", types.GetMapValue("level", ctx.GetArgs(), "1"))
	tf.Set("group", types.GetMapValue("group", ctx.GetArgs(), "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	timeSpan := types.ToInt(ctx.GetArgs()["span"], 5)
	tf.Set("msg", tf.Translate(msg))
	return s.checkAndSave("db", db, tf, result, timeSpan)
}
func (s *collectProxy) cpuCollect(ctx *context.Context, param []interface{}, db *influxdb.InfluxClient) (rlt string, err error) {
	maxValue, err := strconv.ParseFloat(fmt.Sprintf("%v", param[0]), 64)
	if err != nil {
		err = fmt.Errorf("参数未配置或类型有误,cpu.collector,param至少包含1个参数maxValue:%v", param)
		return
	}
	cpuInfo := cpu.GetInfo()
	result := 0
	if cpuInfo.UsedPercent >= maxValue {
		result = 1
	}
	tf := transform.New()
	tf.Set("host", xnet.LocalIP)
	tf.Set("value", strconv.Itoa(result))
	tf.Set("level", types.GetMapValue("level", ctx.GetArgs(), "1"))
	tf.Set("group", types.GetMapValue("group", ctx.GetArgs(), "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("msg", tf.Translate(types.GetMapValue("msg", ctx.GetArgs(), "-")))
	timeSpan := types.ToInt(ctx.GetArgs()["span"], 5)
	return s.checkAndSave("cpu", db, tf, result, timeSpan)
}
func (s *collectProxy) memCollect(ctx *context.Context, param []interface{}, db *influxdb.InfluxClient) (rlt string, err error) {
	maxValue, err := strconv.ParseFloat(fmt.Sprintf("%v", param[0]), 64)
	if err != nil {
		err = fmt.Errorf("参数未配置或类型有误,mem.collector,param至少包含1个参数maxValue:%v", param)
		return
	}
	memoryInfo := memory.GetInfo()
	result := 0
	if memoryInfo.UsedPercent >= maxValue {
		result = 1
	}
	tf := transform.New()
	tf.Set("host", xnet.LocalIP)
	tf.Set("value", strconv.Itoa(result))
	tf.Set("level", types.GetMapValue("level", ctx.GetArgs(), "1"))
	tf.Set("group", types.GetMapValue("group", ctx.GetArgs(), "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("msg", tf.Translate(types.GetMapValue("msg", ctx.GetArgs(), "-")))
	timeSpan := types.ToInt(ctx.GetArgs()["span"], 5)
	return s.checkAndSave("mem", db, tf, result, timeSpan)
}
func (s *collectProxy) diskCollect(ctx *context.Context, param []interface{}, db *influxdb.InfluxClient) (rlt string, err error) {
	maxValue, err := strconv.ParseFloat(fmt.Sprintf("%v", param[0]), 64)
	if err != nil {
		err = fmt.Errorf("参数未配置或类型有误,disk.collector,param至少包含1个参数maxValue:%v", param)
		return
	}
	diskInfo := disk.GetInfo()
	result := 0
	if diskInfo.UsedPercent >= maxValue {
		result = 1
	}
	tf := transform.New()
	tf.Set("host", xnet.LocalIP)
	tf.Set("value", strconv.Itoa(result))
	tf.Set("level", types.GetMapValue("level", ctx.GetArgs(), "1"))
	tf.Set("group", types.GetMapValue("group", ctx.GetArgs(), "D"))
	tf.Set("time", time.Now().Format("20060102150405"))
	tf.Set("msg", tf.Translate(types.GetMapValue("msg", ctx.GetArgs(), "-")))
	timeSpan := types.ToInt(ctx.GetArgs()["span"], 5)
	return s.checkAndSave("disk", db, tf, result, timeSpan)
}