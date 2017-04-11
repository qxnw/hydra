package web

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"strings"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/net/http"
	"github.com/qxnw/lib4go/ut"
)

type contextHandler struct {
	version int32
}

func (h contextHandler) Handle(name string, method string, s string, p string, c context.Context) (r *context.Response, err error) {
	return &context.Response{Content: "success"}, nil
}
func (h contextHandler) GetPath(p string) (registry.Conf, error) {

	if strings.HasSuffix(p, "influxdb1") {
		return registry.NewJSONConfWithJson(metricStr1, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "router1") {
		return registry.NewJSONConfWithJson(routerStr1, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "influxdb2") {
		return registry.NewJSONConfWithJson(metricStr2, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "router2") {
		return registry.NewJSONConfWithJson(routerStr2, h.version, h.GetPath)
	}
	return nil, errors.New("not find")
}

func TestServer1(t *testing.T) {
	handler := &contextHandler{version: 101}
	conf, err := registry.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := server.NewServer("api.server", handler, nil, conf, nil)
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
	handler := &contextHandler{version: 101}
	conf, err := registry.NewJSONConfWithJson(confstr2, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraWebServer(handler, nil, conf, nil)
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
	handler := &contextHandler{version: 101}
	conf, err := registry.NewJSONConfWithJson(confstr3, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraWebServer(handler, nil, conf, nil)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)

	conf, err = registry.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.ExpectSkip(t, err, nil)

	err = server.Notify(conf)
	ut.Refute(t, err, nil)
	ut.Expect(t, server.server.port, 1033)
	ut.Expect(t, server.server.serverName, "merchant.web")
	ut.ExpectSkip(t, len(server.server.hostNames), 0)

	//wait
	conf, err = registry.NewJSONConfWithJson(confstr3, 101, handler.GetPath)
	ut.ExpectSkip(t, err, nil)
	server.Notify(conf)
	ut.Expect(t, server.server.port, 1033)
	ut.Expect(t, server.server.serverName, "merchant.web")
	ut.ExpectSkip(t, len(server.server.hostNames), 0)
}

func TestServer4(t *testing.T) {
	handler := &contextHandler{version: 101}
	conf, err := registry.NewJSONConfWithJson(confstr4, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraWebServer(handler, nil, conf, nil)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)

	client := http.NewHTTPClient()
	url := fmt.Sprintf("%s/order/request", server.GetAddress())
	_, s, err := client.Post(url, "")
	ut.Expect(t, err, nil)
	ut.Expect(t, s, 200)

	conf, err = registry.NewJSONConfWithJson(confstr5, 101, handler.GetPath)
	ut.ExpectSkip(t, err, nil)
	err = server.Notify(conf)
	ut.Expect(t, err, nil)
	url = fmt.Sprintf("%s/order/request/123", server.GetAddress())

	_, s, err = client.Post(url, "")
	ut.Expect(t, err, nil)
	ut.Expect(t, s, 200)

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
    "host":"192.168.0.92",
    "dataBase":"hydra",
    "userName":"hydra",
    "password":"123456",
    "timeSpan":10
}`
var routerStr1 = `{
    "routers": [
        {
            "name": "/:module/:action",
            "method": "post,get",
            "service": "../@type/@name/script/@module_@action:@method",
            "params": "db=@domain/var/db/influxdb"
        }
    ]
}`

var metricStr2 = `{
    "host":"192.168.0.92",
    "dataBase":"hydra",
    "userName":"hydra",
    "password":"123456",
    "timeSpan":10
}`
var routerStr2 = `{
    "routers": [
        {
            "name": "/:module/:action/:id",
            "method": "post,get",
            "service": "../@type/@name/script/@module_@action:@method",
            "params": "db=@domain/var/db/influxdb"
        }
    ]
}`
