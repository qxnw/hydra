package collect

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

type setting struct {
	WxNotify string     `json:"wx"`
	SMS      string     `json:"sms"`
	Users    []userInfo `json:"users"`
}
type userInfo struct {
	Name   string `json:"name"`
	Group  string `json:"group"`
	Mobile string `json:"mobile"`
	OpenID string `json:"wx_openId"`
}

func (s *collectProxy) notify(ctx *context.Context) (r string, st int, err error) {
	settingName, ok := ctx.GetArgs()["setting"]
	if !ok {
		err = fmt.Errorf("未配置setting属性(%v)，err:%v", ctx.GetArgs(), err)
		return
	}
	settingData, err := s.getVarParam(ctx, "setting", settingName)
	if err != nil {
		err = fmt.Errorf("setting.%s未配置:err:%v", ctx.GetArgs()["setting"], err)
		return
	}
	settingObj := &setting{}
	err = json.Unmarshal([]byte(settingData), settingObj)
	if err != nil {
		err = fmt.Errorf("setting.%s配置文件有错:err:%v", ctx.GetArgs()["setting"], err)
		return
	}

	influxdb, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	tf := transform.New()
	tf.Set("time", types.GetMapValue("time", ctx.GetArgs(), "1m"))
	data, err := influxdb.QueryMaps(tf.Translate(s.reportSQL))
	if err != nil {
		err = fmt.Errorf("从influxdb中查询报警数据失败%s:err:%v", tf.Translate(s.reportSQL), err)
		return
	}
	if len(data) == 0 || len(data[0]) == 0 {
		return "NONEED", 204, nil
	}

	for _, rows := range data {
		for _, item := range rows {
			alarm, title, content, happendTime, group, remark, err := s.getMessage(item)
			if err != nil {
				return "", 500, err
			}
			for _, u := range settingObj.Users {
				if group == u.Group || u.Group == "" {
					st, err = s.sendWXNotify(alarm, u.OpenID, settingObj.WxNotify, title, content, happendTime, remark)
					if err != nil {
						return "", st, err
					}
				}
			}
		}
	}
	return
}
func (s *collectProxy) getMessage(input map[string]interface{}) (alarm bool, title string, content string, happendTime string, group string, remark string, err error) {
	title = "服务器发生错误"
	lastTime, err := time.Parse("20060102150405", fmt.Sprintf("%v", input["t"]))
	if err != nil {
		fmt.Println("日期格式转换出错", err)
		return false, "", "", "", "", "", err
	}
	group = types.GetString(input["group"])
	remark = "请及时处理，如有疑问请联系运营或技术"
	alarm = types.GetString(input["value"]) != "0"
	if !alarm {
		switch types.GetString(input["type"]) {
		case "http", "tcp":
			title = fmt.Sprintf("%v服务器恢复正常", input["type"])
			content = fmt.Sprintf("%v服务器%v已恢复访问", input["type"], input["host"])
		case "registry":
			title = "注册中心服务恢复正常"
			content = fmt.Sprintf("注册中心%s的服务器数量已恢复", input["host"])
		case "cpu", "mem", "disk":
			title = fmt.Sprintf("服务器%v负载恢复正常", input["type"])
			content = fmt.Sprintf("%v服务器%s负载已恢复正常", input["host"], input["type"])
		}
	} else {
		switch types.GetString(input["type"]) {
		case "http", "tcp":
			title = fmt.Sprintf("[%v]服务器无法访问", input["type"])
			content = fmt.Sprintf("%v服务器 %v 服务无法请求，可能已宕机，请及时处理", input["type"], input["host"])
		case "registry":
			title = "注册中心服务器宕机"
			content = fmt.Sprintf("%s的服务器数量小于预设阀值，部分服务器已宕机，请及时处理", input["host"])
		case "cpu", "mem", "disk":
			title = fmt.Sprintf("服务器%v负载过高", input["type"])
			content = fmt.Sprintf("%v服务器%s负载过高，请及时处理", input["host"], input["type"])
		}
	}
	happendTime = lastTime.Format("2006/01/02 15:04")
	return

}
func (s *collectProxy) sendWXNotify(alarm bool, openID string, service string, title string, content string, time string, remark string) (status int, err error) {
	tp := "alarm"
	if !alarm {
		tp = "normal"
	}
	status, _, _, err = s.rpc.Request(service, map[string]string{
		"openId":  openID,
		"title":   title,
		"time":    time,
		"content": content,
		"remark":  remark,
		"alarm":   tp,
	}, false)
	return
}
