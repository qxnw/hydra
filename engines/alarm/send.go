package alarm

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/engines/ssm"

	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
)

type notity struct {
	WxGroup  []string `json:"wx"`
	SmsBroup []string `json:"sms"`
}
type setting struct {
	Notify notity      `json:"notify"`
	Users  []*userInfo `json:"users"`
}
type userInfo struct {
	Name   string   `json:"name"`
	Group  []string `json:"group"`
	Mobile string   `json:"mobile"`
	OpenID string   `json:"wx_openId"`
}

//SendAlarmNotify 发送报警通知
func SendAlarmNotify(c component.IContainer) component.StandardServiceFunc {
	return func(name string, mode string, service string, ctx *context.Context) (response *context.StandardResponse, err error) {
		response = context.GetStandardResponse()
		settingData, err := c.GetVarParam("alarm", ctx.Request.Setting.GetString("notify_setting"))
		if err != nil {
			return
		}
		settingObj := &setting{}
		err = json.Unmarshal([]byte(settingData), settingObj)
		if err != nil {
			err = fmt.Errorf("args.setting配置文件有错:err:%v", err)
			return
		}
		influxdb, err := c.GetInflux("influxdb")
		if err != nil {
			return
		}
		tf := transform.New()
		tf.Set("time", ctx.Request.Setting.GetString("time", "1m"))
		data, err := influxdb.QueryMaps(tf.Translate(reportSQL))
		if err != nil {
			err = fmt.Errorf("从influxdb中查询报警数据失败%s:err:%v", tf.Translate(reportSQL), err)
			return
		}
		//ctx.Info("sql:", tf.Translate(s.reportSQL), data)
		if len(data) == 0 || len(data[0]) == 0 {
			response.SetContent(204, "NONEED")
			return response, nil
		}
		for _, rows := range data {
			for _, item := range rows {
				alarm, title, content, happendTime, group, remark, err := getMessage(item)
				if err != nil {
					response.SetStatus(500)
					return response, nil
				}
				for _, u := range settingObj.Users {
					st, err := Notify(c, ctx, group, alarm, settingObj.Notify, u, title, content, happendTime, remark)
					if err != nil && st != 204 {
						response.SetContent(st, err)
						return response, err
					}
				}
			}
		}
		response.Success("success")
		return
	}
}
func Notify(c component.IContainer, ctx *context.Context, group string, alarm bool, notify notity, u *userInfo, title string, content string, time string, remark string) (st int, err error) {

	dataGroup := strings.Split(group, ",")
	if !checkNeedSend(dataGroup, u.Group) {
		return 204, fmt.Errorf("未找到发送组:data:%v,u:%v", dataGroup, u.Group)
	}
	has := checkNeedSend(notify.WxGroup, u.Group)
	if has {
		st, err = sendWXNotify(c, ctx, alarm, u, title, content, time, remark)
		if err == nil {
			ctx.Log.Infof("微信发送给:%s[发送成功]", u.Name)
		} else {
			ctx.Log.Infof("微信发送给:%s[发送失败] err:%v", u.Name, err)
		}
	}

	if checkNeedSend(notify.SmsBroup, u.Group) {
		has = true
		st, err = sendSMSNotify(c, ctx, alarm, u, title, content, time, remark)
		if err == nil {
			ctx.Log.Infof("短信发送给:%s[发送成功]", u.Name)
		} else {
			ctx.Log.Infof("短信发送给:%s[发送失败] err:%v", u.Name, err)
		}
	}
	if !has {
		st = 404
		err = fmt.Errorf("未找到发送组")
	}

	return
}

func checkNeedSend(dgroup []string, ugroups []string) bool {
	if dgroup == nil {
		return false
	}
	for _, v := range dgroup {
		for _, k := range ugroups {
			if v == k {
				return true
			}
		}
	}
	return false
}
func getMessage(input map[string]interface{}) (alarm bool, title string, content string, happendTime string, group string, remark string, err error) {
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
func sendWXNotify(c component.IContainer, ctx *context.Context, alarm bool, u *userInfo, title string, content string, time string, remark string) (status int, err error) {
	tp := types.DecodeString(alarm, true, "alarm", "normal")
	setting, err := c.GetVarParam("ssm", ctx.Request.Setting.GetString("wx_setting"))
	if err != nil {
		err = fmt.Errorf("未找到微信消息推送的相关配置:%v", err)
		return
	}
	_, status, err = ssm.SendWXM(setting, map[string]string{
		"open_id": u.OpenID,
		"title":   title,
		"time":    time,
		"message": content,
		"remark":  remark,
		"type":    tp,
	})
	return
}
func sendSMSNotify(c component.IContainer, ctx *context.Context, alarm bool, u *userInfo, title string, content string, time string, remark string) (status int, err error) {
	setting, err := c.GetVarParam("ssm", ctx.Request.Setting.GetString("sms_setting"))
	if err != nil {
		err = fmt.Errorf("未找到短信发送配置:%v", err)
		return
	}
	status, _, err = ssm.SendSMS(u.Mobile, fmt.Sprintf("%s;%s;%s;%s", title, content, time, remark), setting)
	return
}
