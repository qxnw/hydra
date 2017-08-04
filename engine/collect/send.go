package collect

import (
	"encoding/json"
	"fmt"
	"strings"
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

func (s *collectProxy) notifySend(ctx *context.Context) (r string, st int, err error) {
	settingData, err := ctx.GetVarParamByArgsName("setting", "setting")
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

	influxdb, err := s.getInfluxClient(ctx, "influxdb")
	if err != nil {
		return
	}
	tf := transform.New()
	tf.Set("time", ctx.GetArgValue("time", "1m"))
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
			groups := strings.Split(group, ",")
			for _, u := range settingObj.Users {
				if s.checkNeedSend(groups, u.Group) {
					st, err = s.sendWXNotify(alarm, u.OpenID, settingObj.WxNotify, title, content, happendTime, remark)
					if err != nil {
						return "", st, err
					}
				}
			}
		}
	}
	return "SUCCESS", 200, nil
}
func (s *collectProxy) checkNeedSend(dataGroups []string, ugroup string) bool {
	if ugroup == "" {
		return true
	}
	ugroups := strings.Split(ugroup, ",")
	for _, v := range dataGroups {
		for _, k := range ugroups {
			if v == k {
				return true
			}
		}
	}
	return false
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
		title = fmt.Sprintf("%s恢复正常", input["title"])
		content = fmt.Sprintf("%s已恢复正常", input["msg"])
	} else {
		title = fmt.Sprintf("%s出现异常", input["title"])
		switch input["type"] {
		case "cpu", "mem", "disk":
			content = fmt.Sprintf("%s负载过高，请及时处理", input["msg"])
		case "http", "tcp":
			content = fmt.Sprintf("%s出现异常,可能服务器已宕机，请及时处理", input["msg"])
		case "registry":
			content = fmt.Sprintf("%s出现异常,可能部分服务器已宕机，请及时处理", input["msg"])
		default:
			content = fmt.Sprintf("%s，请及时处理", input["msg"])
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
