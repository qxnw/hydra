package cache

import (
	"testing"

	"github.com/qxnw/hydra/context"
	"github.com/qxnw/lib4go/ut"
)

func TestGet1(t *testing.T) {
	proxy := &cacheProxy{}
	ctx := context.NewTContext(nil)
	//	ctx.Input.Input.Set("key", "12333")
	key, err := proxy.getInputKey(ctx)
	ut.Expect(t, err, nil)
	ut.Expect(t, key, "")
}

var cache = &context.TCache{Value: "1234"}

func init() {
	context.RegisterTCache("tcache", cache)
}
