package api

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/ut"
)

type contextHandler struct {
	version  int32
	services chan string
}

func (h contextHandler) Handle(name string, mode string, service string, c *context.Context) (r *context.Response, err error) {
	select {
	case h.services <- service:
	default:
	}
	return &context.Response{Content: "success"}, nil
}
func (h contextHandler) GetPath(p string) (conf.Conf, error) {

	if strings.HasSuffix(p, "influxdb1") {
		return conf.NewJSONConfWithJson(metricStr1, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "router1") {
		return conf.NewJSONConfWithJson(routerStr1, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "influxdb2") {
		return conf.NewJSONConfWithJson(metricStr2, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "router2") {
		return conf.NewJSONConfWithJson(routerStr2, h.version, h.GetPath)
	}
	return nil, errors.New("not find")
}

func TestServer1(t *testing.T) {
	handler := &contextHandler{version: 101, services: make(chan string, 1)}
	conf, err := conf.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := server.NewServer("api", handler, nil, conf)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)

	client := http.NewHTTPClient()
	url := fmt.Sprintf("%s/order/request", server.GetAddress())
	c, s, err := client.Post(url, "")
	ut.Expect(t, err, nil)
	ut.Expect(t, s, 200)

	url = fmt.Sprintf("%s/order/request/123", server.GetAddress())
	c, s, err = client.Post(url, "")
	ut.Expect(t, err, nil)
	ut.Expect(t, s, 404)
	ut.Expect(t, c, "Not Found")
}
func TestServer2(t *testing.T) {
	handler := &contextHandler{version: 101, services: make(chan string, 1)}
	conf, err := conf.NewJSONConfWithJson(confstr2, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraWebServer(handler, nil, conf)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)

	time.Sleep(time.Second)
	ut.Expect(t, server.server.port, 1032)
	ut.Expect(t, server.server.serverName, "merchant.api")
	ut.ExpectSkip(t, len(server.server.hostNames), 2)
	ut.Expect(t, server.server.hostNames[0], "www.upay6.com")
	ut.Expect(t, server.server.hostNames[1], "www.upay7.com")
}

func TestServer3(t *testing.T) {
	handler := &contextHandler{version: 101, services: make(chan string, 1)}
	cnf, err := conf.NewJSONConfWithJson(confstr3, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraWebServer(handler, nil, cnf)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)

	cnf, err = conf.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.ExpectSkip(t, err, nil)

	err = server.Notify(cnf)
	ut.Refute(t, err, nil)
	ut.Expect(t, server.server.port, 1033)
	ut.Expect(t, server.server.serverName, "merchant.web")
	ut.ExpectSkip(t, len(server.server.hostNames), 0)

	//wait
	cnf, err = conf.NewJSONConfWithJson(confstr3, 101, handler.GetPath)
	ut.ExpectSkip(t, err, nil)
	server.Notify(cnf)
	ut.Expect(t, server.server.port, 1033)
	ut.Expect(t, server.server.serverName, "merchant.web")
	ut.ExpectSkip(t, len(server.server.hostNames), 0)
}

func TestServer4(t *testing.T) {
	handler := &contextHandler{version: 101, services: make(chan string, 1)}
	conf, err := conf.NewJSONConfWithJson(confstr4, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraWebServer(handler, nil, conf)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)

	client := http.NewHTTPClient()
	url := fmt.Sprintf("%s/order/request", server.GetAddress())
	_, s, err := client.Post(url, "")
	ut.Expect(t, err, nil)
	ut.Expect(t, s, 200)

}

func TestServer5(t *testing.T) {
	handler := &contextHandler{version: 100, services: make(chan string, 1)}
	cnf, err := conf.NewJSONConfWithJson(confstr4, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraWebServer(handler, nil, cnf)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)

	h2 := &contextHandler{version: 102, services: make(chan string, 1)}
	cnf, err = conf.NewJSONConfWithJson(confstr5, 101, h2.GetPath)

	ut.ExpectSkip(t, err, nil)
	err = server.Notify(cnf)
	ut.Expect(t, err, nil)
	url := fmt.Sprintf("%s/order/request/123", server.GetAddress())
	fmt.Println(url)
	client := http.NewHTTPClient()
	_, s, err := client.Post(url, "")
	ut.Expect(t, err, nil)
	ut.Expect(t, s, 200)

	sv := ""
	select {
	case sv = <-handler.services:
	default:

	}
	ut.Expect(t, sv, "order_request:POST")
}

type contextNotifyHandler struct {
	version int32
	notify  chan int
	name    string
	mode    string
	service string
}

func (h *contextNotifyHandler) Handle(name string, mode string, service string, c *context.Context) (r *context.Response, err error) {
	h.name = name
	h.mode = mode
	h.service = service
	h.notify <- 1
	return &context.Response{Content: "success"}, nil
}
func (h *contextNotifyHandler) GetPath(p string) (conf.Conf, error) {
	if strings.HasSuffix(p, "influxdb1") {
		return conf.NewJSONConfWithJson(metricStr1, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "router1") {
		return conf.NewJSONConfWithJson(routerStr1, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "influxdb2") {
		return conf.NewJSONConfWithJson(metricStr2, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "router2") {
		return conf.NewJSONConfWithJson(routerStr2, h.version, h.GetPath)
	}
	return nil, errors.New("not find")
}
func TestServer6(t *testing.T) {
	handler := &contextNotifyHandler{version: 100, notify: make(chan int, 1)}
	cnf, err := conf.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraWebServer(handler, nil, cnf)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)

	url := fmt.Sprintf("%s/order/request", server.GetAddress())

	client := http.NewHTTPClient()
	_, s, err := client.Post(url, "")
	ut.Expect(t, err, nil)
	ut.Expect(t, s, 200)

	select {
	case <-time.After(time.Second * 4):
		t.Error("获取参数超时")
		t.FailNow()
	case <-handler.notify:
		ut.Expect(t, handler.name, "/:m/:a")
		ut.Expect(t, handler.mode, "*")
		ut.Expect(t, handler.service, "order_request:POST")
	}

}

var confstr1 = `{
    "type": "api.server01",
    "name": "merchant.api",
	"address":":1031",
    "status": "starting",
    "package": "1.0.0.1",
    "QPS": 1000,
    "metric": "#@domain/var/db/influxdb1",
    "limiter": "#@path/limiter1",
    "router": "#@path/router1"
}`

var confstr2 = `{
    "type": "api.server01",
    "name": "merchant.api",
	"host":"www.upay6.com,www.upay7.com",
	"address":":1032",
    "status": "starting",
    "package": "1.0.0.1",
    "QPS": 1000,
    "metric": "#@domain/var/db/influxdb1",
    "limiter": "#@path/limiter2",
    "router": "#@path/router2"
}`
var confstr3 = `{
    "type": "api.server01",
    "name": "merchant.web",
	"address":":1033",
    "status": "starting",
    "package": "1.0.0.1",
    "QPS": 1000,
    "metric": "#@domain/var/db/influxdb1",
    "limiter": "#@path/limiter2",
    "router": "#@path/router2"
}`
var confstr4 = `{
    "type": "api.server01",
    "name": "merchant.api",
	"address":":1034",
    "status": "starting",
    "package": "1.0.0.1",
    "QPS": 1000,
    "metric": "#@domain/var/db/influxdb1",
    "limiter": "#@path/limiter1",
    "router": "#@path/router1"
}`
var confstr5 = `{
    "type": "api.server01",
    "name": "merchant.api",
	"host":"",
	"address":":1035",
    "status": "starting",
    "package": "1.0.0.1",
    "QPS": 1000,
    "metric": "#@domain/var/db/influxdb1",
    "limiter": "#@path/limiter2",
    "router": "#@path/router2"
}`
var metricStr1 = `{
    "host":"192.168.0.185:8086",
    "dataBase":"hydra",
    "userName":"hydra",
    "password":"123456"
}`
var routerStr1 = `{
    "routers": [
        {
            "name": "/:m/:a",
            "action": "post,get",
            "service": "{@m}_{@a}:{@method}",
            "args": "db=@domain/var/db/influxdb"
        }
    ]
}`

var metricStr2 = `{
    "host":"192.168.0.185:8086",
    "dataBase":"hydra",
    "userName":"hydra",
    "password":"123456",
    "timeSpan":10
}`
var routerStr2 = `{
    "routers": [
        {
            "name": "/:module/:ac/:id",
            "action": "post,get",
            "service": "{@module}_{@ac}:{@method}",
            "args": "db=@domain/var/db/influxdb"
        }
    ]
}`
