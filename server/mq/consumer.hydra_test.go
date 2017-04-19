package mq

import (
	"errors"
	"fmt"
	"testing"

	"strings"

	"time"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/mq"
	"github.com/qxnw/lib4go/ut"
)

//192.168.0.155:61613
type contextHandler struct {
	service chan string
	version int32
}

func (h contextHandler) Handle(name string, mode string, s string, c *context.Context) (r *context.Response, err error) {
	fmt.Println("Handle:", mode, "-", s)
	select {
	case h.service <- s:
	default:
	}
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
	handler := &contextHandler{version: 101, service: make(chan string, 1)}
	conf, err := conf.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	_, err = server.NewServer("mq", handler, nil, conf)
	ut.ExpectSkip(t, err, nil)
}
func TestMQServer2(t *testing.T) {
	handler := &contextHandler{version: 101, service: make(chan string, 1)}
	conf, err := conf.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	_, err = newHydraMQConsumer(handler, nil, conf)
	ut.ExpectSkip(t, err, nil)
	p, err := mq.NewStompProducer(mq.ProducerConfig{Address: "192.168.0.165:61613"})
	ut.ExpectSkip(t, err, nil)
	go func() {
		err = p.Connect()
		ut.ExpectSkip(t, err, nil)
	}()
	time.Sleep(time.Second * 2)
	err = p.Send("hydra", "hello", 0)
	ut.ExpectSkip(t, err, nil)
	sv := ""
	select {
	case <-time.After(time.Second * 5):
		break
	case sv = <-handler.service:
	}
	ut.Expect(t, sv, "/order/request/request")
}

var confstr1 = `{
    "type": "mq.consumer",
    "name": "mq.consumer.order",
    "status": "starting",
    "package": "1.0.0.1",  
	"address":"192.168.0.165:61613",
	"version":"1.0",
	"metric": "#@domain/var/db/influxdb",
    "queue": "#@path/queue"   
}`

var metricStr1 = `{
    "host":"192.168.0.92",
    "dataBase":"hydra",
    "userName":"hydra",
    "password":"123456",
    "timeSpan":10
}`
var queue1 = `{
    "queues": [
        {
            "name": "hydra",     
            "service": "/order/request/{@action}",
			"action": "request",
            "args": "db=@domain/var/db/influxdb"
        }
    ]
}`
