package alarm

import (
	"fmt"
	"net/url"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/transform"
)

func (s *alarmProxy) getInflux2WxParams(ctx *context.Context) (sql string, wxSetting conf.Conf, wxUsers []conf.Conf, err error) {
	dbSeting, notifySetting, err := s.getQueryParams(ctx)
	if err != nil {
		return
	}
	influxDb, err := dbSeting.GetSection("influxdb")
	if err != nil {
		err = fmt.Errorf("setting[%v]配置错误，未配置db.influxDb节点（err:%v)", dbSeting, err)
		return
	}
	sql = influxDb.String("q")
	if sql == "" {
		err = fmt.Errorf("setting[%v]配置错误，未配置db.influxDb.sql节点（err:%v)", dbSeting, err)
		return
	}

	wx, err := notifySetting.GetSection("wx")
	if err != nil {
		err = fmt.Errorf("setting[%v]配置错误，未配置notify.wx节点（err:%v)", dbSeting, err)
		return
	}
	wxSetting, err = wx.GetSection("settings")
	if err != nil {
		err = fmt.Errorf("setting[%v]配置错误，未配置notify.wx.settings节点（err:%v)", dbSeting, err)
		return
	}
	wxUsers, err = wx.GetSections("users")
	if err != nil {
		err = fmt.Errorf("setting[%v]配置错误，未配置notify.wx.users节点（err:%v)", dbSeting, err)
		return
	}
	return

}
func (s *alarmProxy) influx2wx(ctx *context.Context) (r string, err error) {
	sql, wxSetting, wxUsers, err := s.getInflux2WxParams(ctx)
	if err != nil {
		return
	}
	datas, err := s.influxQuery(ctx, sql)
	if err != nil {
		return
	}
	client := http.NewHTTPClient()
	for _, data := range datas { //循环数据，进行发送
		fm := transform.NewMap(data)
		wxSetting.Each(func(key string) {
			value := wxSetting.String(key)
			if value != "" {
				fm.Set(key, value)
			}
		})
		for _, user := range wxUsers { //每个用户发送消息
			user.Each(func(k string) {
				fm.Set(k, user.String(k))
			})
			host := fm.Translate(wxSetting.String("host"))
			u, err := url.Parse(host)
			if err != nil {
				return "", fmt.Errorf("unable to parse wx url %s. err=%v", host, err)
			}
			values := u.Query()
			data, err := wxSetting.GetSection("data")
			if err != nil {
				return "", err
			}
			c, err := wxSetting.GetSectionString("content")
			if err != nil {
				return "", err
			}
			data.Set("content", fm.Translate(c))
			data.Each(func(key string) {
				values.Set(key, fm.Translate(data.String(key)))
			})
			u.RawQuery = values.Encode()
			content, status, err := client.Get(u.String())
			if err != nil {
				err = fmt.Errorf("请求返回错误:status:%d,%s(host:%s,err:%v)", status, content, host, err)
			}
			if status != 200 {
				err = fmt.Errorf("请求返回错误:status:%d,%s(host:%s)", status, content, host)
				return "", err
			}
		}

	}
	return "SUCCESS", nil

}
