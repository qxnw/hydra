package mq

import (
	"errors"
	"testing"

	"strings"

	"time"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/registry"
	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/mq"
	"github.com/qxnw/lib4go/ut"
)

//192.168.0.155:61613
type contextHandler struct {
	version int32
}

func (h contextHandler) Handle(name string, method string, s string, p string, c context.Context) (r *context.Response, err error) {
	return &context.Response{Content: "success"}, nil
}
func (h contextHandler) GetPath(p string) (registry.Conf, error) {

	if strings.HasSuffix(p, "influxdb") {
		return registry.NewJSONConfWithJson(metricStr1, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "queue") {
		return registry.NewJSONConfWithJson(queue1, h.version, h.GetPath)
	}
	return nil, errors.New("not find")
}

func TestMQServer1(t *testing.T) {
	handler := &contextHandler{version: 101}
	conf, err := registry.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	_, err = server.NewServer("mq.consumer", handler, nil, conf, nil)
	ut.ExpectSkip(t, err, nil)
}
func TestMQServer2(t *testing.T) {
	handler := &contextHandler{version: 101}
	conf, err := registry.NewJSONConfWithJson(confstr1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	_, err = newHydraMQConsumer(handler, nil, conf, nil)
	ut.ExpectSkip(t, err, nil)
	t.Log("abc")
	p, err := mq.NewStompProducer(mq.ProducerConfig{Address: "192.168.0.155:61613"})
	ut.ExpectSkip(t, err, nil)
	t.Log("abc")
	go func() {
		err = p.Connect()
		ut.ExpectSkip(t, err, nil)
	}()
	time.Sleep(time.Second)
	t.Log("abc")
	err = p.Send("hydra", "hello", 0)
	ut.ExpectSkip(t, err, nil)
}

var confstr1 = `{
    "type": "mq.consumer",
    "name": "mq.consumer.order",
    "status": "starting",
    "package": "1.0.0.1",  
	"address":"192.168.0.155:61613",
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
            "queue": "hydra",     
            "service": "/order/request",
			"method": "request",
            "params": "db=@domain/var/db/influxdb"
        }
    ]
}`
