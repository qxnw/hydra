package alarm

import (
	"context"
	"fmt"

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
	sql := influxDb.String("q")
	if sql == "" {
		err = fmt.Errorf("setting[%v]配置错误，未配置db.influxDb.sql节点（err:%v)", dbSeting, err)
		return
	}

	wx, err := notifySetting.GetSection("wx")
	if err != nil {
		err = fmt.Errorf("setting[%v]配置错误，未配置notify.wx节点（err:%v)", dbSeting, err)
		return
	}
	wxSetting, err := wx.GetSection("settings")
	if err != nil {
		err = fmt.Errorf("setting[%v]配置错误，未配置notify.wx.settings节点（err:%v)", dbSeting, err)
		return
	}
	wxUsers, err := wx.GetSections("users")
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
	wxData, err := wxSetting.GetSection("data")
	if err != nil {
		err = fmt.Errorf("setting[%v]配置错误，未配置notify.wx.settings.data节点（err:%v)", dbSeting, err)
		return
	}
	for _, data := range datas {
		fm := transform.NewMap(data)
		fm.Set("appId", wxSetting.String("appId"))
		fm.Set("templateId", wxSetting.String("templateId"))
		for _, user := range wxUsers {
			fm.Set("name", user.String("name"))
			fm.Set("openId", user.String("openId"))
		}
		url := fm.Translate(wxSetting.String("host"))
		data := fm.Translate(wxSetting.String("data"))
		client := http.NewHTTPClient()
		content, status, err := client.Post(url, data)
		if err != nil {
			err = fmt.Errorf("请求返回错误:status:%d,%s(url:%s,err:%v)", status, content, url, err)
		}
		if status != 200 {
			err = fmt.Errorf("请求返回错误:status:%d,%s(url:%s)", status, content, url)
			return
		}
	}
	return "SUCCESS", nil

}
