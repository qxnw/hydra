package cache

import (
	"testing"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/ut"
)

func TestGet1(t *testing.T) {
	proxy := &cacheProxy{}
	ctx := context.NewTContext(nil)
	response, err := proxy.get("", "", "", ctx)
	ut.Refute(t, err, nil)
	ut.Expect(t, response.Status, 406)
}
func TestGet2(t *testing.T) {
	proxy := &cacheProxy{}
	ctx := context.NewTContext(nil)
	ctx.Input.Input.Set("key", "12333")
	response, err := proxy.get("", "", "", ctx)
	ut.Refute(t, err, nil)
	ut.Expect(t, response.Status, 410)
}

func TestGet3(t *testing.T) {
	proxy := &cacheProxy{}
	ctx := context.NewTContext(func(f, n string) (string, error) {
		return `{"server":"tcache://127"}`, nil
	})
	ctx.Input.Input.Set("key", "12333")
	ctx.Input.Args["cache"] = "cache"
	response, err := proxy.get("", "", "", ctx)
	ut.Expect(t, err, nil)
	ut.Expect(t, response.Content, cache.Value)
	ut.Expect(t, response.Status, 200)

}

func TestGet4(t *testing.T) {
	proxy := &cacheProxy{}
	ctx := context.NewTContext(func(f, n string) (string, error) {
		return `{"server":"tcache://127"}`, nil
	})
	ctx.Input.Input.Set("key", "12333")
	ctx.Input.Args["cache"] = "cache"
	response, err := proxy.get("", "", "", ctx)
	ut.Expect(t, err, nil)
	ut.Expect(t, response.Content, cache.Value)
	ut.Expect(t, response.Status, 200)

}

var cache = &context.TCache{Value: "1234"}

func init() {
	context.RegisterTCache("tcache", cache)
}
