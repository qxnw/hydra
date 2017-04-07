package rpc

import (
	"errors"
	"testing"
	"time"

	"strings"

	"github.com/qxnw/hydra/client/rpc"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/hydra/server"
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

func TestRPCServer1(t *testing.T) {
	handler := &contextHandler{version: 101}
	conf, err := registry.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := server.NewServer("rpc.server", handler, nil, conf, nil)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)

	client := rpc.NewRPCClient(server.GetAddress())
	s, r, err := client.Request("/order/request", map[string]string{}, false)
	ut.Expect(t, err, nil)
	ut.Expect(t, r, "success")
	ut.Expect(t, s, 200)
	client.Close()
}
func TestRPCServer2(t *testing.T) {
	handler := &contextHandler{version: 101}
	conf, err := registry.NewJSONConfWithJson(confstr2, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraRPCServer(handler, nil, conf, nil)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)

	time.Sleep(time.Second)
	ut.Expect(t, server.server.port, 2032)
	ut.Expect(t, server.server.serverName, "merchant.api")
}

func TestRPCServer3(t *testing.T) {
	handler := &contextHandler{version: 101}
	conf, err := registry.NewJSONConfWithJson(confstr3, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraRPCServer(handler, nil, conf, nil)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)

	conf, err = registry.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.ExpectSkip(t, err, nil)

	err = server.Notify(conf)
	ut.Refute(t, err, nil)
	ut.Expect(t, server.server.port, 2033)
	ut.Expect(t, server.server.serverName, "merchant.web")

	//wait
	conf, err = registry.NewJSONConfWithJson(confstr3, 101, handler.GetPath)
	ut.ExpectSkip(t, err, nil)
	server.Notify(conf)
	ut.Expect(t, server.server.port, 2033)
	ut.Expect(t, server.server.serverName, "merchant.web")

}

func TestRPCServer4(t *testing.T) {
	handler := &contextHandler{version: 101}
	conf, err := registry.NewJSONConfWithJson(confstr4, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraRPCServer(handler, nil, conf, nil)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)

	client := rpc.NewRPCClient(server.GetAddress())
	s, r, err := client.Request("/order/request", map[string]string{}, false)
	ut.Expect(t, err, nil)
	ut.Expect(t, r, "success")
	ut.Expect(t, s, 200)
	client.Close()

	conf, err = registry.NewJSONConfWithJson(confstr5, 101, handler.GetPath)
	ut.ExpectSkip(t, err, nil)
	err = server.Notify(conf)
	ut.Expect(t, err, nil)

	client = rpc.NewRPCClient(server.GetAddress())
	s, r, err = client.Request("/order/request/1234", map[string]string{}, false)
	ut.Expect(t, err, nil)
	ut.Expect(t, r, "success")
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
    "type": "rpc.server",
    "name": "merchant.rpc",
	"address":":2031",
    "status": "starting",
    "package": "1.0.0.1",
    "QPS": 1000,
    "metric": "#@domain/var/db/influxdb",
    "limiter": "#@path/limiter1",
    "router": "#@path/router1"
}`

var confstr2 = `{
    "type": "api.server01",
    "name": "merchant.api",
	"host":"www.upay6.com,www.upay7.com",
	"address":":2032",
    "status": "starting",
    "package": "1.0.0.1",
    "QPS": 1000,
    "metric": "#@domain/var/db/influxdb",
    "limiter": "#@path/limiter2",
    "router": "#@path/router2"
}`
var confstr3 = `{
    "type": "api.server01",
    "name": "merchant.web",
	"address":":2033",
    "status": "starting",
    "package": "1.0.0.1",
    "QPS": 1000,
    "metric": "#@domain/var/db/influxdb",
    "limiter": "#@path/limiter2",
    "router": "#@path/router2"
}`
var confstr4 = `{
    "type": "api.server01",
    "name": "merchant.api",
	"address":":2034",
    "status": "starting",
    "package": "1.0.0.1",
    "QPS": 1000,
    "metric": "#@domain/var/db/influxdb",
    "limiter": "#@path/limiter1",
    "router": "#@path/router1"
}`
var confstr5 = `{
    "type": "api.server01",
    "name": "merchant.api",
	"host":"",
	"address":":2035",
    "status": "starting",
    "package": "1.0.0.1",
    "QPS": 1000,
    "metric": "#@domain/var/db/influxdb",
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
            "method": "request,query",
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
            "method": "request,query",
            "service": "../@type/@name/script/@module_@action:@method",
            "params": "db=@domain/var/db/influxdb"
        }
    ]
}`
