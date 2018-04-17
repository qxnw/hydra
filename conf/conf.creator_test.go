package conf

import (
	"testing"

	"github.com/qxnw/lib4go/ut"
)

func TestMainConf(t *testing.T) {
	c := NewCreator("/conf")
	c.SetMainConf(`{"address":"@address"}`)
	ut.ExpectSkip(t, len(c.mainParams["/conf"]), 1)
	ut.Expect(t, c.mainParams["/conf"][0], "address")
}
func TestSubConf(t *testing.T) {
	c := NewCreator("/conf")
	c.SetSubConf("app", `{"address":"@address"}`)
	ut.ExpectSkip(t, len(c.subParams["/conf/app"]), 1)
	ut.Expect(t, c.subParams["/conf/app"][0], "address")
}
func TestVarConf(t *testing.T) {
	c := NewCreator("/conf")
	c.SetVarConf("app", "a", `{"address":"@address"}`)
	ut.ExpectSkip(t, len(c.varParams["/conf/app/a"]), 1)
	ut.Expect(t, c.varParams["/conf/app/a"][0], "address")
}
