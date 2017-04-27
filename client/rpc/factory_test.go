package rpc

import "testing"
import "github.com/qxnw/lib4go/ut"

func TestFactoryResolvePath(t *testing.T) {
	def_domain := "hydra"
	def_server := "sys.api"
	f := NewRPCInvoker(def_domain, def_server, "")
	svs, domain, server, err := f.resolvePath("order.request")
	ut.Expect(t, err, nil)
	ut.Expect(t, svs, "/order/request")
	ut.Expect(t, domain, def_domain)
	ut.Expect(t, server, def_server)

	svs, domain, server, err = f.resolvePath("/order/request")
	ut.Expect(t, err, nil)
	ut.Expect(t, svs, "/order/request")
	ut.Expect(t, domain, def_domain)
	ut.Expect(t, server, def_server)

	svs, domain, server, err = f.resolvePath("/order/request@")
	ut.Expect(t, err, nil)
	ut.Expect(t, svs, "/order/request")
	ut.Expect(t, domain, def_domain)
	ut.Expect(t, server, def_server)

	svs, domain, server, err = f.resolvePath("@")
	ut.Refute(t, err, nil)

	svs, domain, server, err = f.resolvePath("@merchant_cron")
	ut.Refute(t, err, nil)

	svs, domain, server, err = f.resolvePath("/order/request@merchant")
	ut.Expect(t, err, nil)
	ut.Expect(t, svs, "/order/request")
	ut.Expect(t, domain, def_domain)
	ut.Expect(t, server, "merchant")

	svs, domain, server, err = f.resolvePath("order.request@merchant.")
	ut.Expect(t, err, nil)
	ut.Expect(t, svs, "/order/request")
	ut.Expect(t, domain, def_domain)
	ut.Expect(t, server, "merchant")

	svs, domain, server, err = f.resolvePath("order.request@merchant.sys")
	ut.Expect(t, err, nil)
	ut.Expect(t, svs, "/order/request")
	ut.Expect(t, domain, "sys")
	ut.Expect(t, server, "merchant")

	svs, domain, server, err = f.resolvePath("order.request@.sys")
	ut.Expect(t, err, nil)
	ut.Expect(t, svs, "/order/request")
	ut.Expect(t, domain, "sys")
	ut.Expect(t, server, def_server)

	svs, domain, server, err = f.resolvePath("order/request@merchant.")
	ut.Expect(t, err, nil)
	ut.Expect(t, svs, "/order/request")
	ut.Expect(t, domain, def_domain)
	ut.Expect(t, server, "merchant")

	svs, domain, server, err = f.resolvePath("order/request@merchant.sys")
	ut.Expect(t, err, nil)
	ut.Expect(t, svs, "/order/request")
	ut.Expect(t, domain, "sys")
	ut.Expect(t, server, "merchant")

	svs, domain, server, err = f.resolvePath("order/request@.sys")
	ut.Expect(t, err, nil)
	ut.Expect(t, svs, "/order/request")
	ut.Expect(t, domain, "sys")
	ut.Expect(t, server, def_server)

}
