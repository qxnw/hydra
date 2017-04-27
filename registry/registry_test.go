package registry

import "testing"
import "github.com/qxnw/lib4go/ut"

func TestGetClusterName1(t *testing.T) {
	address := "zk://192.157.0.1,192.168.0.2"
	n, a, err := ResolveAddress(address)
	ut.Expect(t, err, nil)
	ut.Expect(t, n, "zk")
	ut.Expect(t, len(a), 2)
}
func TestGetClusterName2(t *testing.T) {
	_, _, err := ResolveAddress("102")
	ut.Refute(t, err, nil)

	_, _, err = ResolveAddress("://102")
	ut.Refute(t, err, nil)

	_, _, err = ResolveAddress("://")
	ut.Refute(t, err, nil)
}
