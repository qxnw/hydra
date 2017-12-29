package mq

import (
	"errors"
	"testing"

	"strings"

	"time"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/mq"
	"github.com/qxnw/lib4go/ut"
)

type contextData struct {
	service string
	args    interface{}
}

//192.168.0.155:61613
type contextHandler struct {
	notify  chan *contextData
	version int32
}

func (h contextHandler) Handle(name string, mode string, s string, c *context.Context) (r context.Response, err error) {
	h.notify <- &contextData{service: s, args: c.Input.Args}
	return &context.Response{Content: "success"}, nil
}
func (h contextHandler) GetPath(p string) (conf.Conf, error) {

	if strings.HasSuffix(p, "influxdb") {
		return conf.NewJSONConfWithJson(metricStr1, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "queue") {
		return conf.NewJSONConfWithJson(queue1, h.version, h.GetPath)
	}
	return nil, errors.New("not find")
}

func TestMQServer1(t *testing.T) {
	handler := &contextHandler{version: 101, notify: make(chan *contextData, 1)}
	conf, err := conf.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	_, err = server.NewServer("mq", handler, nil, conf)
	ut.ExpectSkip(t, err, nil)
}
func TestMQServer2(t *testing.T) {
	handler := &contextHandler{version: 101, notify: make(chan *contextData, 1)}
	conf, err := conf.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	h, err := newHydraMQConsumer(handler, nil, conf)
	ut.ExpectSkip(t, err, nil)
	err = h.Start()
	ut.ExpectSkip(t, err, nil)

	p, err := mq.NewStompProducer(mq.ProducerConfig{Address: "192.168.0.142:61613"})
	ut.ExpectSkip(t, err, nil)
	go func() {
		err = p.Connect()
		ut.ExpectSkip(t, err, nil)
	}()
	err = p.Send("hydra", "hello", 0)
	ut.ExpectSkip(t, err, nil)
	select {
	case <-time.After(time.Second * 3):
		t.Error("请求超时")
		break
	case data := <-handler.notify:
		ut.Expect(t, data.service, "/order_query/get")
		ut.RefuteSkip(t, data.args, nil)
		ut.RefuteSkip(t, data.args.(map[string]string), nil)
		ut.Expect(t, data.args.(map[string]string)["db"], "/var/db/influxdb/get")
	}

}

var confstr1 = `{  
    "status": "starting",
    "package": "1.0.0.1",  
	"address":"stomp://192.168.0.142:61613",
	"version":"1.0",
	"metric": "#/@domain/var/db/influxdb",
    "queue": "#@path/queue"   
}`

var metricStr1 = `{
    "host":"192.168.0.185:8086",
    "dataBase":"hydra",
    "userName":"hydra",
    "password":"123456"
}`
var queue1 = `{
    "queues": [
        {
            "name": "hydra", 
			"action": "get",  
            "service": "/order_query/@action",			
            "args": "db=/@domain/var/db/influxdb/@action"
        }
    ]
}`
