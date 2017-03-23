package rpc

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/qxnw/lib4go/rpc/server/pb"
)

func BenchmarkItems(t *testing.B) {
	srv := NewServer("rpc", "127.0.0.1:8989")
	srv.Request("/:name/:request/:id", func(c *Context) string {
		return c.Param("id")
	})

	for i := 0; i < t.N; i++ {
		url := fmt.Sprintf("/order/request/%d", i)
		request := &pb.RequestContext{Service: url}
		result, err := srv.process.Request(context.Background(), request)
		if err != nil {
			t.Error(err)
		}
		expectb(t, int(result.Status), http.StatusOK)
		expectb(t, result.Result, fmt.Sprintf("%d", i))
	}

}
func expectb(t *testing.B, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func refuteb(t *testing.B, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
