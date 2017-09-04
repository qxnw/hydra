package rpc

import (
	"testing"

	"github.com/qxnw/lib4go/ut"
)

func TestPublish1(t *testing.T) {
	lst := getDepthOneService("/order/:a", "/order/request")
	ut.ExpectSkip(t, len(lst), 1)
	ut.Expect(t, lst[0], "/order/request")
}
func TestPublish2(t *testing.T) {
	lst := getDepthOneService("/order/:a", "/user/order/request")
	ut.ExpectSkip(t, len(lst), 0)
}

func TestPublish3(t *testing.T) {
	lst := getServices("/order/request", "/user/order/request", "/user/order/request")
	ut.ExpectSkip(t, len(lst), 1)
	ut.Expect(t, lst[0], "/order/request")
}

func TestPublish4(t *testing.T) {
	lst := getDepthTwoService("/order/:request", "/order/@request", "/order/request")
	ut.ExpectSkip(t, len(lst), 1)
	ut.Expect(t, lst[0], "/order/request")
}
func TestPublish5(t *testing.T) {
	lst := getDepthTwoService("/order/:request", "/user/order/@request", "/user/order/request")
	ut.ExpectSkip(t, len(lst), 1)
	ut.Expect(t, lst[0], "/order/request")
}
