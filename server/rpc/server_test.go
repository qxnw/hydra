package rpc

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/qxnw/hydra/server/rpc/pb"
)

func TestServer1(t *testing.T) {
	srv := NewRPCServer("rpc")
	srv.Request("/:name/:request", func(c *Context) string {
		return "OK"
	})
	request := &pb.RequestContext{Service: "/order/request"}
	result, err := srv.process.Request(context.Background(), request)
	if err != nil {
		t.Error(err)
	}
	expect(t, int(result.Status), http.StatusOK)
	expect(t, string(result.Result), "OK")
}
func TestServer2(t *testing.T) {
	srv := NewRPCServer("rpc")
	srv.Request("/:name/:request", func(c *Context) string {
		return "OK"
	})
	request := &pb.RequestContext{Service: "/order/request/colin"}
	result, err := srv.process.Request(context.Background(), request)
	if err != nil {
		t.Error(err)
	}
	expect(t, int(result.Status), http.StatusNotFound)
	expect(t, string(result.Result), "Not Found")
}

func TestServer3(t *testing.T) {
	srv := NewRPCServer("rpc")
	srv.Insert("/:name/:request", func(c *Context) string {
		return "OK"
	})
	request := &pb.RequestContext{Service: "/order/request"}
	result, err := srv.process.Insert(context.Background(), request)
	if err != nil {
		t.Error(err)
	}
	expect(t, int(result.Status), http.StatusOK)
}

func TestServer4(t *testing.T) {
	srv := NewRPCServer("rpc")
	srv.Insert("/:name/:request", func(c *Context) {
		c.NotFound()
	})
	request := &pb.RequestContext{Service: "/order/request"}
	result, err := srv.process.Insert(context.Background(), request)
	if err != nil {
		t.Error(err)
	}
	expect(t, int(result.Status), http.StatusNotFound)
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
