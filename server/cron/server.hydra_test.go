package cron

import (
	"errors"
	"testing"

	"strings"

	"time"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/ut"
)

type contextHandler struct {
	version  int32
	services chan string
}

func (h contextHandler) Handle(name string, method string, s string, c *context.Context) (r *context.Response, err error) {
	select {
	case h.services <- s:
	default:
	}
	return &context.Response{Content: "success"}, nil
}
func (h contextHandler) GetPath(p string) (conf.Conf, error) {
	if strings.HasSuffix(p, "influxdb") {
		return conf.NewJSONConfWithJson(metricStr1, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "task") {
		return conf.NewJSONConfWithJson(taskStr1, h.version, h.GetPath)
	}
	return nil, errors.New("not find")
}

func TestCronServer1(t *testing.T) {
	handler := &contextHandler{version: 101, services: make(chan string, 1)}
	conf, err := conf.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	_, err = server.NewServer("cron", handler, nil, conf)
	ut.ExpectSkip(t, err, nil)
}
func TestCronServer2(t *testing.T) {
	handler := &contextHandler{version: 101, services: make(chan string, 1)}
	conf, err := conf.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraCronServer(handler, nil, conf)
	ut.ExpectSkip(t, err, nil)
	server.server.execute()
	server.server.execute()
	time.Sleep(time.Millisecond)

	sv := ""
	select {
	case sv = <-handler.services:
	default:
	}
	ut.Expect(t, sv, "/order/request:request")

}

/*
func TestServer41(t *testing.T) {
	handler := &contextHandler{version: 101}
	conf, err := registry.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraWebServer(handler, conf)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)
	time.Sleep(time.Hour)

}
*/
var confstr1 = `{   
    "status": "start",
    "package": "1.0.0.1",  
	"metric": "#@domain/var/db/influxdb",
    "task": "#@path/task"   
}`

var metricStr1 = `{
    "host":"192.168.0.92",
    "dataBase":"hydra",
    "userName":"hydra",
    "password":"123456",
    "timeSpan":10
}`
var taskStr1 = `{
    "tasks": [
        {
            "name": "cron",           
			"cron":"@every 2s",
            "service": "/order/request:@action",
			"action": "request",
            "params": "db=@domain/var/db/influxdb"
        }
    ]
}`
